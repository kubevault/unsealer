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
package aws_ssm

import (
	"encoding/base64"
	"fmt"

	"kubevault.dev/unsealer/pkg/kv"
	"kubevault.dev/unsealer/pkg/kv/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/pkg/errors"
)

type awsSSM struct {
	ssmService *ssm.SSM

	useSecureString bool
	keyPrefix       string
}

var _ kv.Service = &awsSSM{}

func NewWithSession(sess *session.Session, useSecureString bool, keyPrefix string) (*awsSSM, error) {
	var ssmService *ssm.SSM

	region := util.GetAWSRegion()
	if region != "" {
		ssmService = ssm.New(sess, aws.NewConfig().WithRegion(region))
	} else {
		ssmService = ssm.New(sess)
	}

	return &awsSSM{
		ssmService:      ssmService,
		keyPrefix:       keyPrefix,
		useSecureString: useSecureString,
	}, nil
}

func New(useSecureString bool, keyPrefix string) (*awsSSM, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	return NewWithSession(sess, useSecureString, keyPrefix)
}

func (a *awsSSM) Get(key string) ([]byte, error) {
	req := &ssm.GetParametersInput{
		Names: []*string{
			aws.String(a.name(key)),
		},
		WithDecryption: aws.Bool(a.useSecureString),
	}

	out, err := a.ssmService.GetParameters(req)
	if err != nil {
		return []byte{}, err
	}

	if len(out.Parameters) < 1 {
		return []byte{}, kv.NewNotFoundError("key '%s' not found", key)
	}

	return base64.StdEncoding.DecodeString(*out.Parameters[0].Value)
}

func (a *awsSSM) name(key string) string {
	return fmt.Sprintf("%s%s", a.keyPrefix, key)
}

func (a *awsSSM) Set(key string, val []byte) error {
	req := &ssm.PutParameterInput{
		Description: aws.String("vault-unsealer"),
		Name:        aws.String(a.name(key)),
		Overwrite:   aws.Bool(true),
		Value:       aws.String(base64.StdEncoding.EncodeToString(val)),
	}

	if a.useSecureString {
		req.Type = aws.String("SecureString")
	} else {
		req.Type = aws.String("String")
	}

	_, err := a.ssmService.PutParameter(req)
	return err
}

func (a *awsSSM) CheckWriteAccess() error {
	key := "vault-unsealer-dummy-file"
	val := "read write access check"

	err := a.Set(key, []byte(val))
	if err != nil {
		return errors.Wrap(err, "failed to write test file")
	}

	_, err = a.Get(key)
	if err != nil {
		return errors.Wrap(err, "failed to get test file")
	}

	_, err = a.ssmService.DeleteParameter(&ssm.DeleteParameterInput{
		Name: aws.String(a.name(key)),
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete test file")
	}

	return nil
}

func (g *awsSSM) Test(key string) error {
	// TODO: Implement a test if a Set is likely to work, AWS doesn't seemt to provide a dry-run on the parameter store
	return nil
}
