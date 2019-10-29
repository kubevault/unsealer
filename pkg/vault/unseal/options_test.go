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
package unseal

import (
	"testing"

	aggregator "github.com/appscode/go/util/errors"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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
			errors.New("secret threshold must be positive"),
		},
		{
			"secret threshold > secret shares, validation failed",
			&UnsealOptions{
				SecretShares:    1,
				SecretThreshold: 2,
			},
			errors.New("secret threshold must be less than or equal to secret shares"),
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
