package kubernetes

import (
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

type Options struct {
	SecretName string
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.SecretName, "k8s.secret-name", o.SecretName, "Secret name to use when creating secret containing root token and shared keys")
}

func (o *Options) Validate() []error {
	var errs []error
	if o.SecretName == "" {
		errs = append(errs, errors.New("secret name must be non-empty"))
	}
	return errs
}

func (o *Options) Apply() error {
	return nil
}
