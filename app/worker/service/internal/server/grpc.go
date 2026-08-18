package server

import (
	workerpb "github.com/Servora-Kit/servora-example/api/gen/go/worker/service/v1"
	"github.com/Servora-Kit/servora-example/app/worker/service/internal/service"
	corev1 "github.com/Servora-Kit/servora/api/gen/go/servora/core/v1"
	"log/slog"

	"github.com/Servora-Kit/servora/obs/audit"
	"github.com/Servora-Kit/servora/obs/metrics"
	svrgrpc "github.com/Servora-Kit/servora/transport/server/grpc"
	"github.com/Servora-Kit/servora/transport/server/middleware"
	kgrpc "github.com/go-kratos/kratos/v3/transport/grpc"
)

func NewGRPCServer(c *corev1.Server, obs *corev1.Observability, mtc *metrics.Metrics, l *slog.Logger, auditor audit.Auditor, worker *service.WorkerService) *kgrpc.Server {
	grpcLogger := l.With("scope", "grpc/server/worker")

	mw := middleware.NewChainBuilder(grpcLogger).
		WithTrace(obs.GetTrace()).
		WithMetrics(mtc).
		WithoutRateLimit().
		Build()
	// Business-mounted audit middleware driven by generated audit rules.
	mw = append(mw, audit.Middleware(auditor,
		audit.WithRulesFuncs(workerpb.AuditRules),
	))

	opts := []svrgrpc.ServerOption{
		svrgrpc.WithMiddleware(mw...),
		svrgrpc.WithServices(
			func(s *kgrpc.Server) {
				workerpb.RegisterWorkerServiceServer(s, worker)
			},
		),
	}
	if c != nil && c.Grpc != nil {
		opts = append(opts, svrgrpc.WithConfig(c.Grpc))
	}
	return svrgrpc.NewServer(opts...)
}
