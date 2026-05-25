package server

import (
	masterpb "github.com/Servora-Kit/servora-example/api/gen/go/master/service/v1"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/service"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/stubauth"
	corev1 "github.com/Servora-Kit/servora/api/gen/go/servora/core/v1"
	"log/slog"

	"github.com/Servora-Kit/servora/obs/audit"
	"github.com/Servora-Kit/servora/obs/metrics"
	"github.com/Servora-Kit/servora/security/authn"
	"github.com/Servora-Kit/servora/security/authn/apikey"
	authjwt "github.com/Servora-Kit/servora/security/authn/jwt"
	svrgrpc "github.com/Servora-Kit/servora/transport/server/grpc"
	"github.com/Servora-Kit/servora/transport/server/middleware"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
)

func NewGRPCServer(c *corev1.Server, obs *corev1.Observability, mtc *metrics.Metrics, l *slog.Logger, auditor audit.Auditor, master *service.MasterService) *kgrpc.Server {
	grpcLogger := l.With("scope", "grpc/server/master")

	mw := middleware.NewChainBuilder(grpcLogger).
		WithTrace(obs.GetTrace()).
		WithMetrics(mtc).
		WithoutRateLimit().
		Build()
	// Business-mounted audit middleware with subject extraction from jwt + apikey.
	mw = append(mw, audit.Middleware(auditor,
		audit.WithSubjectFunc(authn.SubjectFromAny(authjwt.SubjectFrom, apikey.SubjectFrom)),
		audit.WithAuthTypeFunc(authn.AuthTypeFrom),
	))
	// Lighthouse demo: real authn.Server dispatcher with jwt + apikey engines.
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
