package gcs

import (
	"context"
	"fmt"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"github.com/kubevault/unsealer/pkg/kv"
	"github.com/pkg/errors"
)

type gcsStorage struct {
	cl     *storage.Client
	bucket string
	prefix string
}

func New(bucket, prefix string) (kv.Service, error) {
	cl, err := storage.NewClient(context.Background())

	if err != nil {
		return nil, fmt.Errorf("error creating gcs client: %s", err.Error())
	}

	return &gcsStorage{cl, bucket, prefix}, nil
}

func (g *gcsStorage) Set(key string, val []byte) error {
	ctx := context.Background()
	n := objectNameWithPrefix(g.prefix, key)
	w := g.cl.Bucket(g.bucket).Object(n).NewWriter(ctx)
	if _, err := w.Write(val); err != nil {
		return fmt.Errorf("error writing key '%s' to gcs bucket '%s'", n, g.bucket)
	}

	return w.Close()
}

func (g *gcsStorage) Get(key string) ([]byte, error) {
	ctx := context.Background()
	n := objectNameWithPrefix(g.prefix, key)

	r, err := g.cl.Bucket(g.bucket).Object(n).NewReader(ctx)

	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, kv.NewNotFoundError("error getting object for key '%s': %s", n, err.Error())
		}
		return nil, fmt.Errorf("error getting object for key '%s': %s", n, err.Error())
	}

	b, err := ioutil.ReadAll(r)

	if err != nil {
		return nil, fmt.Errorf("error reading object with key '%s': %s", n, err.Error())
	}

	return b, nil
}

func (g *gcsStorage) CheckWriteAccess() error {
	key := "vault-unsealer-dummy-file"
	val := "read write access check"

	err := g.Set(key, []byte(val))
	if err != nil {
		return errors.Wrap(err, "failed to write test file")
	}

	_, err = g.Get(key)
	if err != nil {
		return errors.Wrap(err, "failed to get test file")
	}

	ctx := context.Background()

	err = g.cl.Bucket(g.bucket).Object(key).Delete(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to delete test file")
	}
	return nil
}

func objectNameWithPrefix(prefix, key string) string {
	return fmt.Sprintf("%s%s", prefix, key)
}

func (g *gcsStorage) Test(key string) error {
	// TODO: Implement me properly
	return nil
}
