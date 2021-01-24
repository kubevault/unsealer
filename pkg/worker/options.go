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

	aws "kubevault.dev/unsealer/pkg/kv/aws_kms"
	"kubevault.dev/unsealer/pkg/kv/azure"
	google "kubevault.dev/unsealer/pkg/kv/cloudkms"
	"kubevault.dev/unsealer/pkg/kv/kubernetes"
	"kubevault.dev/unsealer/pkg/vault/auth"
	"kubevault.dev/unsealer/pkg/vault/policy"
	"kubevault.dev/unsealer/pkg/vault/unseal"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

const (
	ModeGoogleCloudKmsGCS = "google-cloud-kms-gcs"
	ModeAwsKmsSsm         = "aws-kms-ssm"
	ModeAzureKeyVault     = "azure-key-vault"
	ModeKubernetesSecret  = "kubernetes-secret"

	VaultAddressDefault = "https://127.0.0.1:8200"

	RetryPeriod = 10 * time.Second
)

type WorkerOptions struct {
	// Specifies the vault address
	// Address form : scheme://host:port
	Address string

	// ca cert for vault api client, if vault used a self signed certificate
	CaCert string

	// If InsecureSkipTLSVerify is true, then it will skip tls verification when communicating with vault server
	InsecureSkipTLSVerify bool

	// retry period to try initializing and unsealing
	ReTryPeriod time.Duration

	// Select the mode to use
	// 	- 'google-cloud-kms-gcs' => Google Cloud Storage with encryption using Google KMS
	// 	- 'aws-kms-ssm' => AWS SSM parameter store using AWS KMS encryption
	//  - 'azure-key-vault' => Azure Key Vault Secret store
	//  - 'kubernetes-secret' => Kubernetes secret to store unseal keys
	Mode string

	AuthenticatorOptions *auth.K8sAuthenticatorOptions
	UnsealerOptions      *unseal.UnsealOptions
	PolicyManagerOptions *policy.PolicyManagerOptions
	GoogleOptions        *google.Options
	AwsOptions           *aws.Options
	AzureOptions         *azure.Options
	KubernetesOptions    *kubernetes.Options
}

func NewWorkerOptions() *WorkerOptions {
	return &WorkerOptions{
		Address:              VaultAddressDefault,
		ReTryPeriod:          RetryPeriod,
		UnsealerOptions:      unseal.NewUnsealOptions(),
		AuthenticatorOptions: auth.NewK8sAuthOptions(),
		PolicyManagerOptions: policy.NewPolicyOptions(),
		GoogleOptions:        google.NewOptions(),
		AwsOptions:           aws.NewOptions(),
		AzureOptions:         azure.NewOptions(),
		KubernetesOptions:    kubernetes.NewOptions(),
	}
}

func (o *WorkerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Address, "vault.address", o.Address, "Specifies the vault address. Address form : scheme://host:port")
	fs.StringVar(&o.CaCert, "vault.ca-cert", o.CaCert, "Specifies the CA cert that will be used to verify self signed vault server certificate")
	fs.BoolVar(&o.InsecureSkipTLSVerify, "vault.insecure-skip-tls-verify", o.InsecureSkipTLSVerify, "To skip tls verification when communicating with vault server")
	fs.StringVar(&o.Mode, "mode", o.Mode, "Select the mode to use 'google-cloud-kms-gcs' => Google Cloud Storage with encryption using Google KMS; 'aws-kms-ssm' => AWS SSM parameter store using AWS KMS; 'azure-key-vault' => Azure Key Vault Secret store; 'kubernetes-secret' => Kubernetes secret to store unseal keys")
	fs.DurationVar(&o.ReTryPeriod, "retry-period", o.ReTryPeriod, "How often to attempt to unseal the vault instance")

	o.UnsealerOptions.AddFlags(fs)
	o.AuthenticatorOptions.AddFlags(fs)
	o.PolicyManagerOptions.AddFlags(fs)
	o.GoogleOptions.AddFlags(fs)
	o.AwsOptions.AddFlags(fs)
	o.AzureOptions.AddFlags(fs)
	o.KubernetesOptions.AddFlags(fs)
}

func (o *WorkerOptions) Validate() []error {
	var errs []error
	if o.Mode != ModeGoogleCloudKmsGCS &&
		o.Mode != ModeAwsKmsSsm &&
		o.Mode != ModeKubernetesSecret &&
		o.Mode != ModeAzureKeyVault {
		errs = append(errs, errors.New("invalid mode"))
	}

	errs = append(errs, o.UnsealerOptions.Validate()...)
	errs = append(errs, o.AuthenticatorOptions.Validate()...)
	errs = append(errs, o.PolicyManagerOptions.Validate()...)

	if o.Mode == ModeGoogleCloudKmsGCS {
		errs = append(errs, o.GoogleOptions.Validate()...)
	}
	if o.Mode == ModeAwsKmsSsm {
		errs = append(errs, o.AwsOptions.Validate()...)
	}
	if o.Mode == ModeAzureKeyVault {
		errs = append(errs, o.AzureOptions.Validate()...)
	}
	if o.Mode == ModeKubernetesSecret {
		errs = append(errs, o.KubernetesOptions.Validate()...)
	}

	return errs
}

func (o *WorkerOptions) Apply() error {
	return nil
}
