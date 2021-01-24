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

package vault

import (
	"crypto/x509"
	"net/http"
	"net/url"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

// if caCert is empty, then TLS verification will be skipped
func NewVaultClient(addr string, insecureSkipVerify bool, caCert []byte) (*vaultapi.Client, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr

	clientTLSConfig := cfg.HttpClient.Transport.(*http.Transport).TLSClientConfig
	if insecureSkipVerify {
		clientTLSConfig.InsecureSkipVerify = true
	} else {
		if len(caCert) != 0 {
			pool := x509.NewCertPool()
			ok := pool.AppendCertsFromPEM(caCert)
			if !ok {
				return nil, errors.New("error loading CA File: couldn't parse PEM data in CA bundle")
			}
			clientTLSConfig.RootCAs = pool
		} else {
			clientTLSConfig.InsecureSkipVerify = true
		}
	}

	var err error
	clientTLSConfig.ServerName, err = getHostName(cfg.Address)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get hostname from url %s", cfg.Address)
	}

	return vaultapi.NewClient(cfg)
}

func getHostName(addr string) (string, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return u.Hostname(), nil
}
