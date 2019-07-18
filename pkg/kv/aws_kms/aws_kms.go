package aws_kms

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"kubevault.dev/unsealer/pkg/kv"
	"kubevault.dev/unsealer/pkg/kv/util"
)

type awsKMS struct {
	store      kv.Service
	kmsService *kms.KMS

	kmsID string
}

var _ kv.Service = &awsKMS{}

func NewWithSession(sess *session.Session, store kv.Service, kmsID string) (kv.Service, error) {
	if kmsID == "" {
		return nil, fmt.Errorf("invalid kmsID specified: '%s'", kmsID)
	}

	var ksmService *kms.KMS

	region := util.GetAWSRegion()
	if region != "" {
		ksmService = kms.New(sess, aws.NewConfig().WithRegion(region))
	} else {
		ksmService = kms.New(sess)
	}

	return &awsKMS{
		store:      store,
		kmsService: ksmService,
		kmsID:      kmsID,
	}, nil
}

func New(store kv.Service, kmsID string) (kv.Service, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	return NewWithSession(sess, store, kmsID)
}

func (a *awsKMS) decrypt(cipherText []byte) ([]byte, error) {
	out, err := a.kmsService.Decrypt(&kms.DecryptInput{
		CiphertextBlob: cipherText,
		EncryptionContext: map[string]*string{
			"Tool": aws.String("vault-unsealer"),
		},
		GrantTokens: []*string{},
	})
	return out.Plaintext, err
}

func (a *awsKMS) Get(key string) ([]byte, error) {
	cipherText, err := a.store.Get(key)
	if err != nil {
		return nil, err
	}

	return a.decrypt(cipherText)
}

func (a *awsKMS) encrypt(plainText []byte) ([]byte, error) {

	out, err := a.kmsService.Encrypt(&kms.EncryptInput{
		KeyId:     aws.String(a.kmsID),
		Plaintext: plainText,
		EncryptionContext: map[string]*string{
			"Tool": aws.String("vault-unsealer"),
		},
		GrantTokens: []*string{},
	})
	return out.CiphertextBlob, err
}

func (a *awsKMS) Set(key string, val []byte) error {
	cipherText, err := a.encrypt(val)

	if err != nil {
		return err
	}

	return a.store.Set(key, cipherText)
}

func (a *awsKMS) CheckWriteAccess() error {
	return a.store.CheckWriteAccess()
}

func (g *awsKMS) Test(key string) error {
	inputString := "test"

	err := g.store.Test(key)
	if err != nil {
		return fmt.Errorf("test of backend store failed: %s", err.Error())
	}

	cipherText, err := g.encrypt([]byte(inputString))
	if err != nil {
		return err
	}

	plainText, err := g.decrypt(cipherText)
	if err != nil {
		return err
	}

	if string(plainText) != inputString {
		return fmt.Errorf("encryped and decryped text doesn't match: exp: '%v', act: '%v'", inputString, string(plainText))
	}

	return nil
}
