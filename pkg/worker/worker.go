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

	"github.com/golang/glog"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
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

	unsl, err := unseal.New(keyStore, vc, *o.UnsealerOptions)
	if err != nil {
		glog.Error("failed create unsealer client:", err)
	}

	period := time.Second

	for {
		time.Sleep(period)
		period = retryPeriod

		glog.Infoln("checking if vault is initialized...")

		initialized, err := unsl.IsInitialized()
		if err != nil {
			glog.Error("failed to get initialized status. reason :", err)
			continue
		}

		if !initialized {
			glog.Infoln("initialize vault")

			if err = unsl.CheckReadWriteAccess(); err != nil {
				glog.Errorf("Failed to check read/write access to key store. reason: %v\n", err)
				continue
			}

			if err = unsl.Init(); err != nil {
				glog.Error("error initializing vault: ", err)
				continue
			}
		}

		glog.Infoln("checking if vault is sealed...")

		sealed, err := unsl.IsSealed()
		if err != nil {
			glog.Error("failed to get unseal status. reason: ", err)
			continue
		}

		if !sealed {
			glog.Infoln("vault is unsealed")
			continue
		}

		glog.Infoln("unseal vault")

		if err := unsl.Unseal(); err != nil {
			glog.Error("failed to unseal vault. reason: ", err)
			continue
		}

		for {
			glog.Infoln("configure vault")

			err := o.configureVault(vc, keyStore, rootTokenID)
			if err == nil {
				glog.Infoln("vault is configured")
				break
			}

			glog.Error("failed to configure vault. reason: ", err)
		}
	}
}

// configureVault will do:
//	- enable and configure kubernetes auth
//	- create policy and policy binding
func (o *WorkerOptions) configureVault(vc *vaultapi.Client, keyStore kv.Service, rootTokenID string) error {
	rootToken, err := keyStore.Get(rootTokenID)
	if err != nil {
		return errors.Wrap(err, "failed to get root token")
	}
	vc.SetToken(string(rootToken))

	k8sAuth := auth.NewKubernetesAuthenticator(vc, o.AuthenticatorOptions)

	glog.Infoln("enable kubernetes auth")
	err = k8sAuth.EnsureAuth()
	if err != nil {
		return errors.Wrap(err, "failed to enable kubernetes auth")
	}
	glog.Infoln("kubernetes auth is enabled")

	glog.Infoln("configure kubernetes auth")
	err = k8sAuth.ConfigureAuth()
	if err != nil {
		return errors.Wrap(err, "failed to configure kubernetes auth")
	}
	glog.Infoln("kubernetes auth is configured")

	glog.Infoln("write policy and policy binding for policy controller")
	err = policy.EnsurePolicyAndPolicyBinding(vc, o.PolicyManagerOptions)
	if err != nil {
		return errors.Wrap(err, "failed to write policy and policy binding for policy controller")
	}
	glog.Infoln("policy for policy and policy binding controller is written")
	return nil
}

func (o *WorkerOptions) getKVService() (kv.Service, error) {
	if o.Mode == ModeAwsKmsSsm {
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
	}
	if o.Mode == ModeGoogleCloudKmsGCS {
		gcsService, err := gcs.New(o.GoogleOptions.StorageBucket, o.GoogleOptions.StoragePrefix)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create google gcs service")
		}

		kvService, err := cloudkms.New(gcsService, o.GoogleOptions.KmsProject, o.GoogleOptions.KmsLocation, o.GoogleOptions.KmsKeyRing, o.GoogleOptions.KmsCryptoKey)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create kv service for aws")
		}

		return kvService, nil
	}
	if o.Mode == ModeAzureKeyVault {
		kvService, err := azure.NewKVService(o.AzureOptions)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create azure kv service")
		}

		return kvService, nil
	}
	if o.Mode == ModeKubernetesSecret {
		kvService, err := kubernetes.NewKVService(o.KubernetesOptions)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create kv service for kubernetes")
		}

		return kvService, nil
	}

	return nil, errors.New("Invalid mode")
}
