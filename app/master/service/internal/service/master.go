package service

import (
	"context"
	"fmt"

	masterpb "github.com/Servora-Kit/servora-example/api/gen/go/servora/master/service/v1"
	workerpb "github.com/Servora-Kit/servora-example/api/gen/go/servora/worker/service/v1"
	"github.com/Servora-Kit/servora/obs/logging"
	"github.com/Servora-Kit/servora/transport/client"
)

type MasterService struct {
	masterpb.UnimplementedMasterServiceServer
	client client.Client
	log    *logger.Helper
}

func NewMasterService(c client.Client, l logger.Logger) *MasterService {
	return &MasterService{
		client: c,
		log:    logger.For(l, "master/service"),
	}
}

func (s *MasterService) Hello(ctx context.Context, req *workerpb.HelloRequest) (*workerpb.HelloResponse, error) {
	conn, err := client.GetGRPCConn(ctx, s.client, "worker.service")
	if err != nil {
		return nil, fmt.Errorf("create worker grpc conn: %w", err)
	}

	resp, err := workerpb.NewWorkerServiceClient(conn).Hello(ctx, req)
	if err != nil {
		s.log.Errorf("worker Hello failed: %v", err)
		return nil, fmt.Errorf("call worker hello: %w", err)
	}

	return &workerpb.HelloResponse{Reply: "master relay -> " + resp.GetReply()}, nil
}
