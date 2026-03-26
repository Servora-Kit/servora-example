//go:build wireinject
// +build wireinject

package main

import (
	"github.com/Servora-Kit/servora-example/app/worker/service/internal/server"
	"github.com/Servora-Kit/servora-example/app/worker/service/internal/service"
	conf "github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
	"github.com/Servora-Kit/servora/platform/bootstrap"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

func wireApp(*conf.Server, *conf.Registry, *conf.Data, *conf.App, *conf.Trace, *conf.Metrics, bootstrap.SvcIdentity, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(service.ProviderSet, server.ProviderSet, newApp))
}
