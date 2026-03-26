package server

import (
	workerpb "github.com/Servora-Kit/servora-example/api/gen/go/servora/worker/service/v1"
	"github.com/Servora-Kit/servora-example/app/worker/service/internal/service"
	conf "github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
	"github.com/Servora-Kit/servora/obs/logging"
	"github.com/Servora-Kit/servora/obs/telemetry"
	"github.com/Servora-Kit/servora/transport/server/grpc"
	"github.com/Servora-Kit/servora/transport/server/middleware"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
)

func NewGRPCServer(c *conf.Server, trace *conf.Trace, mtc *telemetry.Metrics, l logger.Logger, worker *service.WorkerService) *kgrpc.Server {
	grpcLogger := logger.With(l, "grpc/server/worker")

	mw := middleware.NewChainBuilder(grpcLogger).
		WithTrace(trace).
		WithMetrics(mtc).
		WithoutRateLimit().
		Build()

	opts := []grpc.ServerOption{
		grpc.WithLogger(grpcLogger),
		grpc.WithMiddleware(mw...),
		grpc.WithServices(func(s *kgrpc.Server) {
			workerpb.RegisterWorkerServiceServer(s, worker)
		}),
	}
	if c != nil && c.Grpc != nil {
		opts = append(opts, grpc.WithConfig(c.Grpc))
	}
	return grpc.NewServer(opts...)
}
