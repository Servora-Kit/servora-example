package server

import (
	kmw "github.com/go-kratos/kratos/v2/middleware"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"

	authpb "github.com/Servora-Kit/servora/api/gen/go/auth/service/v1"
	"github.com/Servora-Kit/servora/api/gen/go/conf/v1"
	orgpb "github.com/Servora-Kit/servora/api/gen/go/organization/service/v1"
	projectpb "github.com/Servora-Kit/servora/api/gen/go/project/service/v1"
	testpb "github.com/Servora-Kit/servora/api/gen/go/test/service/v1"
	userpb "github.com/Servora-Kit/servora/api/gen/go/user/service/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/service"
	"github.com/Servora-Kit/servora/pkg/governance/telemetry"
	"github.com/Servora-Kit/servora/pkg/logger"
	"github.com/Servora-Kit/servora/pkg/transport/server/grpc"
	svrmw "github.com/Servora-Kit/servora/pkg/transport/server/middleware"
)

// GRPCMiddleware 用于 Wire 注入的中间件切片包装类型
type GRPCMiddleware []kmw.Middleware

// NewGRPCMiddleware 创建 gRPC 中间件
func NewGRPCMiddleware(
	trace *conf.Trace,
	mtc *telemetry.Metrics,
	l logger.Logger,
) GRPCMiddleware {
	return svrmw.NewChainBuilder(logger.With(l, logger.WithModule("grpc/server/iam-service"))).
		WithTrace(trace).
		WithMetrics(mtc).
		Build()
}

// NewGRPCServer new a gRPC server.
func NewGRPCServer(
	c *conf.Server,
	mw GRPCMiddleware,
	l logger.Logger,
	auth *service.AuthService,
	user *service.UserService,
	test *service.TestService,
	org *service.OrganizationService,
	proj *service.ProjectService,
) *kgrpc.Server {
	glog := logger.With(l, logger.WithModule("grpc/server/iam-service"))

	opts := []grpc.ServerOption{
		grpc.WithLogger(glog),
		grpc.WithMiddleware(mw...),
	}
	if c != nil && c.Grpc != nil {
		opts = append(opts, grpc.WithConfig(c.Grpc))
	}

	srv := grpc.NewServer(opts...)

	authpb.RegisterAuthServiceServer(srv, auth)
	userpb.RegisterUserServiceServer(srv, user)
	testpb.RegisterTestServiceServer(srv, test)
	orgpb.RegisterOrganizationServiceServer(srv, org)
	projectpb.RegisterProjectServiceServer(srv, proj)

	return srv
}
