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

package worker

import (
	"time"

	"kubevault.dev/unsealer/pkg/kv"
	"kubevault.dev/unsealer/pkg/kv/aws_kms"
	"kubevault.dev/unsealer/pkg/kv/aws_ssm"
	"kubevault.dev/unsealer/pkg/kv/azure"
	"kubevault.dev/unsealer/pkg/kv/cloudkms"
	"kubevault.dev/unsealer/pkg/kv/gcs"
	"kubevault.dev/unsealer/pkg/kv/kubernetes"
	"kubevault.dev/unsealer/pkg/vault"
	"kubevault.dev/unsealer/pkg/vault/auth"
	"kubevault.dev/unsealer/pkg/vault/policy"
	"kubevault.dev/unsealer/pkg/vault/unseal"
	"kubevault.dev/unsealer/pkg/vault/util"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

func (o *WorkerOptions) Run() error {
	keyStore, err := o.getKVService()
	if err != nil {
		return errors.Wrap(err, "failed to create kv service")
	}

	vc, err := vault.NewVaultClient(o.Address, o.InsecureSkipTLSVerify, []byte(o.CaCert))
	if err != nil {
		return errors.Wrap(err, "failed to create vault api client")
	}

	o.unsealAndConfigureVault(vc, keyStore, o.ReTryPeriod)

	return nil
}

// It will do:
//	- If vault is not initialized, then initialize vault
//	- If vault is not unsealed, then unseal it
//  - configure vault
// it will periodically check for infinite time
func (o *WorkerOptions) unsealAndConfigureVault(vc *vaultapi.Client, keyStore kv.Service, retryPeriod time.Duration) {
	rootTokenID := util.RootTokenID(o.UnsealerOptions.KeyPrefix)

	unsealer, err := unseal.New(keyStore, vc, *o.UnsealerOptions)
	if err != nil {
		klog.Errorf("failed to create the unsealer client with %s", err.Error())
	}

	period := time.Second

	for {
		time.Sleep(period)
		period = retryPeriod

		klog.Info("checking if the vault is initialized or not.")

		initialized, err := unsealer.IsInitialized()
		if err != nil {
			klog.Errorf("failed to get the initialized status with %s", err.Error())
			continue
		}

		// the vault is not initialized, check the read/write access & try to initialize the vault
		if !initialized {
			klog.Info("trying to initialize the vault")

			if err = unsealer.CheckReadWriteAccess(); err != nil {
				klog.Errorf("failed to check the read/write access to the key store with %s", err.Error())
				continue
			}

			// try to Initialize the vault
			if err = unsealer.Init(); err != nil {
				klog.Errorf("failed to initialize the vault with %s", err.Error())
				continue
			}
		}

		klog.Infof("vault must be initialized here, initialized value: %v", initialized)
		klog.Infoln("checking if the vault is sealed or not")

		// checking the sealed status of the vault
		sealed, err := unsealer.IsSealed()
		if err != nil {
			klog.Errorf("failed to get the sealed status with %s", err.Error())
			continue
		}

		if !sealed {
			klog.Infoln("vault is unsealed")
			continue
		}

		klog.Infoln("making the unseal vault request")

		if err := unsealer.Unseal(); err != nil {
			klog.Errorf("failed to unseal the vault with %s", err.Error())
			continue
		}

		for {
			klog.Infoln("trying to configure the vault")

			err := o.configureVault(vc, keyStore, rootTokenID)
			if err == nil {
				klog.Infoln("vault is configured")
				break
			}

			klog.Errorf("failed to configure the vault with %s", err.Error())
		}
	}
}

// configureVault will do:
//	- enable and configure kubernetes auth
//	- create policy and policy binding
func (o *WorkerOptions) configureVault(vc *vaultapi.Client, keyStore kv.Service, rootTokenID string) error {
	rootToken, err := keyStore.Get(rootTokenID)
	if err != nil {
		return errors.Wrap(err, "failed to get the root token")
	}

	// set the rootToken
	vc.SetToken(string(rootToken))

	k8sAuth := auth.NewKubernetesAuthenticator(vc, o.AuthenticatorOptions)

	klog.Infoln("enable kubernetes auth")

	err = k8sAuth.EnsureAuth()
	if err != nil {
		return errors.Wrap(err, "failed to enable kubernetes auth")
	}

	klog.Infoln("kubernetes auth is enabled")

	klog.Infoln("configure kubernetes auth")

	err = k8sAuth.ConfigureAuth()
	if err != nil {
		return errors.Wrap(err, "failed to configure kubernetes auth")
	}

	klog.Infoln("kubernetes auth is configured")

	klog.Infoln("write policy and policy binding for policy controller")

	err = policy.EnsurePolicyAndPolicyBinding(vc, o.PolicyManagerOptions)
	if err != nil {
		return errors.Wrap(err, "failed to write policy and policy binding for policy controller")
	}

	klog.Infoln("policy for policy and policy binding controller is written")

	return nil
}

func (o *WorkerOptions) getKVService() (kv.Service, error) {
	switch o.Mode {
	case ModeAwsKmsSsm:
		ssmService, err := aws_ssm.New(o.AwsOptions.UseSecureString, o.AwsOptions.SsmKeyPrefix)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create aws ssm service")
		}

		var kvService kv.Service

		if o.AwsOptions.UseSecureString {
			kvService = ssmService
		} else {
			kvService, err = aws_kms.New(ssmService, o.AwsOptions.KmsKeyID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create kv service for aws")
			}
		}

		return kvService, nil

	case ModeGoogleCloudKmsGCS:
		gcsService, err := gcs.New(o.GoogleOptions.StorageBucket, o.GoogleOptions.StoragePrefix)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create google gcs service")
		}

		kvService, err := cloudkms.New(gcsService, o.GoogleOptions.KmsProject, o.GoogleOptions.KmsLocation, o.GoogleOptions.KmsKeyRing, o.GoogleOptions.KmsCryptoKey)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create kv service for aws")
		}

		return kvService, nil

	case ModeAzureKeyVault:
		kvService, err := azure.NewKVService(o.AzureOptions)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create azure kv service")
		}

		return kvService, nil

	case ModeKubernetesSecret:
		kvService, err := kubernetes.NewKVService(o.KubernetesOptions)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create kv service for kubernetes")
		}

		return kvService, nil

	default:
		return nil, errors.Errorf("failed to create unkown mode %q", o.Mode)
	}
}
