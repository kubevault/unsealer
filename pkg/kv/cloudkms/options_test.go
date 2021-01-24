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

package cloudkms

import (
	"testing"

	aggregator "github.com/appscode/go/util/errors"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	nonEmpty = "non-empty"
)

func getOptions() *Options {
	return &Options{
		nonEmpty,
		nonEmpty,
		nonEmpty,
		nonEmpty,
		nonEmpty,
		nonEmpty,
	}
}

func getValidationError() []error {
	var errs []error
	errs = append(errs, errors.New("google kms crypto key must be non-empty"))
	errs = append(errs, errors.New("google kms key ring must be non-empty"))
	errs = append(errs, errors.New("google kms location must be non-empty"))
	errs = append(errs, errors.New("google kms project must be non-empty"))
	errs = append(errs, errors.New("google storage bucket name must be non-empty"))
	return errs
}

func TestOptions_Validate(t *testing.T) {
	testData := []struct {
		testName    string
		opts        *Options
		expectedErr error
	}{
		{
			"crypto key is empty, validation failed",
			func() *Options {
				opts := getOptions()
				opts.KmsCryptoKey = ""
				return opts
			}(),
			errors.New("google kms crypto key must be non-empty"),
		},
		{
			"key ring is empty, validation failed",
			func() *Options {
				opts := getOptions()
				opts.KmsKeyRing = ""
				return opts
			}(),
			errors.New("google kms key ring must be non-empty"),
		},
		{
			"location is empty, validation failed",
			func() *Options {
				opts := getOptions()
				opts.KmsLocation = ""
				return opts
			}(),
			errors.New("google kms location must be non-empty"),
		},
		{
			"project is empty, validation failed",
			func() *Options {
				opts := getOptions()
				opts.KmsProject = ""
				return opts
			}(),
			errors.New("google kms project must be non-empty"),
		},
		{
			"bucket is empty, validation failed",
			func() *Options {
				opts := getOptions()
				opts.StorageBucket = ""
				return opts
			}(),
			errors.New("google storage bucket name must be non-empty"),
		},
		{
			"all is non empty, validation successful",
			getOptions(),
			nil,
		},
		{
			"all is empty, validation failed",
			&Options{},
			aggregator.NewAggregate(getValidationError()),
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
