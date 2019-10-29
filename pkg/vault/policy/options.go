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
