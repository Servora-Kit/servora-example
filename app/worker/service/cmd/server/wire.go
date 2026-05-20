//go:build wireinject
// +build wireinject

package main

import (
	"github.com/Servora-Kit/servora-example/app/worker/service/internal/server"
	"github.com/Servora-Kit/servora-example/app/worker/service/internal/service"
	"log/slog"

	corev1 "github.com/Servora-Kit/servora/api/gen/go/servora/core/v1"
	"github.com/Servora-Kit/servora/core/bootstrap"
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
)

func wireApp(*corev1.Server, *corev1.Registry, *corev1.Data, *corev1.App, *corev1.Trace, *corev1.Metrics, bootstrap.SvcIdentity, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(service.ProviderSet, server.ProviderSet, newApp))
}
