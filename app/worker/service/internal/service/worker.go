package service

import (
	"context"
	"fmt"

	workerpb "github.com/Servora-Kit/servora-example/api/gen/go/servora/worker/service/v1"
)

type WorkerService struct {
	workerpb.UnimplementedWorkerServiceServer
}

func NewWorkerService() *WorkerService {
	return &WorkerService{}
}

func (s *WorkerService) Hello(ctx context.Context, req *workerpb.HelloRequest) (*workerpb.HelloResponse, error) {
	_ = ctx
	return &workerpb.HelloResponse{Reply: fmt.Sprintf("worker says hello, %s", req.GetGreeting())}, nil
}
