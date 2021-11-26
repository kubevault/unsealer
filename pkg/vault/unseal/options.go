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

package unseal

import (
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

// That configures the vault API
type UnsealOptions struct {
	KeyPrefix string

	// how many key parts exist
	SecretShares int
	// how many of these parts are needed to unseal vault  (secretThreshold <= secretShares)
	SecretThreshold int

	// should the root token be stored in the keyStore
	StoreRootToken bool

	// overwrite existing tokens
	OverwriteExisting bool
}

func NewUnsealOptions() *UnsealOptions {
	return &UnsealOptions{
		KeyPrefix:       "vault",
		SecretThreshold: 3,
		SecretShares:    5,
		StoreRootToken:  true,
	}
}

func (o *UnsealOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.StoreRootToken, "store-root-token", o.StoreRootToken, "should the root token be stored in the key store")
	fs.BoolVar(&o.OverwriteExisting, "overwrite-existing", o.OverwriteExisting, "overwrite existing unseal keys and root tokens, possibly dangerous!")
	fs.IntVar(&o.SecretShares, "secret-shares", o.SecretShares, "Total count of secret shares that exist")
	fs.IntVar(&o.SecretThreshold, "secret-threshold", o.SecretThreshold, "Minimum required secret shares to unseal")
	fs.StringVar(&o.KeyPrefix, "key-prefix", o.KeyPrefix, "root token and unseal key prefix")
}

func (o *UnsealOptions) Validate() []error {
	var errs []error
	if o.SecretThreshold <= 0 {
		errs = append(errs, errors.New("secret threshold must be positive"))
	}
	if o.SecretShares <= 0 {
		errs = append(errs, errors.New("secret shares must be positive"))
	}
	if o.SecretThreshold > o.SecretShares {
		errs = append(errs, errors.New("secret threshold must be less than or equal to secret shares"))
	}
	return errs
}

func (o *UnsealOptions) Apply() error {
	return nil
}
