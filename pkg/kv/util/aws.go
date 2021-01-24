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

package util

import (
	"net/http"
	"time"

	"github.com/appscode/go/net/httpclient"
)

// http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-identity-documents.html
func GetAWSRegion() string {
	md := struct {
		PrivateIP        string    `json:"privateIp"`
		AvailabilityZone string    `json:"availabilityZone"`
		AccountID        string    `json:"accountId"`
		Version          string    `json:"version"`
		InstanceID       string    `json:"instanceId"`
		InstanceType     string    `json:"instanceType"`
		ImageID          string    `json:"imageId"`
		PendingTime      time.Time `json:"pendingTime"`
		Architecture     string    `json:"architecture"`
		Region           string    `json:"region"`
	}{}

	hc := httpclient.Default()
	resp, err := hc.Call(http.MethodGet, "http://169.254.169.254/latest/dynamic/instance-identity/document", nil, &md, false)
	if err == nil &&
		resp.StatusCode == http.StatusOK {
		return md.Region
	}

	return ""
}
