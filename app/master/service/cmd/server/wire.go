//go:build wireinject
// +build wireinject

package main

import (
	"github.com/Servora-Kit/servora-example/app/master/service/internal/biz"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/data"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/server"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/service"
	"log/slog"

	tcpconf "github.com/Servora-Kit/servora-transport/server/tcp/gen/conf"
	corev1 "github.com/Servora-Kit/servora/api/gen/go/servora/core/v1"
	"github.com/Servora-Kit/servora/core/bootstrap"
	"github.com/Servora-Kit/servora/core/registry"
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
)

func wireApp(*corev1.Server, *corev1.Discovery, *corev1.Registry, *corev1.Data, *corev1.App, *corev1.Trace, *corev1.Metrics, *tcpconf.Server, bootstrap.SvcIdentity, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(registry.NewDiscovery, data.ProviderSet, biz.ProviderSet, service.ProviderSet, server.ProviderSet, newApp))
}
