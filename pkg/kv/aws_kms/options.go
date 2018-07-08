package aws_kms

import (
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

type Options struct {
	// The ID or ARN of the AWS KMS key to encrypt values
	KmsKeyID string

	// Use secure string parameter
	// More info: https://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-paramstore-about.html#sysman-paramstore-securestring
	UseSecureString bool

	// TODO: should make it auto generated
	SsmKeyPrefix string
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.KmsKeyID, "aws.kms-key-id", o.KmsKeyID, "The ID or ARN of the AWS KMS key to encrypt values")
	fs.StringVar(&o.SsmKeyPrefix, "aws.ssm-key-prefix", o.SsmKeyPrefix, "The Key Prefix for SSM Parameter store")
	fs.BoolVar(&o.UseSecureString, "aws.use-secure-string", o.UseSecureString, "Use secure string parameter, for more info https://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-paramstore-about.html#sysman-paramstore-securestring")
}

func (o *Options) Validate() []error {
	var errs []error
	if o.KmsKeyID == "" && !o.UseSecureString {
		errs = append(errs, errors.New("--aws.kms-key-id or --aws.use-secure-string must be defined"))
	}
	if o.KmsKeyID != "" && o.UseSecureString {
		errs = append(errs, errors.New("--aws.kms-key-id and --aws.use-secure-string both are defined, but only one of them is needed"))
	}

	return errs
}

func (o *Options) Apply() error {
	return nil
}
