package kubernetes

import (
	"fmt"

	"kubevault.dev/unsealer/pkg/kv"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/client-go/tools/clientcmd"
)

type KVService struct {
	KubeClient kubernetes.Interface
	SecretName string
	Namespace  string
}

func NewKVService(c *Options) (*KVService, error) {
	k := &KVService{
		SecretName: c.SecretName,
		Namespace:  meta_util.Namespace(),
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create in cluster config")
	}
	clientcmd.Fix(config)

	k.KubeClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kubernetes clientset")
	}

	return k, nil
}

func (k *KVService) Set(key string, value []byte) error {
	secretMeta := metav1.ObjectMeta{
		Name:      k.SecretName,
		Namespace: k.Namespace,
	}
	_, _, err := core_util.CreateOrPatchSecret(k.KubeClient, secretMeta, func(s *corev1.Secret) *corev1.Secret {
		if s.Data == nil {
			s.Data = map[string][]byte{}
		}

		s.Data[key] = value
		return s
	})
	if err != nil {
		return errors.Wrapf(err, "failed set data in secret(%s)", k.SecretName)
	}

	return nil
}

func (k *KVService) Get(key string) ([]byte, error) {
	sr, err := k.KubeClient.CoreV1().Secrets(k.Namespace).Get(k.SecretName, metav1.GetOptions{})
	if kerror.IsNotFound(err) {
		return nil, kv.NewNotFoundError(fmt.Sprintf("secret not found. reason: %v", err))
	} else if err != nil {
		return nil, fmt.Errorf("failed to get secret. reason: %v", err)
	}

	if sr.Data == nil {
		return nil, kv.NewNotFoundError(fmt.Sprintf("key not found in secret data. reason: %v", err))
	}

	if value, ok := sr.Data[key]; ok {
		return value, nil
	} else {
		return nil, kv.NewNotFoundError("key not found in secret data")
	}
}

func (k *KVService) CheckWriteAccess() error {
	key := "vault-unsealer-dummy-file"
	val := "read write access check"

	err := k.Set(key, []byte(val))
	if err != nil {
		return errors.Wrap(err, "failed to write test data")
	}

	_, err = k.Get(key)
	if err != nil {
		return errors.Wrap(err, "failed to get test data")
	}

	sr, err := k.KubeClient.CoreV1().Secrets(k.Namespace).Get(k.SecretName, metav1.GetOptions{})
	if kerror.IsNotFound(err) {
		return kv.NewNotFoundError(fmt.Sprintf("secret not found. reason: %v", err))
	} else if err != nil {
		return fmt.Errorf("failed to get secret. reason: %v", err)
	}

	newData := map[string][]byte{}

	for k, v := range sr.Data {
		if k != key {
			newData[k] = v
		}
	}

	_, _, err = core_util.CreateOrPatchSecret(k.KubeClient, sr.ObjectMeta, func(s *corev1.Secret) *corev1.Secret {
		s.Data = newData
		return s
	})
	if err != nil {
		return errors.Wrapf(err, "failed delete data in secret(%s)", k.SecretName)
	}

	return nil
}

func (k *KVService) Test(key string) error {
	return nil
}
