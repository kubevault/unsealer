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

	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
		return false, fmt.Errorf("error checking status: %s", err.Error())
	}
	return resp.Sealed, nil
}

func (u *unsealer) IsInitialized() (bool, error) {
	resp, err := u.cl.Sys().InitStatus()
	if err != nil {
		return false, fmt.Errorf("error checking init status: %s", err.Error())
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

		logrus.Debugf("retrieving key from kms service...")
		k, err := u.keyStore.Get(keyID)

		if err != nil {
			return fmt.Errorf("unable to get key '%s': %s", keyID, err.Error())
		}

		logrus.Debugf("sending unseal request to vault...")
		resp, err := u.cl.Sys().Unseal(string(k))

		if err != nil {
			return fmt.Errorf("fail to send unseal request to vault: %s", err.Error())
		}

		logrus.Debugf("got unseal response: %+v", *resp)

		if !resp.Sealed {
			return nil
		}

		// if progress is 0, we failed to unseal vault.
		if resp.Progress == 0 {
			return fmt.Errorf("failed to unseal vault. progress reset to 0")
		}
	}
}

func (u *unsealer) keyStoreNotFound(key string) bool {
	_, err := u.keyStore.Get(key)
	if err != nil {
		glog.Errorf("error response when checking whether key(%s) exists or not: %v", key, err)
	}
	if _, ok := err.(*kv.NotFoundError); ok {
		return true
	}
	return false
}

func (u *unsealer) keyStoreSet(key string, val []byte) error {
	if !u.config.OverwriteExisting && !u.keyStoreNotFound(key) {
		return fmt.Errorf("error setting key '%s': it already exists", key)
	}
	return u.keyStore.Set(key, val)
}

func (u *unsealer) Init() error {
	// test backend first
	err := u.keyStore.Test(testKey(u.config.KeyPrefix))
	if err != nil {
		return fmt.Errorf("error testing keystore before init: %s", err.Error())
	}

	// test for an existing keys
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
		return fmt.Errorf("error initialising vault: %s", err.Error())
	}

	for i, k := range resp.Keys {
		keyID := util.UnsealKeyID(u.config.KeyPrefix, i)
		err := u.keyStoreSet(keyID, []byte(k))

		if err != nil {
			return fmt.Errorf("error storing unseal key '%s': %s", keyID, err.Error())
		}
	}

	rootToken := resp.RootToken

	if u.config.StoreRootToken {
		rootTokenID := util.RootTokenID(u.config.KeyPrefix)
		if err = u.keyStoreSet(rootTokenID, []byte(resp.RootToken)); err != nil {
			return fmt.Errorf("error storing root token '%s' in key'%s'", rootToken, rootTokenID)
		}
		logrus.WithField("key", rootTokenID).Info("root token stored in key store")
	} else {
		logrus.WithField("root-token", resp.RootToken).Warnf("won't store root token in key store, this token grants full privileges to vault, so keep this secret")
	}

	return nil

}

func testKey(prefix string) string {
	return fmt.Sprintf("%s-test", prefix)
}

// CheckReadWriteAccess will test read write access
func (u *unsealer) CheckReadWriteAccess() error {
	glog.Infoln("Testing the read/write access...")

	err := u.keyStore.CheckWriteAccess()
	if err != nil {
		return errors.Wrap(err, "read/write access test failed")
	}

	glog.Infoln("Testing the read/write access is successful")
	return nil
}
