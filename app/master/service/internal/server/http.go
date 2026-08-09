package server

import (
	masterpb "github.com/Servora-Kit/servora-example/api/gen/go/master/service/v1"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/service"
	corev1 "github.com/Servora-Kit/servora/api/gen/go/servora/core/v1"
	"log/slog"

	"github.com/Servora-Kit/servora/obs/audit"
	"github.com/Servora-Kit/servora/obs/metrics"
	svrhttp "github.com/Servora-Kit/servora/transport/server/http"
	"github.com/Servora-Kit/servora/transport/server/middleware"
	khttp "github.com/go-kratos/kratos/v3/transport/http"
)

func NewHTTPServer(c *corev1.Server, obs *corev1.Observability, mtc *metrics.Metrics, l *slog.Logger, auditor audit.Auditor, master *service.MasterService) *khttp.Server {
	httpLogger := l.With("scope", "http/server/master")

	mw := middleware.NewChainBuilder(httpLogger).
		WithTrace(obs.GetTrace()).
		WithMetrics(mtc).
		WithoutRateLimit().
		Build()
	// Business-mounted audit middleware driven by generated audit rules.
	mw = append(mw, audit.Middleware(auditor,
		audit.WithRulesFuncs(masterpb.AuditRules),
	))

	opts := []svrhttp.ServerOption{
		svrhttp.WithLogger(httpLogger),
		svrhttp.WithMiddleware(mw...),
		svrhttp.WithMetrics(mtc),
		svrhttp.WithServices(
			func(s *khttp.Server) {
				masterpb.RegisterMasterServiceHTTPServer(s, master)
			},
		),
	}
	if c != nil && c.Http != nil {
		opts = append(opts, svrhttp.WithConfig(c.Http))
	}
	return svrhttp.NewServer(opts...)
}
