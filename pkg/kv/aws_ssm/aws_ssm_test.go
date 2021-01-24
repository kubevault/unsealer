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

package aws_ssm

import (
	"os"
	"testing"

	"kubevault.dev/unsealer/pkg/kv"
)

func TestAWSIntegration(t *testing.T) {
	region := os.Getenv("AWS_REGION")

	if region == "" {
		t.Skip("Skip AWS integration tests: not environment variable 'AWS_REGION' specified")
	}

	payloadKey := "test123"
	payloadValue := "payload123"

	a, err := New(false, "test-integration-")
	if err != nil {
		t.Errorf("Unexpected error creating SSM kv: %s", err)
	}

	// graceful set (in case it's already existing)
	err = a.Set(payloadKey, []byte(payloadValue))
	if err != nil {
		t.Errorf("Unexpected error storing value in SSM kv: %s", err)
	}

	// this should also work and overwrite a key
	err = a.Set(payloadKey, []byte(payloadValue))
	if err != nil {
		t.Errorf("Unexpected error storing value in SSM kv: %s", err)
	}

	_, err = a.Get(payloadKey)
	if _, ok := err.(*kv.NotFoundError); !ok {
		t.Errorf("Expected an kv.NotFoundError for a non existing key")
	}

	err = a.Set(payloadKey, []byte(payloadValue))
	if err != nil {
		t.Errorf("Unexpected error storing value in SSM kv: %s", err)
	}

	out, err := a.Get("test123")
	if err != nil {
		t.Errorf("Unexpected error storing value in SSM kv: %s", err)
	}

	if exp, act := payloadValue, string(out); exp != act {
		t.Errorf("Unexpected decrypt output: exp=%s act=%s", exp, act)
	}

}
