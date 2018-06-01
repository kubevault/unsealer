package azure

import (
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

type Options struct {
	VaultBaseUrl string
	AuthConfig   *AzureAuthConfig
	// TODO: should make it auto generated
	SecretPrefix string // prefix to use in secret name for azure key vault
}

func NewOptions() *Options {
	return &Options{
		AuthConfig: NewAzureAuthConfig(),
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.VaultBaseUrl, "azure.vault-base-url", o.VaultBaseUrl, "Azure key vault url, for example https://myvault.vault.azure.net")
	fs.StringVar(&o.SecretPrefix, "azure.secret-prefix", o.VaultBaseUrl, "Prefix to use in secret name for azure key vault")

	o.AuthConfig.AddFlags(fs)
}

func (o *Options) Validate() []error {
	var errs []error
	if o.VaultBaseUrl == "" {
		errs = append(errs, errors.New("azure key vault url must be non-empty"))
	}

	errs = append(errs, o.AuthConfig.Validate()...)
	return errs
}

func (o *Options) Apply() error {
	return nil
}
