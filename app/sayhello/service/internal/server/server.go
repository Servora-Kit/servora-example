package server

import (
	"github.com/google/wire"
	"github.com/Servora-Kit/servora/platform/registry"
	"github.com/Servora-Kit/servora/obs/telemetry"
)

var ProviderSet = wire.NewSet(registry.NewRegistrar, telemetry.NewMetrics, NewGRPCServer)
