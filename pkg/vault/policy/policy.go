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

package policy

import (
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

const policyAdmin = `
path "sys/mounts" {
  capabilities = ["read", "list"]
}

path "sys/mounts/*" {
  capabilities = ["create", "read", "update", "delete"]
}

path "sys/leases/revoke/*" {
    capabilities = ["update"]
}

path "sys/policy/*" {
	capabilities = ["create", "update", "read", "delete", "list"]
}

path "sys/policy" {
	capabilities = ["read", "list"]
}

path "sys/policies" {
	capabilities = ["read", "list"]
}

path "sys/policies/*" {
	capabilities = ["create", "update", "read", "delete", "list"]
}

path "auth/kubernetes/role" {
	capabilities = ["read", "list"]
}

path "auth/kubernetes/role/*" {
	capabilities = ["create", "update", "read", "delete", "list"]
}
`

// EnsurePolicyAndPolicyBinding will ensure policy and kubernetes role
// Name of the policy will be 'config.Name'
// Name of the kubernetes role will be 'config.Name'
func EnsurePolicyAndPolicyBinding(vc *vaultapi.Client, config *PolicyManagerOptions) error {
	if vc == nil {
		return errors.New("vault client is nil")
	}
	if config == nil {
		return errors.New("config is nil")
	}

	err := vc.Sys().PutPolicy(config.Name, policyAdmin)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/auth/kubernetes/role/%s", config.Name)
	payload := map[string]interface{}{
		"bound_service_account_names":      config.ServiceAccountName,
		"bound_service_account_namespaces": config.ServiceAccountNamespace,
		"policies":                         []string{config.Name, "default"},
		"ttl":                              "24h",
		"max_ttl":                          "24h",
		"period":                           "24h",
	}

	_, err = vc.Logical().Write(path, payload)
	return err
}
