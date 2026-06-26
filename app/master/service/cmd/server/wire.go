//go:build wireinject
// +build wireinject

package main

import (
	"github.com/Servora-Kit/servora-example/app/master/service/internal/biz"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/data"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/server"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/service"

	tcpconf "github.com/Servora-Kit/servora-transport/server/tcp/gen/conf"
	"github.com/Servora-Kit/servora/core/bootstrap"
	"github.com/Servora-Kit/servora/core/registry"
	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"
)

func wireApp(*bootstrap.Runtime, *tcpconf.Server) (*kratos.App, func(), error) {
	panic(wire.Build(bootstrap.ProviderSet, registry.NewDiscovery, data.ProviderSet, biz.ProviderSet, service.ProviderSet, server.ProviderSet, newApp))
}
