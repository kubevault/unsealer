package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appscode/pat"
	"github.com/stretchr/testify/assert"
	"kubevault.dev/unsealer/pkg/vault"
)

const authList = `
{
"data": {
	  "github/": {
		"type": "github",
		"description": "GitHub auth"
	  },
	  "token/": {
		"config": {
		  "default_lease_ttl": 0,
		  "max_lease_ttl": 0
		},
		"description": "token based credentials",
		"type": "token"
	  }
	}
}
`

func NewFakeVaultServer() *httptest.Server {
	m := pat.New()
	m.Post("/v1/auth/kubernetes/config", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v map[string]interface{}
		defer r.Body.Close()
		json.NewDecoder(r.Body).Decode(&v)
		fmt.Println("***")
		fmt.Println(v)
		fmt.Println("***")
		if val, ok := v["kubernetes_host"]; !ok || (val.(string) != "ok") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if val, ok := v["kubernetes_ca_cert"]; !ok || (val.(string) != "ok") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if val, ok := v["token_reviewer_jwt"]; !ok || (val.(string) != "ok") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	m.Get("/v1/sys/auth", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(authList))
		w.WriteHeader(http.StatusOK)
	}))
	m.Post("/v1/sys/auth/kubernetes", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	return httptest.NewServer(m)
}

func TestEnsureKubernetesAuth2(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()

	cases := []struct {
		testName  string
		expectErr bool
	}{
		{
			testName:  "no error",
			expectErr: false,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			vc, err := vault.NewVaultClient(srv.URL, true, nil)
			if assert.Nil(t, err) {
				k := NewKubernetesAuthenticator(vc, nil)
				err = k.EnsureAuth()
				if c.expectErr {
					assert.NotNil(t, err)
				} else {
					assert.Nil(t, err)
				}
			}
		})
	}
}

func TestConfigureKubernetesAuth2(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()

	cases := []struct {
		testName  string
		k8sHost   string
		k8sCA     string
		jwt       string
		expectErr bool
	}{
		{
			testName:  "no error",
			k8sHost:   "ok",
			k8sCA:     "ok",
			jwt:       "ok",
			expectErr: false,
		},
		{
			testName:  "error, bad k8s host",
			k8sHost:   "bad",
			k8sCA:     "ok",
			jwt:       "ok",
			expectErr: true,
		},
		{
			testName:  "error, bad k8s ca",
			k8sHost:   "ok",
			k8sCA:     "bad",
			jwt:       "ok",
			expectErr: true,
		},
		{
			testName:  "error, bad jwt",
			k8sHost:   "ok",
			k8sCA:     "ok",
			jwt:       "bad",
			expectErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			vc, err := vault.NewVaultClient(srv.URL, true, nil)
			if assert.Nil(t, err) {
				k := NewKubernetesAuthenticator(vc, &K8sAuthenticatorOptions{c.k8sHost, c.k8sCA, c.jwt})
				err = k.ConfigureAuth()
				if c.expectErr {
					assert.NotNil(t, err)
				} else {
					assert.Nil(t, err)
				}
			}
		})
	}
}

func TestEnsureKubernetesAuth(t *testing.T) {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		t.Skip()
	}
	vc, err := vault.NewVaultClient(addr, true, nil)
	vc.SetToken(token)
	if !assert.Nil(t, err) {
		return
	}
	k := NewKubernetesAuthenticator(vc, nil)
	err = k.EnsureAuth()
	assert.Nil(t, err)
}

func TestConfigureKubernetesAuth(t *testing.T) {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	k8s_host := os.Getenv("K8S_HOST")
	k8s_ca := os.Getenv("K8S_CA")
	jwt := os.Getenv("K8S_JWT")
	if addr == "" || token == "" || k8s_host == "" || k8s_ca == "" || jwt == "" {
		t.Skip()
	}
	vc, err := vault.NewVaultClient(addr, true, nil)
	vc.SetToken(token)
	if !assert.Nil(t, err) {
		return
	}

	k := NewKubernetesAuthenticator(vc, &K8sAuthenticatorOptions{k8s_host, k8s_ca, jwt})
	err = k.ConfigureAuth()
	assert.Nil(t, err)
}
