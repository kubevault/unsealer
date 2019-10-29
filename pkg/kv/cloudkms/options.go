/*
Copyright The KubeVault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cloudkms

import (
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

type Options struct {
	// TODO: use kms key id

	KmsCryptoKey string
	KmsKeyRing   string
	KmsLocation  string
	KmsProject   string

	StorageBucket string // name of the Google Cloud Storage bucket to store values in
	// TODO: should make it auto generated
	StoragePrefix string // prefix to use for values store in Google Cloud Storage
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.KmsCryptoKey, "google.kms-crypto-key", o.KmsCryptoKey, "The name of the Google Cloud KMS crypto key to use")
	fs.StringVar(&o.KmsKeyRing, "google.kms-key-ring", o.KmsKeyRing, "The name of the Google Cloud KMS key ring to use")
	fs.StringVar(&o.KmsLocation, "google.kms-location", o.KmsLocation, "The Google Cloud KMS location to use (eg. 'global', 'europe-west1')")
	fs.StringVar(&o.KmsProject, "google.kms-project", o.KmsProject, "The Google Cloud KMS project to use")
	fs.StringVar(&o.StorageBucket, "google.storage-bucket", o.StorageBucket, "The name of the Google Cloud Storage bucket to store values in")
	fs.StringVar(&o.StoragePrefix, "google.storage-prefix", o.StoragePrefix, "The prefix to use for values store in Google Cloud Storage")
}

func (o *Options) Validate() []error {
	var errs []error
	if o.KmsCryptoKey == "" {
		errs = append(errs, errors.New("google kms crypto key must be non-empty"))
	}
	if o.KmsKeyRing == "" {
		errs = append(errs, errors.New("google kms key ring must be non-empty"))
	}
	if o.KmsLocation == "" {
		errs = append(errs, errors.New("google kms location must be non-empty"))
	}
	if o.KmsProject == "" {
		errs = append(errs, errors.New("google kms project must be non-empty"))
	}
	if o.StorageBucket == "" {
		errs = append(errs, errors.New("google storage bucket name must be non-empty"))
	}
	return errs
}

func (o *Options) Apply() error {
	return nil
}
