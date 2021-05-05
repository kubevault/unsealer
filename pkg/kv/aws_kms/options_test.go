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

package aws_kms

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	aggregator "gomodules.xyz/errors"
)

func getValidationErrorForFlagsNotProvided() []error {
	var errs []error
	errs = append(errs, errors.New("--aws.kms-key-id or --aws.use-secure-string must be defined"))
	return errs
}

func getValidationErrorForBothFlagProvided() []error {
	var errs []error
	errs = append(errs, errors.New("--aws.kms-key-id and --aws.use-secure-string both are defined, but only one of them is needed"))
	return errs
}

func TestOptions_Validate(t *testing.T) {
	testData := []struct {
		testName    string
		opts        *Options
		expectedErr error
	}{
		{
			"aws key id is provided, validation successful",
			&Options{
				"test-key",
				false,
				"",
			},
			nil,
		},
		{
			"aws key and useSecureString is provided, error expected",
			&Options{
				"test-key",
				true,
				"",
			},
			aggregator.NewAggregate(getValidationErrorForBothFlagProvided()),
		},
		{
			"useSecureString is provided",
			&Options{
				"",
				true,
				"",
			},
			nil,
		},
		{
			"aws key id and use secure string not provided, validation failed",
			&Options{
				"",
				false,
				"",
			},
			aggregator.NewAggregate(getValidationErrorForFlagsNotProvided()),
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
