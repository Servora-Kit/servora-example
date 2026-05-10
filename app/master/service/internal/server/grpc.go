package server

import (
	masterpb "github.com/Servora-Kit/servora-example/api/gen/go/servora/master/service/v1"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/service"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/stubauth"
	conf "github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
	"github.com/Servora-Kit/servora/obs/audit"
	logger "github.com/Servora-Kit/servora/obs/logging"
	"github.com/Servora-Kit/servora/obs/telemetry"
	"github.com/Servora-Kit/servora/security/authn"
	"github.com/Servora-Kit/servora/security/authn/apikey"
	authjwt "github.com/Servora-Kit/servora/security/authn/jwt"
	svrgrpc "github.com/Servora-Kit/servora/transport/server/grpc"
	"github.com/Servora-Kit/servora/transport/server/middleware"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
)

func NewGRPCServer(c *conf.Server, trace *conf.Trace, mtc *telemetry.Metrics, l logger.Logger, rec *audit.Recorder, master *service.MasterService) *kgrpc.Server {
	grpcLogger := logger.With(l, "grpc/server/master")

	mw := middleware.NewChainBuilder(grpcLogger).
		WithTrace(trace).
		WithMetrics(mtc).
		WithoutRateLimit().
		WithAudit(rec).
		Build()
	// Lighthouse demo: real authn.Server dispatcher with jwt + apikey engines.
	// AuthnDetail is written by authn.Server; the OUTER audit.Collector emits
	// AUTHN_RESULT.
	_, jwtVerifier := stubauth.SharedKeypair()
	mw = append(mw, authn.Server(
		authn.Multi(
			authn.Named(authjwt.Scheme, authjwt.NewAuthenticator(authjwt.WithVerifier(jwtVerifier))),
			authn.Named(apikey.Scheme, apikey.NewAuthenticator(apikey.WithStore(stubauth.NewAPIKeyStore()))),
		),
		authn.WithRulesFuncs(masterpb.AuthnRules),
	))

	opts := []svrgrpc.ServerOption{
		svrgrpc.WithLogger(grpcLogger),
		svrgrpc.WithMiddleware(mw...),
		svrgrpc.WithServices(
			func(s *kgrpc.Server) {
				masterpb.RegisterMasterServiceServer(s, master)
			},
		),
	}
	if c != nil && c.Grpc != nil {
		opts = append(opts, svrgrpc.WithConfig(c.Grpc))
	}
	return svrgrpc.NewServer(opts...)
}
