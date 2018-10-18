package policy

import (
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

type PolicyOptions struct {
	Name                    string
	ServiceAccountName      string
	ServiceAccountNamespace string
}

func NewPolicyOptions() *PolicyOptions {
	return &PolicyOptions{}
}

func (o *PolicyOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ServiceAccountName, "policy.service-account-name", o.ServiceAccountName, "Name of the service account")
	fs.StringVar(&o.ServiceAccountNamespace, "policy.service-account-namespace", o.ServiceAccountNamespace, "Namespace of the service account")
	fs.StringVar(&o.Name, "policy.name", o.Name, "Name of the policy. A policy and a kubernetes role will be created using this name")
}

func (o *PolicyOptions) Validate() []error {
	var errs []error
	if o.Name == "" {
		errs = append(errs, errors.New("policy.name must be non empty"))
	}
	if o.ServiceAccountName == "" {
		errs = append(errs, errors.New("policy.service-account-name must be non empty"))
	}
	if o.ServiceAccountNamespace == "" {
		errs = append(errs, errors.New("policy.service-account-namespace must be non empty"))
	}
	return errs
}

func (o *PolicyOptions) Apply() error {
	return nil
}
