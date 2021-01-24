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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"kubevault.dev/unsealer/pkg/vault"

	"github.com/appscode/pat"
	"github.com/stretchr/testify/assert"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
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
		utilruntime.Must(json.NewDecoder(r.Body).Decode(&v))
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
		_, err := w.Write([]byte(authList))
		utilruntime.Must(err)
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
	utilruntime.Must(err)
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
	k8sHOST := os.Getenv("K8S_HOST")
	k8sCA := os.Getenv("K8S_CA")
	jwt := os.Getenv("K8S_JWT")
	if addr == "" || token == "" || k8sHOST == "" || k8sCA == "" || jwt == "" {
		t.Skip()
	}
	vc, err := vault.NewVaultClient(addr, true, nil)
	utilruntime.Must(err)
	vc.SetToken(token)
	if !assert.Nil(t, err) {
		return
	}

	k := NewKubernetesAuthenticator(vc, &K8sAuthenticatorOptions{k8sHOST, k8sCA, jwt})
	err = k.ConfigureAuth()
	assert.Nil(t, err)
}
