package server

import (
	kmw "github.com/go-kratos/kratos/v2/middleware"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"

	"github.com/Servora-Kit/servora/api/gen/go/conf/v1"
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
	return svrmw.NewChainBuilder(logger.With(l, logger.WithModule("grpc/server/servora-service"))).
		WithTrace(trace).
		WithMetrics(mtc).
		Build()
}

// NewGRPCServer new a gRPC server.
func NewGRPCServer(
	c *conf.Server,
	mw GRPCMiddleware,
	l logger.Logger,
) *kgrpc.Server {
	glog := logger.With(l, logger.WithModule("grpc/server/servora-service"))

	opts := []grpc.ServerOption{
		grpc.WithLogger(glog),
		grpc.WithMiddleware(mw...),
	}
	if c != nil && c.Grpc != nil {
		opts = append(opts, grpc.WithConfig(c.Grpc))
	}

	return grpc.NewServer(opts...)
}
