/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package labeler lets every vault node advertise its own HA role.
//
// Each vault pod runs this loop inside the unsealer sidecar. It polls the
// LOCAL vault node's /v1/sys/health endpoint and patches the pod's own
// kubevault.com/role label: `primary` for the unsealed leader,
// `standby` for unsealed standbys, and removed while the node is sealed,
// uninitialized, or unreachable.
//
// The KubeVault operator adds `kubevault.com/role: primary` to the
// client Service selector for raft-backed servers, so the Service always
// points at the raft leader. Because every node maintains its own label,
// leadership changes propagate without any central actor: the demoted node
// relabels itself standby and the new leader relabels itself primary within
// one poll period, even if the operator is down.
package labeler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/tools/clientcmd"
)

const (
	// RoleLabelKey is the pod label carrying the node's HA role. Must match
	// apis.VaultPodRoleLabelKey in kubevault.dev/apimachinery.
	RoleLabelKey = "kubevault.com/role"

	RolePrimary = "primary"
	RoleStandby = "standby"

	// EnvPodName and EnvPodNamespace are injected by the KubeVault operator
	// via the downward API. Without them the labeler stays disabled.
	EnvPodName      = "POD_NAME"
	EnvPodNamespace = "POD_NAMESPACE"
)

// RoleLabeler patches its own pod's role label from the local vault
// node's health.
type RoleLabeler struct {
	vc           *vaultapi.Client
	kc           kubernetes.Interface
	podName      string
	podNamespace string
	period       time.Duration
}

// NewFromEnv builds a RoleLabeler using the downward-API pod identity.
// Returns (nil, nil) when POD_NAME/POD_NAMESPACE are not set or the period
// is zero, so callers can treat labeling as an optional feature.
func NewFromEnv(vc *vaultapi.Client, period time.Duration) (*RoleLabeler, error) {
	podName, podNamespace := os.Getenv(EnvPodName), os.Getenv(EnvPodNamespace)
	if podName == "" || podNamespace == "" || period <= 0 {
		return nil, nil
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load in-cluster config for role labeler: %w", err)
	}
	clientcmd.Fix(config)

	kc, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client for role labeler: %w", err)
	}

	return &RoleLabeler{
		vc:           vc,
		kc:           kc,
		podName:      podName,
		podNamespace: podNamespace,
		period:       period,
	}, nil
}

// Run polls forever. Meant to be started as a goroutine next to the unseal
// loop; it never returns.
func (l *RoleLabeler) Run(ctx context.Context) {
	klog.Infof("starting role labeler for pod %s/%s (period %s)", l.podNamespace, l.podName, l.period)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(l.period):
		}
		if err := l.sync(ctx); err != nil {
			klog.Errorf("failed to sync role label: %s", err.Error())
		}
	}
}

// sync classifies the local node and converges the pod label.
func (l *RoleLabeler) sync(ctx context.Context) error {
	role := l.currentRole(ctx)

	pod, err := l.kc.CoreV1().Pods(l.podNamespace).Get(ctx, l.podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get own pod: %w", err)
	}

	current, exists := pod.Labels[RoleLabelKey]
	if (role == "" && !exists) || (role != "" && current == role) {
		return nil
	}
	return l.patchLabel(ctx, role)
}

// currentRole maps the local node's health to a label value. Empty string
// means the node must not receive client traffic.
func (l *RoleLabeler) currentRole(ctx context.Context) string {
	hctx, cancel := context.WithTimeout(ctx, l.period)
	defer cancel()

	hr, err := l.vc.Sys().HealthWithContext(hctx)
	if err != nil || hr == nil || !hr.Initialized || hr.Sealed {
		return ""
	}
	if hr.Standby || hr.PerformanceStandby {
		return RoleStandby
	}
	return RolePrimary
}

// patchLabel applies a strategic merge patch setting or removing the label.
func (l *RoleLabeler) patchLabel(ctx context.Context, role string) error {
	var value any
	if role == "" {
		value = nil // null removes the key in a merge patch
	} else {
		value = role
	}
	patch, err := json.Marshal(map[string]any{
		"metadata": map[string]any{
			"labels": map[string]any{RoleLabelKey: value},
		},
	})
	if err != nil {
		return err
	}

	_, err = l.kc.CoreV1().Pods(l.podNamespace).Patch(ctx, l.podName, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil && !kerr.IsNotFound(err) {
		return fmt.Errorf("failed to patch role label: %w", err)
	}
	if err == nil {
		if role == "" {
			klog.Infof("removed %s label from pod %s/%s", RoleLabelKey, l.podNamespace, l.podName)
		} else {
			klog.Infof("set %s=%s on pod %s/%s", RoleLabelKey, role, l.podNamespace, l.podName)
		}
	}
	return nil
}
