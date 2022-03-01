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

package azure

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

// AzureAuthConfig holds auth related part of cloud config
type AzureAuthConfig struct {
	// The cloud environment identifier. Takes values from https://github.com/Azure/go-autorest/blob/ec5f4903f77ed9927ac95b19ab8e44ada64c1356/autorest/azure/environments.go#L13
	Cloud string `json:"cloud"`
	// The AAD Tenant ID for the Subscription that the cluster is deployed in
	TenantID string `json:"tenantId"`
	// The ClientID for an AAD application with RBAC access to talk to Azure RM APIs
	AADClientID string `json:"aadClientId"`
	// The ClientSecret for an AAD application with RBAC access to talk to Azure RM APIs
	AADClientSecret string `json:"aadClientSecret"`
	// The path of a client certificate for an AAD application with RBAC access to talk to Azure RM APIs
	AADClientCertPath string `json:"aadClientCertPath"`
	// The password of the client certificate for an AAD application with RBAC access to talk to Azure RM APIs
	AADClientCertPassword string `json:"aadClientCertPassword"`
	// Use managed service identity for the virtual machine to access Azure ARM APIs
	UseManagedIdentityExtension bool `json:"useManagedIdentityExtension"`
}

func NewAzureAuthConfig() *AzureAuthConfig {
	return &AzureAuthConfig{
		Cloud:                 "AZUREPUBLICCLOUD",
		AADClientID:           os.Getenv("AZURE_CLIENT_ID"),
		AADClientSecret:       os.Getenv("AZURE_CLIENT_SECRET"),
		AADClientCertPassword: os.Getenv("AZURE_CLIENT_CERT_PASSWORD"),
	}
}

func (o *AzureAuthConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Cloud, "azure.cloud", o.Cloud, "The cloud environment identifier")
	fs.StringVar(&o.AADClientID, "azure.client-id", o.AADClientID, "The ClientID for an AAD application.")
	fs.StringVar(&o.AADClientSecret, "azure.client-secret", o.AADClientSecret, "The ClientSecret for an AAD application")
	fs.StringVar(&o.AADClientCertPath, "azure.client-cert-path", o.AADClientCertPath, "The path of a client certificate for an AAD application")
	fs.StringVar(&o.AADClientCertPassword, "azure.client-cert-password", o.AADClientCertPassword, "The password of the client certificate for an AAD application")
	fs.BoolVar(&o.UseManagedIdentityExtension, "azure.use-managed-identity", o.UseManagedIdentityExtension, "Use managed service identity for the virtual machine")
	fs.StringVar(&o.TenantID, "azure.tenant-id", o.TenantID, "The AAD Tenant ID")
}

func (o *AzureAuthConfig) Validate() []error {
	var errs []error
	if o.Cloud == "" {
		errs = append(errs, errors.New("cloud environment identifier must be non-empty"))
	}
	if o.TenantID == "" {
		errs = append(errs, errors.New("azure tenant id must be non-empty"))
	}
	if (o.AADClientID == "" || o.AADClientSecret == "") &&
		(o.AADClientCertPath == "" || o.AADClientCertPassword == "") &&
		(!o.UseManagedIdentityExtension) {
		errs = append(errs, errors.New("at least one of the following auth mechanism have to use. auth mechanism: (clientID,cleintSecret) or (clientCert,clientCertPassword) or (useManagedIdentity)"))
	}

	return errs
}
