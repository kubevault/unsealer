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
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	aggregator "gomodules.xyz/errors"
)

func TestVaultOptions_Validate(t *testing.T) {
	testData := []struct {
		testName    string
		opts        *UnsealOptions
		expectedErr error
	}{
		{
			"secret threshold is zero, validation failed",
			&UnsealOptions{
				SecretShares:    1,
				SecretThreshold: 0,
			},
			errors.New("secret threshold must be positive, cluster-name flag not set"),
		},
		{
			"secret threshold > secret shares, validation failed",
			&UnsealOptions{
				SecretShares:    1,
				SecretThreshold: 2,
			},
			errors.New("secret threshold must be less than or equal to secret shares, cluster-name flag not set"),
		},
		{
			"validation successful",
			&UnsealOptions{
				SecretShares:    10,
				SecretThreshold: 2,
			},
			nil,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			errs := test.opts.Validate()
			if test.expectedErr != nil {
				assert.EqualError(t, aggregator.NewAggregate(errs), test.expectedErr.Error())
			} else {
				assert.Nil(t, errs)
			}
		})
	}
}
