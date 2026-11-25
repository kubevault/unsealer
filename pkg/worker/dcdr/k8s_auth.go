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
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

// Files of the k8s-auth ConfigMap the operator ships and mounts (cloud KMS
// modes), the keys of kubevault/operator pkg/vault/k8sauth.
const (
	k8sAuthFileHost         = "kubernetes_host"
	k8sAuthFileCACert       = "kubernetes_ca_cert"
	k8sAuthFileRole         = "role"
	k8sAuthFileSAName       = "service_account_name"
	k8sAuthFileSANamespace  = "service_account_namespace"
	k8sAuthFilePolicies     = "policies"
	k8sAuthFilePolicyPrefix = "policy."
	k8sAuthFilePolicySuffix = ".hcl"
	k8sAuthFileGeneration   = "generation"

	k8sAuthPath = "kubernetes"
	k8sAuthType = "kubernetes"
	// k8sAuthBootstrapDescriptionPrefix starts the kubernetes mount's
	// description; the rest is the bootstrap generation, in Vault's storage so
	// every data center sees it.
	k8sAuthBootstrapDescriptionPrefix = "kubevault-bootstrap:"

	k8sAuthRetryPeriod = 30 * time.Second
)

// K8sAuthBootstrap bootstraps Vault's Kubernetes auth in the cloud KMS modes,
// where only the Vault pods can read the root token. The unsealer of the active
// node writes, with the root token, what the operator needs to log in: the
// kubernetes mount, its configuration without a reviewer JWT, the policies and
// the operator's role. The operator then keeps the configuration current
// itself. It runs when the generation in the mounted ConfigMap is higher than
// the one recorded in the mount's description; after a failover the new active
// node finds it recorded and does nothing.
type K8sAuthBootstrap struct {
	// Dir is where the k8s-auth ConfigMap is mounted.
	Dir string
	// VC is the unsealer's client for the local node; it is not modified.
	VC *vaultapi.Client
	// RootToken reads the root token from the key store.
	RootToken func() (string, error)

	// done is the highest generation seen recorded in Vault.
	done int64
	next time.Time
}

// Ensure bootstraps when needed; failures are logged and retried after
// k8sAuthRetryPeriod, never blocking the unseal loop.
func (b *K8sAuthBootstrap) Ensure() {
	if time.Now().Before(b.next) {
		return
	}
	if err := b.ensure(); err != nil {
		klog.Errorf("dc-dr: kubernetes auth bootstrap: %s", err.Error())
		b.next = time.Now().Add(k8sAuthRetryPeriod)
	}
}

func (b *K8sAuthBootstrap) ensure() error {
	want, err := b.readGeneration()
	if os.IsNotExist(errors.Cause(err)) {
		klog.V(4).Infof("dc-dr: kubernetes auth bootstrap values not shipped yet (%s)", b.Dir)
		b.next = time.Now().Add(k8sAuthRetryPeriod)
		return nil
	}
	if err != nil {
		return err
	}
	if want <= b.done {
		return nil
	}

	// Only the active node writes: one writer, and no forwarding of writes.
	hr, err := b.VC.Sys().Health()
	if err != nil {
		return errors.Wrap(err, "failed to read sys/health")
	}
	if !hr.Initialized || hr.Sealed || hr.Standby {
		return nil
	}

	root, err := b.RootToken()
	if err != nil {
		return errors.Wrap(err, "failed to read the root token from the key store")
	}
	vc, err := b.VC.Clone()
	if err != nil {
		return errors.Wrap(err, "failed to clone the vault client")
	}
	vc.SetToken(root)

	auths, err := vc.Sys().ListAuth()
	if err != nil {
		return errors.Wrap(err, "failed to list auth methods")
	}
	mount := auths[k8sAuthPath+"/"]
	if mount != nil {
		if mount.Type != k8sAuthType {
			return errors.Errorf("auth path %s/ is mounted with type %s", k8sAuthPath, mount.Type)
		}
		if g, ok := bootstrapGeneration(mount.Description); ok && g >= want {
			b.done = g
			return nil
		}
	}

	klog.Infof("dc-dr: bootstrapping kubernetes auth, generation %d", want)
	if err := b.write(vc, mount == nil, want); err != nil {
		return err
	}
	b.done = want
	klog.Infof("dc-dr: kubernetes auth bootstrapped, generation %d", want)
	return nil
}

// write makes the bootstrap writes, the same as the operator's
// (kubevault/operator pkg/vault/k8sauth Bootstrap), recording the generation
// last so a bootstrap that failed half way is repeated.
func (b *K8sAuthBootstrap) write(vc *vaultapi.Client, enable bool, generation int64) error {
	host, err := b.read(k8sAuthFileHost)
	if err != nil {
		return err
	}
	ca, err := b.read(k8sAuthFileCACert)
	if err != nil {
		return err
	}
	role, err := b.read(k8sAuthFileRole)
	if err != nil {
		return err
	}
	saName, err := b.read(k8sAuthFileSAName)
	if err != nil {
		return err
	}
	saNamespace, err := b.read(k8sAuthFileSANamespace)
	if err != nil {
		return err
	}
	list, err := b.read(k8sAuthFilePolicies)
	if err != nil {
		return err
	}
	var policies []string
	for _, p := range strings.Split(list, ",") {
		if p = strings.TrimSpace(p); p != "" {
			policies = append(policies, p)
		}
	}
	if host == "" || role == "" || saName == "" || saNamespace == "" || len(policies) == 0 {
		return errors.New("incomplete bootstrap values")
	}

	if enable {
		if err := vc.Sys().EnableAuthWithOptions(k8sAuthPath, &vaultapi.EnableAuthOptions{Type: k8sAuthType}); err != nil {
			return errors.Wrap(err, "failed to enable kubernetes auth")
		}
	}
	// No reviewer JWT: Vault reviews each login with the login's own JWT, and the
	// hub lets the operator's ServiceAccount do that. Never the pod's own token,
	// which belongs to this member cluster, not to host.
	if _, err := vc.Logical().Write("auth/"+k8sAuthPath+"/config", map[string]any{
		"kubernetes_host":        host,
		"kubernetes_ca_cert":     ca,
		"token_reviewer_jwt":     "",
		"disable_local_ca_jwt":   true,
		"disable_iss_validation": true,
	}); err != nil {
		return errors.Wrap(err, "failed to configure kubernetes auth")
	}
	for _, name := range policies {
		hcl, err := b.read(k8sAuthFilePolicyPrefix + name + k8sAuthFilePolicySuffix)
		if err != nil {
			return err
		}
		if err := vc.Sys().PutPolicy(name, hcl); err != nil {
			return errors.Wrapf(err, "failed to write policy %s", name)
		}
	}
	if _, err := vc.Logical().Write(fmt.Sprintf("auth/%s/role/%s", k8sAuthPath, role), map[string]any{
		"bound_service_account_names":      saName,
		"bound_service_account_namespaces": saNamespace,
		"policies":                         append(policies, "default"),
		"ttl":                              "24h",
		"max_ttl":                          "24h",
		"period":                           "24h",
	}); err != nil {
		return errors.Wrapf(err, "failed to write kubernetes auth role %s", role)
	}
	desc := k8sAuthBootstrapDescriptionPrefix + strconv.FormatInt(generation, 10)
	if err := vc.Sys().TuneMount("auth/"+k8sAuthPath, vaultapi.MountConfigInput{Description: &desc}); err != nil {
		return errors.Wrap(err, "failed to record the bootstrap generation")
	}
	return nil
}

func (b *K8sAuthBootstrap) readGeneration() (int64, error) {
	s, err := b.read(k8sAuthFileGeneration)
	if err != nil {
		return 0, err
	}
	g, err := strconv.ParseInt(s, 10, 64)
	if err != nil || g < 1 {
		return 0, errors.Errorf("invalid bootstrap generation %q", s)
	}
	return g, nil
}

func (b *K8sAuthBootstrap) read(name string) (string, error) {
	data, err := os.ReadFile(filepath.Join(b.Dir, name))
	if err != nil {
		return "", errors.Wrapf(err, "failed to read %s", name)
	}
	return strings.TrimSpace(string(data)), nil
}

func bootstrapGeneration(description string) (int64, bool) {
	s, ok := strings.CutPrefix(description, k8sAuthBootstrapDescriptionPrefix)
	if !ok {
		return 0, false
	}
	g, err := strconv.ParseInt(s, 10, 64)
	return g, err == nil
}
