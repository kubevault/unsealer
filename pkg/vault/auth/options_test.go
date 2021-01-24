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

package auth

import (
	"testing"

	aggregator "github.com/appscode/go/util/errors"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestVaultOptions_Validate(t *testing.T) {
	testData := []struct {
		testName    string
		opts        *K8sAuthenticatorOptions
		expectedErr error
	}{
		{
			"host is empty, validation failed",
			&K8sAuthenticatorOptions{
				Token: "ok",
			},
			errors.New("auth.k8s-host must be non empty"),
		},
		{
			"env K8S_TOKEN_REVIEWER_JWT is empty, validation failed",
			&K8sAuthenticatorOptions{
				Host: "hi.com",
			},
			errors.New("env K8S_TOKEN_REVIEWER_JWT must be non empty"),
		},
		{
			"validation successful",
			&K8sAuthenticatorOptions{
				Host:  "hi.com",
				Token: "ok",
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
