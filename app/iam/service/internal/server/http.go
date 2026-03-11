package server

import (
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	"github.com/Servora-Kit/servora/api/gen/go/conf/v1"
	iamv1 "github.com/Servora-Kit/servora/api/gen/go/iam/service/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/consts"
	innermw "github.com/Servora-Kit/servora/app/iam/service/internal/server/middleware"
	"github.com/Servora-Kit/servora/app/iam/service/internal/service"
	"github.com/Servora-Kit/servora/pkg/governance/telemetry"
	"github.com/Servora-Kit/servora/pkg/health"
	"github.com/Servora-Kit/servora/pkg/logger"
	"github.com/Servora-Kit/servora/pkg/redis"
	"github.com/Servora-Kit/servora/pkg/transport/server/http"
	svrmw "github.com/Servora-Kit/servora/pkg/transport/server/middleware"
)

// HTTPMiddleware 用于 Wire 注入的中间件切片包装类型
type HTTPMiddleware []middleware.Middleware

func NewHTTPMiddleware(
	trace *conf.Trace,
	m *telemetry.Metrics,
	l logger.Logger,
	authJWT innermw.AuthJWT,
) HTTPMiddleware {
	ms := svrmw.NewChainBuilder(logger.With(l, logger.WithModule("http/server/iam-service"))).
		WithTrace(trace).
		WithMetrics(m).
		Build()

	publicWhitelist := svrmw.NewWhiteList(svrmw.Exact,
		iamv1.OperationAuthServiceLoginByEmailPassword,
		iamv1.OperationAuthServiceRefreshToken,
		iamv1.OperationAuthServiceSignupByEmail,
		iamv1.OperationTestServiceTest,
		iamv1.OperationTestServiceHello,
	)

	userWhitelist := svrmw.NewWhiteList(svrmw.Exact,
		iamv1.OperationUserServiceCurrentUserInfo,
		iamv1.OperationUserServiceUpdateUser,
		iamv1.OperationAuthServiceLogout,
		iamv1.OperationTestServicePrivateTest,
	)

	// Admin 权限排除白名单 = 公开接口 ∪ User 级接口
	adminExcludeWhitelist := publicWhitelist.Merge(userWhitelist)

	ms = append(ms,
		selector.Server(authJWT(consts.User)).
			Match(publicWhitelist.MatchFunc()).
			Build(),
		selector.Server(authJWT(consts.Admin)).
			Match(adminExcludeWhitelist.MatchFunc()).
			Build(),
	)

	return ms
}

func NewHealthHandler(redisClient *redis.Client) *health.Handler {
	return health.NewHandlerWithDefaults(health.DefaultDeps{
		Redis: redisClient,
	})
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Server,
	mw HTTPMiddleware,
	mtc *telemetry.Metrics,
	l logger.Logger,
	h *health.Handler,
	auth *service.AuthService,
	user *service.UserService,
	test *service.TestService,
) *khttp.Server {
	hlog := logger.With(l, logger.WithModule("http/server/iam-service"))

	opts := []http.ServerOption{
		http.WithLogger(hlog),
		http.WithMiddleware(mw...),
		http.WithMetrics(mtc),
		http.WithHealthCheck(h),
		http.WithServices(
			func(s *khttp.Server) { iamv1.RegisterAuthServiceHTTPServer(s, auth) },
			func(s *khttp.Server) { iamv1.RegisterUserServiceHTTPServer(s, user) },
			func(s *khttp.Server) { iamv1.RegisterTestServiceHTTPServer(s, test) },
		),
	}
	if c != nil && c.Http != nil {
		opts = append(opts, http.WithConfig(c.Http))
		opts = append(opts, http.WithCORS(c.Http.Cors))
	}

	return http.NewServer(opts...)
}
