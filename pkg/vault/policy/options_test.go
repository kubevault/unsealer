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

package policy

import (
	"testing"

	aggregator "github.com/appscode/go/util/errors"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestVaultOptions_Validate(t *testing.T) {
	testData := []struct {
		testName    string
		opts        *PolicyManagerOptions
		expectedErr error
	}{
		{
			"validation successful",
			&PolicyManagerOptions{
				Name:                    "ok",
				ServiceAccountName:      "ok",
				ServiceAccountNamespace: "ok",
			},
			nil,
		},
		{
			"name is empty, validation failed",
			&PolicyManagerOptions{
				Name:                    "",
				ServiceAccountName:      "ok",
				ServiceAccountNamespace: "ok",
			},
			errors.New("policy-manager.name must be non empty"),
		},
		{
			"service account name is empty, validation failed",
			&PolicyManagerOptions{
				Name:                    "ok",
				ServiceAccountName:      "",
				ServiceAccountNamespace: "ok",
			},
			errors.New("policy-manager.service-account-name must be non empty"),
		},
		{
			"service account namespace is empty, validation failed",
			&PolicyManagerOptions{
				Name:                    "ok",
				ServiceAccountName:      "ok",
				ServiceAccountNamespace: "",
			},
			errors.New("policy-manager.service-account-namespace must be non empty"),
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
