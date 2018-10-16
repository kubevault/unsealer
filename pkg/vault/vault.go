package vault

import (
	vaultapi "github.com/hashicorp/vault/api"
)

func NewVaultClient(addr string, tlsConfig *vaultapi.TLSConfig) (*vaultapi.Client, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr
	cfg.ConfigureTLS(tlsConfig)
	return vaultapi.NewClient(cfg)
}
