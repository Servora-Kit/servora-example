package service

import (
	"context"
	"fmt"

	workerpb "github.com/Servora-Kit/servora-example/api/gen/go/worker/service/v1"
	"github.com/Servora-Kit/servora/obs/audit"
)

type WorkerService struct {
	workerpb.UnimplementedWorkerServiceServer
	auditor audit.Auditor
}

func NewWorkerService(auditor audit.Auditor) *WorkerService {
	return &WorkerService{auditor: auditor}
}

func (s *WorkerService) Hello(ctx context.Context, req *workerpb.HelloRequest) (*workerpb.HelloResponse, error) {
	reply := fmt.Sprintf("worker says hello, %s", req.GetGreeting())

	// Tier 2 demo: handler-level direct Emit for a business event.
	// Builds a CloudEvents envelope and emits via the Auditor interface.
	event := audit.NewEvent(ctx,
		audit.WithType("servora.worker.hello.created"),
		audit.WithSubject(req.GetGreeting()),
		audit.WithSeverity("INFO"),
	)
	_ = s.auditor.Emit(ctx, event)

	return &workerpb.HelloResponse{Reply: reply}, nil
}
