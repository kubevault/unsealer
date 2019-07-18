package policy

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appscode/pat"
	"github.com/stretchr/testify/assert"
	"kubevault.dev/unsealer/pkg/vault"
)

func NewFakeVaultServer() *httptest.Server {
	m := pat.New()
	m.Put("/v1/sys/policy/ok", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	m.Put("/v1/sys/policy/err", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	m.Post("/v1/auth/kubernetes/role/ok", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	return httptest.NewServer(m)
}

func TestEnsurePolicyAndPolicyBinding2(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()

	cases := []struct {
		testName    string
		name        string
		saName      string
		saNamespace string
		expectErr   bool
	}{
		{
			testName:    "no error",
			name:        "ok",
			saName:      "ok",
			saNamespace: "ok",
			expectErr:   false,
		},
		{
			testName:    "error, when creating policy",
			name:        "err",
			saName:      "ok",
			saNamespace: "ok",
			expectErr:   true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			vc, err := vault.NewVaultClient(srv.URL, true, nil)
			if assert.Nil(t, err) {
				err = EnsurePolicyAndPolicyBinding(vc, &PolicyManagerOptions{c.name, c.saName, c.saNamespace})
				if c.expectErr {
					assert.NotNil(t, err)
				} else {
					assert.Nil(t, err)
				}
			}
		})
	}
}

func TestConfigureKubernetesAuth(t *testing.T) {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	policy := os.Getenv("VAULT_POLICY_NAME")
	saName := os.Getenv("POLICY_SA_NAME")
	saNamespace := os.Getenv("POLICY_SA_NAMESPACE")

	if addr == "" || token == "" || policy == "" || saName == "" || saNamespace == "" {
		t.Skip()
	}
	vc, err := vault.NewVaultClient(addr, true, nil)
	vc.SetToken(token)
	if !assert.Nil(t, err) {
		return
	}

	err = EnsurePolicyAndPolicyBinding(vc, &PolicyManagerOptions{policy, saName, saNamespace})
	assert.Nil(t, err)
}
