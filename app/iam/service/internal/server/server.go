package server

import (
	"github.com/Servora-Kit/servora/app/iam/service/internal/server/middleware"
	"github.com/Servora-Kit/servora/pkg/governance/registry"
	"github.com/Servora-Kit/servora/pkg/governance/telemetry"
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(middleware.ProviderSet, registry.NewRegistrar, telemetry.NewMetrics, NewGRPCMiddleware, NewGRPCServer, NewHTTPMiddleware, NewHealthHandler, NewHTTPServer)
