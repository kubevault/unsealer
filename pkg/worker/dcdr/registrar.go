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

package dcdr

import (
	"context"
	"fmt"
	"strings"
	"time"

	kubesliceapi "github.com/kubeslice/worker-operator/api/v1beta1"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// registrar labels its own pod and creates the per-pod Service and KubeSlice
// ServiceExport, so the pod's advertised address resolves in every data center.
type Registrar struct {
	Env       *Env
	SliceName string
	KC        client.Client
}

func (r *Registrar) validate() error {
	if r.Env.PodName == "" || r.Env.PodUID == "" || r.Env.PodNamespace == "" {
		return errors.Errorf("env %s, %s and %s must be set in distributed mode", EnvPodName, EnvPodUID, EnvPodNamespace)
	}
	if r.SliceName == "" {
		return errors.New("flag --slice-name must be set in distributed mode")
	}
	if errs := validation.IsDNS1035Label(r.Env.PerPodName()); len(errs) > 0 {
		return errors.Errorf("per-pod name %q is not a valid DNS-1035 label: %s", r.Env.PerPodName(), strings.Join(errs, "; "))
	}
	return nil
}

func (r *Registrar) ownerRef() metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: "v1",
		Kind:       "Pod",
		Name:       r.Env.PodName,
		UID:        types.UID(r.Env.PodUID),
	}
}

// Register is create-or-patch (get, then create or patch), so it is safe to repeat.
func (r *Registrar) Register(ctx context.Context) error {
	if err := r.validate(); err != nil {
		return err
	}
	if err := r.labelPod(ctx); err != nil {
		return err
	}
	if err := r.ensureService(ctx); err != nil {
		return err
	}
	return r.ensureServiceExport(ctx)
}

func (r *Registrar) labelPod(ctx context.Context) error {
	name := r.Env.PerPodName()
	pod := &core.Pod{}
	if err := r.KC.Get(ctx, client.ObjectKey{Namespace: r.Env.PodNamespace, Name: r.Env.PodName}, pod); err != nil {
		return errors.Wrapf(err, "failed to get pod %s/%s", r.Env.PodNamespace, r.Env.PodName)
	}
	if pod.Labels[PodNameLabelKey] == name {
		return nil
	}
	orig := pod.DeepCopy()
	if pod.Labels == nil {
		pod.Labels = map[string]string{}
	}
	pod.Labels[PodNameLabelKey] = name
	if err := r.KC.Patch(ctx, pod, client.MergeFrom(orig)); err != nil {
		return errors.Wrapf(err, "failed to label pod %s/%s", r.Env.PodNamespace, r.Env.PodName)
	}
	klog.Infof("labeled pod %s/%s with %s=%s", r.Env.PodNamespace, r.Env.PodName, PodNameLabelKey, name)
	return nil
}

func upsertOwnerRef(refs []metav1.OwnerReference, ref metav1.OwnerReference) []metav1.OwnerReference {
	for i := range refs {
		if refs[i].UID == ref.UID {
			refs[i] = ref
			return refs
		}
	}
	return append(refs, ref)
}

func (r *Registrar) ensureService(ctx context.Context) error {
	name := r.Env.PerPodName()
	transform := func(svc *core.Service) {
		svc.OwnerReferences = upsertOwnerRef(svc.OwnerReferences, r.ownerRef())
		svc.Spec.Type = core.ServiceTypeClusterIP
		svc.Spec.ClusterIP = core.ClusterIPNone
		svc.Spec.PublishNotReadyAddresses = true
		svc.Spec.Selector = map[string]string{PodNameLabelKey: name}
		svc.Spec.Ports = []core.ServicePort{
			{
				Name:       vaultAPIPortName,
				Port:       vaultAPIPort,
				TargetPort: intstr.FromInt32(vaultAPIPort),
				Protocol:   core.ProtocolTCP,
			},
			{
				Name:       vaultClusterPortName,
				Port:       vaultClusterPort,
				TargetPort: intstr.FromInt32(vaultClusterPort),
				Protocol:   core.ProtocolTCP,
			},
		}
	}

	svc := &core.Service{}
	err := r.KC.Get(ctx, client.ObjectKey{Namespace: r.Env.PodNamespace, Name: name}, svc)
	if kerr.IsNotFound(err) {
		svc = &core.Service{ObjectMeta: metav1.ObjectMeta{Namespace: r.Env.PodNamespace, Name: name}}
		transform(svc)
		if err := r.KC.Create(ctx, svc); err != nil {
			return errors.Wrapf(err, "failed to create service %s/%s", r.Env.PodNamespace, name)
		}
		klog.Infof("created service %s/%s", r.Env.PodNamespace, name)
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "failed to get service %s/%s", r.Env.PodNamespace, name)
	}

	orig := svc.DeepCopy()
	transform(svc)
	if err := r.KC.Patch(ctx, svc, client.MergeFrom(orig)); err != nil {
		return errors.Wrapf(err, "failed to patch service %s/%s", r.Env.PodNamespace, name)
	}
	return nil
}

func (r *Registrar) ensureServiceExport(ctx context.Context) error {
	name := r.Env.PerPodName()
	transform := func(se *kubesliceapi.ServiceExport) {
		se.OwnerReferences = upsertOwnerRef(se.OwnerReferences, r.ownerRef())
		se.Spec.Slice = r.SliceName
		se.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{PodNameLabelKey: name},
		}
		se.Spec.Ports = []kubesliceapi.ServicePort{
			{
				Name:          vaultAPIPortName,
				ContainerPort: vaultAPIPort,
				Protocol:      core.ProtocolTCP,
			},
			{
				Name:          vaultClusterPortName,
				ContainerPort: vaultClusterPort,
				Protocol:      core.ProtocolTCP,
			},
		}
		se.Spec.Aliases = []string{fmt.Sprintf("%s.%s.svc.cluster.local", name, r.Env.PodNamespace)}
	}

	se := &kubesliceapi.ServiceExport{}
	err := r.KC.Get(ctx, client.ObjectKey{Namespace: r.Env.PodNamespace, Name: name}, se)
	if kerr.IsNotFound(err) {
		se = &kubesliceapi.ServiceExport{ObjectMeta: metav1.ObjectMeta{Namespace: r.Env.PodNamespace, Name: name}}
		transform(se)
		if err := r.KC.Create(ctx, se); err != nil {
			return errors.Wrapf(err, "failed to create serviceexport %s/%s", r.Env.PodNamespace, name)
		}
		klog.Infof("created serviceexport %s/%s", r.Env.PodNamespace, name)
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "failed to get serviceexport %s/%s", r.Env.PodNamespace, name)
	}

	orig := se.DeepCopy()
	transform(se)
	if err := r.KC.Patch(ctx, se, client.MergeFrom(orig)); err != nil {
		return errors.Wrapf(err, "failed to patch serviceexport %s/%s", r.Env.PodNamespace, name)
	}
	return nil
}

// Exported reports whether the pod's own ServiceExport lists the pod with an
// NSM IP, i.e. the pod is reachable from the other data centers.
func (r *Registrar) Exported(ctx context.Context) (bool, error) {
	name := r.Env.PerPodName()
	se := &kubesliceapi.ServiceExport{}
	if err := r.KC.Get(ctx, client.ObjectKey{Namespace: r.Env.PodNamespace, Name: name}, se); err != nil {
		return false, errors.Wrapf(err, "failed to get serviceexport %s/%s", r.Env.PodNamespace, name)
	}
	for _, p := range se.Status.Pods {
		if p.Name == r.Env.PodName && p.NsmIP != "" {
			return true, nil
		}
	}
	return false, nil
}

// ReadyToUnseal registers the pod and checks that its export lists it.
// It returns an error when unsealing must wait.
func (r *Registrar) ReadyToUnseal() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := r.Register(ctx); err != nil {
		return errors.Wrap(err, "failed to self-register the pod")
	}
	exported, err := r.Exported(ctx)
	if err != nil {
		return err
	}
	if !exported {
		return errors.Errorf("serviceexport %s/%s does not list pod %s with an NSM IP yet", r.Env.PodNamespace, r.Env.PerPodName(), r.Env.PodName)
	}
	return nil
}

func parseMembers(v string) map[string]bool {
	out := map[string]bool{}
	for _, m := range strings.Split(v, ",") {
		if m = strings.TrimSpace(m); m != "" {
			out[m] = true
		}
	}
	return out
}

// memberList keeps DC_DR_MEMBERS in order (the PlacementPolicy's).
func memberList(v string) []string {
	var out []string
	for _, m := range strings.Split(v, ",") {
		if m = strings.TrimSpace(m); m != "" {
			out = append(out, m)
		}
	}
	return out
}
