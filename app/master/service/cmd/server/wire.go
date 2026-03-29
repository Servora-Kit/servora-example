//go:build wireinject
// +build wireinject

package main

import (
	"github.com/Servora-Kit/servora-example/app/master/service/internal/biz"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/data"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/server"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/service"
	tcpconf "github.com/Servora-Kit/servora-transport/server/tcp/gen/conf"
	conf "github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
	"github.com/Servora-Kit/servora/platform/bootstrap"
	"github.com/Servora-Kit/servora/platform/registry"
	"github.com/Servora-Kit/servora/transport/client"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

func wireApp(*conf.Server, *conf.Discovery, *conf.Registry, *conf.Data, *conf.App, *conf.Trace, *conf.Metrics, *tcpconf.Server, bootstrap.SvcIdentity, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(registry.NewDiscovery, client.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, server.ProviderSet, newApp))
}
