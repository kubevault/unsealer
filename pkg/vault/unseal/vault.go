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

package unseal

import (
	"fmt"

	"kubevault.dev/unsealer/pkg/kv"
	"kubevault.dev/unsealer/pkg/vault/util"

	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

// unsealer is an implementation of the Unsealer interface that will perform actions
// against a Vault server, using a provided KMS to retrieve
type unsealer struct {
	keyStore kv.Service
	cl       *api.Client
	config   *UnsealOptions
}

var _ Unsealer = &unsealer{}

// Unsealer is an interface that can be used to attempt to perform actions against
// a Vault server.
type Unsealer interface {
	IsSealed() (bool, error)
	IsInitialized() (bool, error)
	Unseal() error
	Init() error
	CheckReadWriteAccess() error
}

// New returns a new Unsealer, or an error.
func New(k kv.Service, cl *api.Client, config UnsealOptions) (Unsealer, error) {
	return &unsealer{
		keyStore: k,
		cl:       cl,
		config:   &config,
	}, nil
}

func (u *unsealer) IsSealed() (bool, error) {
	resp, err := u.cl.Sys().SealStatus()
	if err != nil {
		return false, fmt.Errorf("failed to check seal status with %s", err.Error())
	}
	return resp.Sealed, nil
}

func (u *unsealer) IsInitialized() (bool, error) {
	resp, err := u.cl.Sys().InitStatus()
	if err != nil {
		return false, fmt.Errorf("failed to check initilized status with %s", err.Error())
	}
	return resp, nil
}

// Unseal will attempt to unseal vault by retrieving keys from the kms service
// and sending unseal requests to vault. It will return an error if retrieving
// a key fails, or if the unseal progress is reset to 0 (indicating that a key)
// was invalid.
func (u *unsealer) Unseal() error {
	for i := 0; ; i++ {
		keyID := util.UnsealKeyID(u.config.KeyPrefix, i)

		klog.Infof("try to retrieve key with keyID = %s, from kms service", keyID)
		k, err := u.keyStore.Get(keyID)
		if err != nil {
			return fmt.Errorf("failed to get key = %s with %s", keyID, err.Error())
		}

		klog.Infof("try to send unseal request to the vault with keyID = %s", keyID)
		resp, err := u.cl.Sys().Unseal(string(k))
		if err != nil {
			return fmt.Errorf("failed to send unseal request to the vault with %s", err.Error())
		}

		klog.Infof("got an unseal response: %+v", *resp)

		if !resp.Sealed {
			return nil
		}

		// if progress is 0, we failed to unseal vault.
		if resp.Progress == 0 {
			return fmt.Errorf("failed to unseal the vault, progress is reset to 0")
		}
	}
}

func (u *unsealer) keyStoreNotFound(key string) bool {
	_, err := u.keyStore.Get(key)
	if err != nil {
		klog.Errorf("error while checking whether key (%s) exists or not with %v", key, err)
	}
	if _, ok := err.(*kv.NotFoundError); ok {
		return true
	}
	return false
}

func (u *unsealer) keyStoreSet(key string, val []byte) error {
	// We do not want to overwrite the existing keys, but key is already present.
	if !u.config.OverwriteExisting && !u.keyStoreNotFound(key) {
		return fmt.Errorf("error setting key %s to keystore, it already exists", key)
	}
	return u.keyStore.Set(key, val)
}

func (u *unsealer) Init() error {
	// test the backend first
	err := u.keyStore.Test(testKey(u.config.KeyPrefix))
	if err != nil {
		return fmt.Errorf("error testing keystore before init with %s", err.Error())
	}

	// test for an existing key
	if !u.config.OverwriteExisting {
		keys := []string{
			util.RootTokenID(u.config.KeyPrefix),
		}

		// add unseal keys
		for i := 0; i <= u.config.SecretShares; i++ {
			keys = append(keys, util.UnsealKeyID(u.config.KeyPrefix, i))
		}

		// test every key
		for _, key := range keys {
			if !u.keyStoreNotFound(key) {
				return fmt.Errorf("error before init: keystore value for '%s' already exists", key)
			}
		}
	}

	resp, err := u.cl.Sys().Init(&api.InitRequest{
		SecretShares:    u.config.SecretShares,
		SecretThreshold: u.config.SecretThreshold,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize the vault with %s", err.Error())
	}

	for i, k := range resp.Keys {
		keyID := util.UnsealKeyID(u.config.KeyPrefix, i)
		err := u.keyStoreSet(keyID, []byte(k))
		if err != nil {
			return fmt.Errorf("failed to store the unseal key = '%s' with %s", keyID, err.Error())
		}
	}

	if u.config.StoreRootToken {
		rootTokenID := util.RootTokenID(u.config.KeyPrefix)
		if err = u.keyStoreSet(rootTokenID, []byte(resp.RootToken)); err != nil {
			return fmt.Errorf("failed to store the root token with %v", err)
		}
		klog.Info("successfully stored the root token")
	} else {
		klog.Warning("will not store the rootToken in the key store, this token grants full privileges to vault, so keep this secret")
	}

	return nil
}

func testKey(prefix string) string {
	return fmt.Sprintf("%s-test", prefix)
}

// CheckReadWriteAccess will test read write access
func (u *unsealer) CheckReadWriteAccess() error {
	err := u.keyStore.CheckWriteAccess()
	if err != nil {
		return errors.Wrap(err, "read/write access test failed")
	}

	klog.Info("successfully tested the read/write access")
	return nil
}
