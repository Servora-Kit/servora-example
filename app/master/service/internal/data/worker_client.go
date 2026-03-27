package data

import (
	"context"
	"fmt"

	workerpb "github.com/Servora-Kit/servora-example/api/gen/go/servora/worker/service/v1"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/biz"
	"github.com/Servora-Kit/servora/obs/logging"
	"github.com/Servora-Kit/servora/transport/client"
	"github.com/Servora-Kit/servora/transport/runtime"
	gogrpc "google.golang.org/grpc"
)

const workerServiceName = "worker.service"

// workerRepo 封装 master 到 worker 的 RPC 访问。
type workerRepo struct {
	client client.Client
	log    *logger.Helper
}

func NewWorkerRepo(c client.Client, l logger.Logger) biz.WorkerRepo {
	return &workerRepo{
		client: c,
		log:    logger.For(l, "data/worker-client"),
	}
}

func (c *workerRepo) Hello(ctx context.Context, req *workerpb.HelloRequest) (*workerpb.HelloResponse, error) {
	conn, err := client.GetValue[gogrpc.ClientConnInterface](ctx, c.client, runtime.ClientDialInput{
		Protocol: "grpc",
		Target:   workerServiceName,
	})
	if err != nil {
		return nil, fmt.Errorf("create worker grpc conn: %w", err)
	}

	resp, err := workerpb.NewWorkerServiceClient(conn).Hello(ctx, req)
	if err != nil {
		c.log.Errorf("worker hello failed: %v", err)
		return nil, fmt.Errorf("call worker hello: %w", err)
	}

	return resp, nil
}
