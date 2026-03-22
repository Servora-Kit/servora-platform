package oidc

import (
	"encoding/hex"
	"fmt"

	"github.com/zitadel/oidc/v3/pkg/op"

	"github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
)

func NewProvider(appCfg *conf.App, storage op.Storage) (*op.Provider, error) {
	cryptoKey, err := decodeCryptoKey(appCfg.GetOidc().GetCryptoKey())
	if err != nil {
		return nil, fmt.Errorf("oidc: decode crypto_key: %w", err)
	}

	config := &op.Config{
		CryptoKey:                cryptoKey,
		DefaultLogoutRedirectURI: appCfg.GetOidc().GetDefaultLogoutRedirectUri(),
		CodeMethodS256:           true,
		AuthMethodPost:           true,
		GrantTypeRefreshToken:    appCfg.GetOidc().GetGrantTypeRefreshToken(),
	}

	issuer := appCfg.GetExternalUrl()
	opts := []op.Option{
		op.WithAllowInsecure(),
	}

	return op.NewProvider(config, storage, op.StaticIssuer(issuer), opts...)
}

func decodeCryptoKey(hexKey string) ([32]byte, error) {
	var key [32]byte
	b, err := hex.DecodeString(hexKey)
	if err != nil {
		return key, fmt.Errorf("hex decode: %w", err)
	}
	if len(b) != 32 {
		return key, fmt.Errorf("expected 32 bytes, got %d", len(b))
	}
	copy(key[:], b)
	return key, nil
}
