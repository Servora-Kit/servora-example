package service

import (
	"context"
	"fmt"

	workerpb "github.com/Servora-Kit/servora-example/api/gen/go/servora/worker/service/v1"
	auditpb "github.com/Servora-Kit/servora/api/gen/go/servora/audit/v1"
	"github.com/Servora-Kit/servora/obs/audit"
)

type WorkerService struct {
	workerpb.UnimplementedWorkerServiceServer
	rec *audit.Recorder
}

func NewWorkerService(rec *audit.Recorder) *WorkerService {
	return &WorkerService{rec: rec}
}

func (s *WorkerService) Hello(ctx context.Context, req *workerpb.HelloRequest) (*workerpb.HelloResponse, error) {
	reply := fmt.Sprintf("worker says hello, %s", req.GetGreeting())

	// Tier 2 demo: handler-level direct Emit for a business event.
	// (Tier 1 push-ctx is wired by demoIdentityMiddleware → audit.Collector.)
	s.rec.RecordResourceMutation(ctx,
		"/servora.worker.service.v1.WorkerService/Hello", nil,
		&auditpb.AuditTarget{Type: "hello.reply", Id: req.GetGreeting()},
		&auditpb.ResourceMutationDetail{
			MutationType: auditpb.ResourceMutationType_RESOURCE_MUTATION_TYPE_CREATE,
			ResourceType: "hello.reply",
			ResourceId:   req.GetGreeting(),
		},
		nil,
	)

	return &workerpb.HelloResponse{Reply: reply}, nil
}
