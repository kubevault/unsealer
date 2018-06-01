---
title: Vault-Unsealer Run
menu:
  product_unsealer_0.0.1:
    identifier: vault-unsealer-run
    name: Vault-Unsealer Run
    parent: reference
product_name: unsealer
menu_name: product_unsealer_0.0.1
section_menu_id: reference
---
## vault-unsealer run

Launch Vault unsealer

### Synopsis

Launch Vault unsealer

```
vault-unsealer run [flags]
```

### Options

```
      --aws.kms-key-id string               The ID or ARN of the AWS KMS key to encrypt values
      --aws.ssm-key-prefix string           The Key Prefix for SSM Parameter store
      --azure.client-cert-password string   The password of the client certificate for an AAD application
      --azure.client-cert-path string       The path of a client certificate for an AAD application
      --azure.client-id string              The ClientID for an AAD application.
      --azure.client-secret string          The ClientSecret for an AAD application
      --azure.cloud string                  The cloud environment identifier (default "AZUREPUBLICCLOUD")
      --azure.secret-prefix string          Prefix to use in secret name for azure key vault
      --azure.tenant-id string              The AAD Tenant ID
      --azure.use-managed-identity          Use managed service identity for the virtual machine
      --azure.vault-base-url string         Azure key vault url, for example https://myvault.vault.azure.net
      --ca-cert-file string                 Path to the ca cert file that will be used to verify self signed vault server certificate
      --google.kms-crypto-key string        The name of the Google Cloud KMS crypto key to use
      --google.kms-key-ring string          The name of the Google Cloud KMS key ring to use
      --google.kms-location string          The Google Cloud KMS location to use (eg. 'global', 'europe-west1')
      --google.kms-project string           The Google Cloud KMS project to use
      --google.storage-bucket string        The name of the Google Cloud Storage bucket to store values in
      --google.storage-prefix string        The prefix to use for values store in Google Cloud Storage
  -h, --help                                help for run
      --insecure-tls                        To skip tls verification when communicating with vault server
      --k8s.secret-name string              Secret name to use when creating secret containing root token and shared keys
      --mode string                         Select the mode to use 'google-cloud-kms-gcs' => Google Cloud Storage with encryption using Google KMS; 'aws-kms-ssm' => AWS SSM parameter store using AWS KMS; 'azure-key-vault' => Azure Key Vault Secret store; 'kubernetes-secret' => Kubernetes secret to store unseal keys
      --overwrite-existing                  overwrite existing unseal keys and root tokens, possibly dangerous!
      --retry-period duration               How often to attempt to unseal the vault instance (default 10s)
      --secret-shares int                   Total count of secret shares that exist (default 5)
      --secret-threshold int                Minimum required secret shares to unseal (default 3)
      --store-root-token                    should the root token be stored in the key store (default true)
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --enable-analytics                 Send analytical events to Google Analytics (default true)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO

* [vault-unsealer](/docs/reference/vault-unsealer.md)	 - Automates initialisation and unsealing of Hashicorp Vault

