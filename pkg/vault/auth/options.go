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

package auth

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

const (
	K8sTokenReviewerJwtEnv = "K8S_TOKEN_REVIEWER_JWT"
)

type K8sAuthenticatorOptions struct {
	// Host must be a host string, a host:port pair
	// or a URL to the base of the Kubernetes API server.
	Host string

	// PEM encoded CA cert for use by the TLS client used
	// to talk with the Kubernetes API
	CA string

	// A service account JWT used to access the TokenReview API
	// to validate other JWTs during login. If not set the JWT
	// used for login will be used to access the API.
	Token string
}

func NewK8sAuthOptions() *K8sAuthenticatorOptions {
	return &K8sAuthenticatorOptions{
		Token: os.Getenv(K8sTokenReviewerJwtEnv),
	}
}

func (o *K8sAuthenticatorOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Host, "auth.k8s-host", o.Host, "Host must be a host string, a host:port pair, or a URL to the base of the Kubernetes API server")
	fs.StringVar(&o.CA, "auth.k8s-ca-cert", o.CA, "PEM encoded CA cert for use by the TLS client used to talk with the Kubernetes API")
	fs.StringVar(&o.Token, "auth.k8s-token-reviewer-jwt", o.Token, "A service account JWT used to access the TokenReview API to validate other JWTs during login. If this flag is not provided, then the value from K8S_TOKEN_REVIEWER_JWT environment variable will be used")
}

func (o *K8sAuthenticatorOptions) Validate() []error {
	var errs []error
	if o.Host == "" {
		errs = append(errs, errors.New("auth.k8s-host must be non empty"))
	}
	if o.Token == "" {
		errs = append(errs, errors.New("env K8S_TOKEN_REVIEWER_JWT must be non empty"))
	}
	return errs
}

func (o *K8sAuthenticatorOptions) Apply() error {
	return nil
}
