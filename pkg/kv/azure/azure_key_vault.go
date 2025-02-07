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

package azure

import (
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"kubevault.dev/unsealer/pkg/kv"

	azurekv "github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/pkg/errors"
)

type KVService struct {
	KeyClient    azurekv.BaseClient
	Ctx          context.Context
	VaultBaseUrl string
	SecretPrefix string
}

func NewKVService(opts *Options) (kv.Service, error) {
	k := &KVService{
		Ctx:          context.Background(),
		VaultBaseUrl: opts.VaultBaseUrl,
		SecretPrefix: opts.SecretPrefix,
	}

	auth, err := opts.AuthConfig.GetKeyVaultToken(AuthGrantType())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get  OAuth token for key vault resources")
	}

	k.KeyClient = azurekv.New()
	k.KeyClient.Authorizer = auth
	err = k.KeyClient.AddToUserAgent(azurekv.UserAgent())
	if err != nil {
		return nil, errors.Wrap(err, "failed to add to user agent")
	}
	return k, nil
}

func (k *KVService) Set(key string, value []byte) error {
	data := base64.StdEncoding.EncodeToString(value)
	return k.SetSecret(strings.Replace(k.getKeyName(key), ".", "-", -1), data)
}

func (k *KVService) Get(key string) ([]byte, error) {
	data, err := k.GetSecret(strings.Replace(k.getKeyName(key), ".", "-", -1))
	if err != nil {
		return nil, kv.NewNotFoundError("unable to get secret(%s) from key vault. reason: %v", key, err)
	}

	value, err := base64.StdEncoding.DecodeString(to.String(data))
	if err != nil {
		return nil, kv.NewNotFoundError("failed to decode base64 string. reason: %v", err)
	}

	return value, nil
}

func (k *KVService) CheckWriteAccess() error {
	key := "vault-unsealer-dummy-file"
	val := "read write access check"

	err := k.Set(key, []byte(val))
	if err != nil {
		return errors.Wrap(err, "failed to write test file")
	}

	_, err = k.Get(key)
	if err != nil {
		return errors.Wrap(err, "failed to get test file")
	}

	_, err = k.KeyClient.DeleteSecret(k.Ctx, k.VaultBaseUrl, key)
	if err != nil {
		return errors.Wrap(err, "failed to delete test file")
	}

	return nil
}

func (k *KVService) Test(key string) error {
	return nil
}

// SetSecret will store secret in azure key vault
func (k *KVService) SetSecret(secretName, value string) error {
	parameter := azurekv.SecretSetParameters{
		Value:       to.StringPtr(value),
		ContentType: to.StringPtr("password"),
	}

	_, err := k.KeyClient.SetSecret(k.Ctx, k.VaultBaseUrl, secretName, parameter)
	if err != nil {
		return errors.Wrap(err, "unable to set secrets in key vault")
	}

	return nil
}

// GetSecret will give secret in response
func (k *KVService) GetSecret(secretName string) (*string, error) {
	version, err := k.GetLatestVersionOfSecret(k.VaultBaseUrl, secretName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get latest version of secret")
	}
	sr, err := k.KeyClient.GetSecret(k.Ctx, k.VaultBaseUrl, secretName, version)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get secret(%s) of version(%s)", secretName, version)
	}

	return sr.Value, nil
}

// GetLatestVersionOfSecret will give latest version of secret according to created time
func (k *KVService) GetLatestVersionOfSecret(vaultBaseUrl, secretName string) (string, error) {
	var version string
	var createdTime time.Duration
	resp, err := k.KeyClient.GetSecretVersions(k.Ctx, vaultBaseUrl, secretName, to.Int32Ptr(20))
	if err != nil {
		return "", errors.Wrap(err, "unable to get secret versions")
	}

	// finding latest version according to created time
	for resp.NotDone() {
		items := resp.Values()
		for _, item := range items {
			if version == "" {
				version = filepath.Base(to.String(item.ID))
				createdTime = item.Attributes.Created.Duration()
			} else if createdTime < item.Attributes.Created.Duration() {
				version = filepath.Base(to.String(item.ID))
				createdTime = item.Attributes.Created.Duration()
			}
		}

		err = resp.Next()
		if err != nil {
			return "", errors.Wrap(err, "unable to get next pages of version")
		}
	}
	return version, nil
}

func (k *KVService) getKeyName(key string) string {
	return fmt.Sprintf("%s%s", k.SecretPrefix, key)
}
