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

package auth

import (
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

const (
	kubernetesAuthType = "kubernetes"
	kubernetesAuthPath = "kubernetes/"
)

type Authenticator interface {
	EnsureAuth() error
	ConfigureAuth() error
}

type KubernetesAuthenticator struct {
	vc     *vaultapi.Client
	config *K8sAuthenticatorOptions
}

func NewKubernetesAuthenticator(vc *vaultapi.Client, cf *K8sAuthenticatorOptions) Authenticator {
	return &KubernetesAuthenticator{
		vc:     vc,
		config: cf,
	}
}

// EnsureAuth will ensure kubernetes auth
// it's safe to call multiple times
func (k *KubernetesAuthenticator) EnsureAuth() error {
	if k.vc == nil {
		return errors.New("vault client is nil")
	}

	authList, err := k.vc.Sys().ListAuth()
	if err != nil {
		return err
	}
	for path, auth := range authList {
		if auth.Type == kubernetesAuthType && path == kubernetesAuthPath {
			// kubernetes auth already enabled
			return nil
		}
	}

	err = k.vc.Sys().EnableAuthWithOptions(kubernetesAuthPath, &vaultapi.EnableAuthOptions{
		Type: kubernetesAuthType,
	})
	return err
}

// links: https://www.vaultproject.io/api/auth/kubernetes/index.html#configure-method
// ConfigureAuth will set the kubernetes config
// it's safe to call multiple times
func (k *KubernetesAuthenticator) ConfigureAuth() error {
	if k.vc == nil {
		return errors.New("vault client is nil")
	}
	if k.config == nil {
		return errors.New("kubernetes config is nil")
	}

	payload := map[string]any{
		"kubernetes_host":        k.config.Host,
		"kubernetes_ca_cert":     k.config.CA,
		"token_reviewer_jwt":     k.config.Token,
		"disable_iss_validation": true,
		"disable_local_ca_jwt":   true,
	}

	_, err := k.vc.Logical().Write("auth/kubernetes/config", payload)
	return err
}
