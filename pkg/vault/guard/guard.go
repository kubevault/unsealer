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

package guard

import (
	"context"
	"net/http"
	"time"

	"kubevault.dev/unsealer/pkg/vault"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

const (
	// NoRequestForwardingHeader makes a standby answer with a redirect instead of
	// forwarding the request to the active node.
	NoRequestForwardingHeader = "X-Vault-No-Request-Forwarding"

	clientTimeout = 5 * time.Second
)

// Client talks to the local Vault node only. It is used by the cross data
// center guard and is independent of the unseal loop's client.
type Client struct {
	// local sends X-Vault-No-Request-Forwarding and never follows redirects,
	// so every request is answered by the local node itself.
	local *vaultapi.Client
	// fwd allows forwarding to the active node (used for renew-self).
	fwd *vaultapi.Client
}

func New(addr string, insecureSkipVerify bool, caCert []byte) (*Client, error) {
	local, err := vault.NewVaultClientWithoutRedirects(addr, insecureSkipVerify, caCert)
	if err != nil {
		return nil, err
	}
	local.AddHeader(NoRequestForwardingHeader, "true")
	local.SetClientTimeout(clientTimeout)
	local.SetMaxRetries(0)
	local.ClearToken()

	fwd, err := vault.NewVaultClient(addr, insecureSkipVerify, caCert)
	if err != nil {
		return nil, err
	}
	fwd.SetClientTimeout(clientTimeout)
	fwd.SetMaxRetries(0)
	fwd.ClearToken()

	return &Client{local: local, fwd: fwd}, nil
}

// Health returns whether the local node is active (initialized, unsealed and
// not a standby) and whether it is unsealed.
func (c *Client) Health(ctx context.Context) (active bool, unsealed bool, err error) {
	resp, err := c.local.Sys().HealthWithContext(ctx)
	if err != nil {
		return false, false, errors.Wrap(err, "failed to get health status")
	}
	unsealed = resp.Initialized && !resp.Sealed
	active = unsealed && !resp.Standby && !resp.PerformanceStandby
	return active, unsealed, nil
}

// StepDown asks the local node to give up the HA lock. The request is never
// forwarded and redirects are not followed: a redirect means the local node is
// no longer active, and false is returned without error.
func (c *Client) StepDown(ctx context.Context, token string) (bool, error) {
	c.local.SetToken(token)
	defer c.local.ClearToken()

	resp, err := c.local.Logical().WriteRawWithContext(ctx, "sys/step-down", nil)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return false, errors.Wrap(err, "failed to step down")
	}
	if resp != nil && resp.StatusCode >= http.StatusMultipleChoices && resp.StatusCode < http.StatusBadRequest {
		// standby answered with a redirect to the active node
		return false, nil
	}
	return true, nil
}

// RenewSelf renews the given token. From a standby the request is forwarded
// to the active node.
func (c *Client) RenewSelf(ctx context.Context, token string) error {
	c.fwd.SetToken(token)
	defer c.fwd.ClearToken()

	if _, err := c.fwd.Auth().Token().RenewSelfWithContext(ctx, 0); err != nil {
		return errors.Wrap(err, "failed to renew token")
	}
	return nil
}
