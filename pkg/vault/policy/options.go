package policy

import (
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

type PolicyManagerOptions struct {
	Name                    string
	ServiceAccountName      string
	ServiceAccountNamespace string
}

func NewPolicyOptions() *PolicyManagerOptions {
	return &PolicyManagerOptions{}
}

func (o *PolicyManagerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ServiceAccountName, "policy-manager.service-account-name", o.ServiceAccountName, "Name of the service account")
	fs.StringVar(&o.ServiceAccountNamespace, "policy-manager.service-account-namespace", o.ServiceAccountNamespace, "Namespace of the service account")
	fs.StringVar(&o.Name, "policy-manager.name", o.Name, "Name of the policy. A policy and a  vault kubernetes auth role will be created using this name")
}

func (o *PolicyManagerOptions) Validate() []error {
	var errs []error
	if o.Name == "" {
		errs = append(errs, errors.New("policy-manager.name must be non empty"))
	}
	if o.ServiceAccountName == "" {
		errs = append(errs, errors.New("policy-manager.service-account-name must be non empty"))
	}
	if o.ServiceAccountNamespace == "" {
		errs = append(errs, errors.New("policy-manager.service-account-namespace must be non empty"))
	}
	return errs
}

func (o *PolicyManagerOptions) Apply() error {
	return nil
}
