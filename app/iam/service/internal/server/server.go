package server

import (
	"fmt"

	"github.com/Servora-Kit/servora/api/gen/go/conf/v1"
	"github.com/Servora-Kit/servora/pkg/governance/registry"
	"github.com/Servora-Kit/servora/pkg/governance/telemetry"
	"github.com/Servora-Kit/servora/pkg/jwks"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(registry.NewRegistrar, telemetry.NewMetrics, NewKeyManager, NewGRPCMiddleware, NewGRPCServer, NewHTTPMiddleware, NewHealthHandler, NewHTTPServer)

func NewKeyManager(cfg *conf.App) (*jwks.KeyManager, error) {
	if cfg.Jwt == nil {
		return nil, fmt.Errorf("jwt configuration is required")
	}
	var opts []jwks.Option
	if cfg.Jwt.PrivateKeyPath != "" {
		opts = append(opts, jwks.WithPrivateKeyPath(cfg.Jwt.PrivateKeyPath))
	} else if cfg.Jwt.PrivateKeyPem != "" {
		opts = append(opts, jwks.WithPrivateKeyPEM([]byte(cfg.Jwt.PrivateKeyPem)))
	} else {
		return nil, fmt.Errorf("jwt: no private key configured (set private_key_path or private_key_pem)")
	}
	return jwks.NewKeyManager(opts...)
}
