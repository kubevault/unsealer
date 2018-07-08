package aws_kms

import (
	"testing"

	aggregator "github.com/appscode/go/util/errors"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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
