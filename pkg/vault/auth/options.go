package auth

import (
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

type K8sAuthOptions struct {
	// Host must be a host string, a host:port pair
	// or a URL to the base of the Kubernetes API server.
	Host string

	// PEM encoded CA cert for use by the TLS client used
	// to talk with the Kubernetes API
	CA string

	// A service account JWT used to access the TokenReview API
	// to validate other JWTs during login. If not set the JWT
	// used for login will be used to access the API.
	Jwt string
}

func NewK8sAuthOptions() *K8sAuthOptions {
	return &K8sAuthOptions{}
}

func (o *K8sAuthOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Host, "auth.k8s-host", o.Host, "Host must be a host string, a host:port pair, or a URL to the base of the Kubernetes API server")
	fs.StringVar(&o.CA, "auth.k8s-ca-cert ", o.CA, "PEM encoded CA cert for use by the TLS client used to talk with the Kubernetes API")
	fs.StringVar(&o.Jwt, "auth.k8s-token-reviewer-jwt ", o.Jwt, "A service account JWT used to access the TokenReview API to validate other JWTs during login")
}

func (o *K8sAuthOptions) Validate() []error {
	var errs []error
	if o.Host == "" {
		errs = append(errs, errors.New("auth.k8s-host must be non empty"))
	}
	return errs
}

func (o *K8sAuthOptions) Apply() error {
	return nil
}