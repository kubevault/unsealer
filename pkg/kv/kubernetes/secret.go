package kubernetes

import (
	"fmt"

	kutilpatch "github.com/appscode/kutil/core/v1"
	kutilmeta "github.com/appscode/kutil/meta"
	"github.com/kubevault/unsealer/pkg/kv"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KVService struct {
	KubeClient kubernetes.Interface
	SecretName string
	Namespace  string
}

func NewKVService(c *Options) (*KVService, error) {
	k := &KVService{
		SecretName: c.SecretName,
		Namespace:  kutilmeta.Namespace(),
	}

	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create in cluster config")
	}

	k.KubeClient, err = kubernetes.NewForConfig(kubeConfig)
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
	_, _, err := kutilpatch.CreateOrPatchSecret(k.KubeClient, secretMeta, func(s *corev1.Secret) *corev1.Secret {
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
		return nil, kv.NewNotFoundError(fmt.Sprintf("failed to get secret. reason: %v", err))
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

func (k *KVService) Test(key string) error {
	return nil
}
