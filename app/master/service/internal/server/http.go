package server

import (
	masterpb "github.com/Servora-Kit/servora-example/api/gen/go/servora/master/service/v1"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/service"
	conf "github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
	logger "github.com/Servora-Kit/servora/obs/logging"
	"github.com/Servora-Kit/servora/obs/telemetry"
	svrhttp "github.com/Servora-Kit/servora/transport/server/http"
	"github.com/Servora-Kit/servora/transport/server/middleware"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

func NewHTTPServer(c *conf.Server, trace *conf.Trace, mtc *telemetry.Metrics, l logger.Logger, master *service.MasterService) *khttp.Server {
	httpLogger := logger.With(l, "http/server/master")

	mw := middleware.NewChainBuilder(httpLogger).
		WithTrace(trace).
		WithMetrics(mtc).
		WithoutRateLimit().
		Build()

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
		opts = append(opts, svrhttp.WithCORS(c.Http.Cors))
	}
	return svrhttp.NewServer(opts...)
}
