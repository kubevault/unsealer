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
	"kubevault.dev/unsealer/pkg/worker/dcdr"

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

	// KubeSlice slice name for the per-pod ServiceExport (distributed mode)
	SliceName string

	// File holding the token the dc-dr guard uses for sys/step-down (distributed mode)
	DCDRGuardTokenFile string

	// Directory of the mounted ConfigMap the active node bootstraps Kubernetes
	// auth from (distributed mode, cloud KMS unseal modes)
	DCDRK8sAuthDir string

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
		DCDRGuardTokenFile:   dcdr.DCDRGuardTokenFileDefault,
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
	fs.StringVar(&o.SliceName, "slice-name", o.SliceName, "KubeSlice slice name for the per-pod ServiceExport. Used in cross data center mode (env DC_NAME set)")
	fs.StringVar(&o.DCDRGuardTokenFile, "dc-dr-guard-token-file", o.DCDRGuardTokenFile, "File holding the Vault token the cross data center guard uses to step down its own node")
	fs.StringVar(&o.DCDRK8sAuthDir, "dc-dr-k8s-auth-dir", o.DCDRK8sAuthDir, "Directory of the mounted ConfigMap the active node bootstraps Vault's Kubernetes auth from, in cross data center mode with a cloud KMS unseal mode")

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
	// Kubernetes auth is configured by the unsealer only when the operator
	// passes --auth.k8s-host (older operators); newer operators configure it.
	if o.configureEnabled() {
		errs = append(errs, o.AuthenticatorOptions.Validate()...)
		errs = append(errs, o.PolicyManagerOptions.Validate()...)
	}

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

// configureEnabled reports whether the unsealer configures Kubernetes auth and
// the policy controller's policy and role after unsealing.
func (o *WorkerOptions) configureEnabled() bool {
	return o.AuthenticatorOptions.Host != ""
}

func (o *WorkerOptions) Apply() error {
	return nil
}
