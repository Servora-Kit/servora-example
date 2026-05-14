package server

import (
	masterpb "github.com/Servora-Kit/servora-example/api/gen/go/servora/master/service/v1"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/service"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/stubauth"
	corev1 "github.com/Servora-Kit/servora/api/gen/go/servora/core/v1"
	"github.com/Servora-Kit/servora/obs/audit"
	logger "github.com/Servora-Kit/servora/obs/logging"
	"github.com/Servora-Kit/servora/obs/telemetry"
	"github.com/Servora-Kit/servora/security/authn"
	"github.com/Servora-Kit/servora/security/authn/apikey"
	authjwt "github.com/Servora-Kit/servora/security/authn/jwt"
	svrhttp "github.com/Servora-Kit/servora/transport/server/http"
	"github.com/Servora-Kit/servora/transport/server/middleware"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

func NewHTTPServer(c *corev1.Server, trace *corev1.Trace, mtc *telemetry.Metrics, l logger.Logger, auditor audit.Auditor, master *service.MasterService) *khttp.Server {
	httpLogger := logger.With(l, "http/server/master")

	mw := middleware.NewChainBuilder(httpLogger).
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
		authn.WithRulesFuncs(masterpb.AuthnRules),
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
