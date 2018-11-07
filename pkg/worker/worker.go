package worker

import (
	"time"

	"github.com/golang/glog"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/kubevault/unsealer/pkg/kv"
	"github.com/kubevault/unsealer/pkg/kv/aws_kms"
	"github.com/kubevault/unsealer/pkg/kv/aws_ssm"
	"github.com/kubevault/unsealer/pkg/kv/azure"
	"github.com/kubevault/unsealer/pkg/kv/cloudkms"
	"github.com/kubevault/unsealer/pkg/kv/gcs"
	"github.com/kubevault/unsealer/pkg/kv/kubernetes"
	"github.com/kubevault/unsealer/pkg/vault"
	"github.com/kubevault/unsealer/pkg/vault/auth"
	"github.com/kubevault/unsealer/pkg/vault/policy"
	"github.com/kubevault/unsealer/pkg/vault/unseal"
	"github.com/kubevault/unsealer/pkg/vault/util"
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

	for {
		glog.Infoln("checking if vault is initialized...")
		initialized, err := unsl.IsInitialized()
		if err != nil {
			glog.Error("failed to get initialized status. reason :", err)

		} else {
			if !initialized {
				if err = unsl.CheckReadWriteAccess(); err != nil {
					glog.Errorf("Failed to check read/write access to key store. reason: %v\n", err)
					continue
				}
				if err = unsl.Init(); err != nil {
					glog.Error("error initializing vault: ", err)
				} else {
					glog.Infoln("vault is initialized")
				}
			} else {
				glog.Infoln("vault is already initialized")

				glog.Infoln("checking if vault is sealed...")
				sealed, err := unsl.IsSealed()
				if err != nil {
					glog.Error("failed to get unseal status. reason: ", err)
				} else {
					if sealed {
						if err := unsl.Unseal(); err != nil {
							glog.Error("failed to unseal vault. reason: ", err)
						} else {
							glog.Infoln("vault is unsealed")

							for {
								glog.Infoln("configure vault")
								err := o.configureVault(vc, keyStore, rootTokenID)
								if err != nil {
									glog.Error("failed to configure vault. reason: ", err)
								} else {
									glog.Infoln("vault is configured")
									break
								}
							}

						}
					} else {
						glog.Infoln("vault is unsealed")
					}
				}
			}
		}

		time.Sleep(retryPeriod)
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
