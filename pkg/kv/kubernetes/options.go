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
