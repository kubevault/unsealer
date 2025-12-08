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

package kv

import "fmt"

type NotFoundError struct {
	msg string // description of error
}

func (e *NotFoundError) Error() string { return e.msg }

func NewNotFoundError(msg string, args ...any) *NotFoundError {
	return &NotFoundError{
		msg: fmt.Sprintf(msg, args...),
	}
}

// Service defines a basic key-value store. Implementations of this interface
// may or may not guarantee consistency or security properties.
type Service interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	CheckWriteAccess() error
	Test(key string) error
}
