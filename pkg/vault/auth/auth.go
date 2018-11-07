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

	req := k.vc.NewRequest("POST", "/v1/auth/kubernetes/config")
	payload := map[string]interface{}{
		"kubernetes_host":    k.config.Host,
		"kubernetes_ca_cert": k.config.CA,
		"token_reviewer_jwt": k.config.Token,
	}
	if err := req.SetJSONBody(payload); err != nil {
		return err
	}

	_, err := k.vc.RawRequest(req)
	return err
}
