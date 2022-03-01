/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package unseal

import (
	"fmt"
	"testing"

	"kubevault.dev/unsealer/pkg/kv"
)

type fakeKV struct {
	Values map[string]*[]byte
}

func NewFakeKV() *fakeKV {
	return &fakeKV{
		Values: map[string]*[]byte{},
	}
}

func (f *fakeKV) Test(key string) error {
	return fmt.Errorf("not-implemented")
}

func (f *fakeKV) CheckWriteAccess() error {
	return fmt.Errorf("not-implemented")
}

func (f *fakeKV) Set(key string, data []byte) error {
	return fmt.Errorf("not-implemented")
}

func (f *fakeKV) Get(key string) ([]byte, error) {
	if key == "exists" {
		return []byte("data"), nil
	} else if key == "not-found" {
		return nil, kv.NewNotFoundError("not-found")
	}

	return nil, fmt.Errorf("not-implemented")
}

func TestKeyStoreNotFound(t *testing.T) {
	fakeKV := NewFakeKV()
	v := &unsealer{
		keyStore: fakeKV,
	}

	if !v.keyStoreNotFound("not-found") {
		t.Error("not returning true for notfound")
	}

	if v.keyStoreNotFound("exists") {
		t.Error("not returing false for existing")
	}

	if v.keyStoreNotFound("error") {
		t.Error("not returning false for error case")
	}
}
