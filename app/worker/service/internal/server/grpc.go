package server

import (
	workerpb "github.com/Servora-Kit/servora-example/api/gen/go/servora/worker/service/v1"
	"github.com/Servora-Kit/servora-example/app/worker/service/internal/service"
	"github.com/Servora-Kit/servora-example/app/worker/service/internal/stubauth"
	corev1 "github.com/Servora-Kit/servora/api/gen/go/servora/core/v1"
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

func NewGRPCServer(c *corev1.Server, trace *corev1.Trace, mtc *telemetry.Metrics, l logger.Logger, auditor audit.Auditor, worker *service.WorkerService) *kgrpc.Server {
	grpcLogger := logger.With(l, "grpc/server/worker")

	mw := middleware.NewChainBuilder(grpcLogger).
		WithTrace(trace).
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
		authn.WithRulesFuncs(workerpb.AuthnRules),
	))

	opts := []svrgrpc.ServerOption{
		svrgrpc.WithLogger(grpcLogger),
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
