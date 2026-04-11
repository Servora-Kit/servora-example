package data

import (
	"context"
	"fmt"

	workerpb "github.com/Servora-Kit/servora-example/api/gen/go/servora/worker/service/v1"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/biz"
	conf "github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
	"github.com/Servora-Kit/servora/obs/logging"
	grpcclient "github.com/Servora-Kit/servora/transport/client/grpc"
	clientmw "github.com/Servora-Kit/servora/transport/client/middleware"
	"github.com/go-kratos/kratos/v2/registry"
)

const workerServiceName = "worker.service"

// workerRepo 封装 master 到 worker 的 RPC 访问。
type workerRepo struct {
	dialer *grpcclient.Dialer
	log    *logger.Helper
}

func NewWorkerDialer(data *conf.Data, trace *conf.Trace, discovery registry.Discovery, l logger.Logger) *grpcclient.Dialer {
	mw := clientmw.NewChainBuilder(l).
		WithTrace(trace).
		Build()
	return grpcclient.NewDialer(
		grpcclient.WithData(data),
		grpcclient.WithDiscovery(discovery),
		grpcclient.WithLogger(l),
		grpcclient.WithMiddleware(mw...),
	)
}

func NewWorkerRepo(d *grpcclient.Dialer, l logger.Logger) biz.WorkerRepo {
	return &workerRepo{
		dialer: d,
		log:    logger.For(l, "data/worker-client"),
	}
}

func (c *workerRepo) Hello(ctx context.Context, req *workerpb.HelloRequest) (*workerpb.HelloResponse, error) {
	conn, err := c.dialer.Dial(ctx, workerServiceName)
	if err != nil {
		return nil, fmt.Errorf("create worker grpc conn: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	resp, err := workerpb.NewWorkerServiceClient(conn).Hello(ctx, req)
	if err != nil {
		c.log.Errorf("worker hello failed: %v", err)
		return nil, fmt.Errorf("call worker hello: %w", err)
	}

	return resp, nil
}
