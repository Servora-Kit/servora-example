//go:build wireinject
// +build wireinject

package main

import (
	"github.com/Servora-Kit/servora-example/app/worker/service/internal/server"
	"github.com/Servora-Kit/servora-example/app/worker/service/internal/service"

	"github.com/Servora-Kit/servora/core/bootstrap"
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
)

func wireApp(*bootstrap.Runtime) (*kratos.App, func(), error) {
	panic(wire.Build(bootstrap.ProviderSet, service.ProviderSet, server.ProviderSet, newApp))
}
