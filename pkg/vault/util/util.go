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
package util

import "fmt"

// UnsealKeyID is the ID that used as key name when storing unseal key
func UnsealKeyID(prefix string, i int) string {
	return fmt.Sprintf("%s-unseal-key-%d", prefix, i)
}

// RootTokenID is the ID that used as key name when storing root token
func RootTokenID(prefix string) string {
	return fmt.Sprintf("%s-root-token", prefix)
}
