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
