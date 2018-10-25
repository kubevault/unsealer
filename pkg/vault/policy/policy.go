package policy

import (
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

const policyAdmin = `
path "sys/policy/*" {
	capabilities = ["create", "update", "read", "delete", "list"]
}

path "sys/policy" {
	capabilities = ["read", "list"]
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
func EnsurePolicyAndPolicyBinding(vc *vaultapi.Client, config *PolicyOptions) error {
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

	path := fmt.Sprintf("/v1/auth/kubernetes/role/%s", config.Name)
	req := vc.NewRequest("POST", path)
	payload := map[string]interface{}{
		"bound_service_account_names":      config.ServiceAccountName,
		"bound_service_account_namespaces": config.ServiceAccountNamespace,
		"policies":                         []string{config.Name, "default"},
		"ttl":                              "24h",
		"max_ttl":                          "24h",
		"period":                           "24h",
	}

	err = req.SetJSONBody(payload)
	if err != nil {
		return err
	}

	_, err = vc.RawRequest(req)
	if err != nil {
		return err
	}
	return nil
}
