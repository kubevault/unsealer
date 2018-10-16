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
		opts        *K8sAuthOptions
		expectedErr error
	}{
		{
			"host is empty, validation failed",
			&K8sAuthOptions{},
			errors.New("auth.k8s-host must be non empty"),
		},
		{
			"validation successful",
			&K8sAuthOptions{
				Host: "hi.com",
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
