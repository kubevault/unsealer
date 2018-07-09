package worker

import (
	"fmt"
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
	"github.com/pkg/errors"
)

func (o *WorkerOptions) Run() error {

	kvService, err := o.getKVService()
	if err != nil {
		return errors.Wrap(err, "failed to create kv service")
	}

	var tlsConfig *vaultapi.TLSConfig
	if o.InSecureTLS {
		tlsConfig = &vaultapi.TLSConfig{
			Insecure: true,
		}
	} else if o.CaCertFile != "" {
		tlsConfig = &vaultapi.TLSConfig{
			CACert: o.CaCertFile,
		}
	}

	vaultApiClient, err := NewVaultClient("127.0.0.1", "8200", tlsConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create vault api client")
	}

	v, err := vault.New(kvService, vaultApiClient, *o.Vault)
	if err != nil {
		return errors.Wrap(err, "failed create vault helper")
	}

	//initialize
	for {
		glog.Infoln("checking if vault is initialized...")

		initialized, err := vaultApiClient.Sys().InitStatus()
		if err != nil {
			glog.Error("failed to get initialized status. reason :", err)
		} else {
			if !initialized {
				if !o.SkipReadWriteAccessCheck {
					if err = v.CheckReadWriteAccess(); err != nil {
						glog.Errorf("Failed to check read/write access to key store. reason: %v\n", err)
						continue
					}
				}

				if err = v.Init(); err != nil {
					glog.Error("error initializing vault: ", err)
				} else {
					glog.Infoln("vault is initialized")
					break
				}
			} else {
				glog.Infoln("vault is already initialized")
				break
			}
		}

		time.Sleep(o.ReTryPeriod)
	}

	// unseal
	for {
		glog.Infoln("checking if vault is sealed...")

		sealed, err := v.Sealed()
		if err != nil {
			glog.Error("failed to get initialized status. reason: ", err)
		} else {
			if sealed {
				if err := v.Unseal(); err != nil {
					glog.Error("failed to unseal vault. reason: ", err)
				} else {
					glog.Infoln("vault is unsealed")
				}
			} else {
				glog.Infoln("vault is unsealed")
			}
		}

		time.Sleep(o.ReTryPeriod)
	}

	return nil
}

func (o *WorkerOptions) getKVService() (kv.Service, error) {
	if o.Mode == ModeAwsKmsSsm {
		ssmService, err := aws_ssm.New(o.Aws.SsmKeyPrefix)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create aws ssm service")
		}

		kvService, err := aws_kms.New(ssmService, o.Aws.KmsKeyID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create kv service for aws")
		}

		return kvService, nil
	}
	if o.Mode == ModeGoogleCloudKmsGCS {
		gcsService, err := gcs.New(o.Google.StorageBucket, o.Google.StoragePrefix)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create google gcs service")
		}

		kvService, err := cloudkms.New(gcsService, o.Google.KmsProject, o.Google.KmsLocation, o.Google.KmsKeyRing, o.Google.KmsCryptoKey)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create kv service for aws")
		}

		return kvService, nil
	}
	if o.Mode == ModeAzureKeyVault {
		kvService, err := azure.NewKVService(o.Azure)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create azure kv service")
		}

		return kvService, nil
	}
	if o.Mode == ModeKubernetesSecret {
		kvService, err := kubernetes.NewKVService(o.Kubernetes)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create kv service for kubernetes")
		}

		return kvService, nil
	}

	return nil, errors.New("Invalid mode")
}

func NewVaultClient(hostname string, port string, tlsConfig *vaultapi.TLSConfig) (*vaultapi.Client, error) {
	cfg := vaultapi.DefaultConfig()
	podURL := fmt.Sprintf("https://%s:%s", hostname, port)
	cfg.Address = podURL
	cfg.ConfigureTLS(tlsConfig)
	return vaultapi.NewClient(cfg)
}
