package jwks

import (
	"fmt"

	conf "github.com/Servora-Kit/servora/api/gen/go/conf/v1"
)

// NewKeyManagerFromConfig creates a KeyManager by reading JWT settings from
// the shared app configuration. It bridges conf.App.Jwt fields to jwks.Option
// so callers don't need to repeat the mapping logic.
func NewKeyManagerFromConfig(cfg *conf.App) (*KeyManager, error) {
	if cfg.Jwt == nil {
		return nil, fmt.Errorf("jwt configuration is required")
	}
	var opts []Option
	if cfg.Jwt.PrivateKeyPath != "" {
		opts = append(opts, WithPrivateKeyPath(cfg.Jwt.PrivateKeyPath))
	} else if cfg.Jwt.PrivateKeyPem != "" {
		opts = append(opts, WithPrivateKeyPEM([]byte(cfg.Jwt.PrivateKeyPem)))
	} else {
		return nil, fmt.Errorf("jwt: no private key configured (set private_key_path or private_key_pem)")
	}
	return NewKeyManager(opts...)
}
