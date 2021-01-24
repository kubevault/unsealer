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

package cloudkms

import (
	"context"
	"encoding/base64"
	"fmt"

	"kubevault.dev/unsealer/pkg/kv"

	cloudkms "google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/option"
)

// googleKms is an implementation of the kv.Service interface, that encrypts
// and decrypts data using Google Cloud KMS before storing into another kv
// backend.
type googleKms struct {
	svc     *cloudkms.Service
	store   kv.Service
	keyPath string
}

var _ kv.Service = &googleKms{}

func New(store kv.Service, project, location, keyring, cryptoKey string) (kv.Service, error) {
	ctx := context.Background()
	kmsService, err := cloudkms.NewService(ctx, option.WithScopes(cloudkms.CloudPlatformScope))
	if err != nil {
		return nil, fmt.Errorf("error creating google kms service client: %s", err.Error())
	}

	return &googleKms{
		store:   store,
		svc:     kmsService,
		keyPath: fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", project, location, keyring, cryptoKey),
	}, nil
}

func (g *googleKms) encrypt(s []byte) ([]byte, error) {
	resp, err := g.svc.Projects.Locations.KeyRings.CryptoKeys.Encrypt(g.keyPath, &cloudkms.EncryptRequest{
		Plaintext: base64.StdEncoding.EncodeToString(s),
	}).Do()

	if err != nil {
		return nil, fmt.Errorf("error encrypting data: %s", err.Error())
	}

	return base64.StdEncoding.DecodeString(resp.Ciphertext)
}

func (g *googleKms) decrypt(s []byte) ([]byte, error) {
	resp, err := g.svc.Projects.Locations.KeyRings.CryptoKeys.Decrypt(g.keyPath, &cloudkms.DecryptRequest{
		Ciphertext: base64.StdEncoding.EncodeToString(s),
	}).Do()

	if err != nil {
		return nil, fmt.Errorf("error decrypting data: %s", err.Error())
	}

	return base64.StdEncoding.DecodeString(resp.Plaintext)
}

func (g *googleKms) Get(key string) ([]byte, error) {
	cipherText, err := g.store.Get(key)

	if err != nil {
		return nil, err
	}

	return g.decrypt(cipherText)
}

func (g *googleKms) Set(key string, val []byte) error {
	cipherText, err := g.encrypt(val)

	if err != nil {
		return err
	}

	return g.store.Set(key, cipherText)
}

func (g *googleKms) CheckWriteAccess() error {
	return g.store.CheckWriteAccess()
}

func (g *googleKms) Test(key string) error {
	// TODO: Implement me properly
	return nil
}
