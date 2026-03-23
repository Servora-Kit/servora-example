package server

import (
	"github.com/Servora-Kit/servora/pkg/governance/registry"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(registry.NewRegistrar, NewGRPCServer, NewHTTPServer)
