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
			&K8sAuthOptions{
				Jwt: "ok",
			},
			errors.New("auth.k8s-host must be non empty"),
		},
		{
			"env K8S_TOKEN_REVIEWER_JWT is empty, validation failed",
			&K8sAuthOptions{
				Host: "hi.com",
			},
			errors.New("env K8S_TOKEN_REVIEWER_JWT must be non empty"),
		},
		{
			"validation successful",
			&K8sAuthOptions{
				Host: "hi.com",
				Jwt:  "ok",
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
