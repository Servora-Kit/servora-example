package biz

import (
	"context"
	"fmt"

	workerpb "github.com/Servora-Kit/servora-example/api/gen/go/servora/worker/service/v1"
	"github.com/Servora-Kit/servora/obs/logging"
)

type WorkerRepo interface {
	Hello(ctx context.Context, req *workerpb.HelloRequest) (*workerpb.HelloResponse, error)
}

type MasterUsecase struct {
	worker WorkerRepo
	log    *logger.Helper
}

func NewMasterUsecase(worker WorkerRepo, l logger.Logger) *MasterUsecase {
	return &MasterUsecase{
		worker: worker,
		log:    logger.For(l, "master/biz"),
	}
}

func (uc *MasterUsecase) Hello(ctx context.Context, req *workerpb.HelloRequest) (*workerpb.HelloResponse, error) {
	resp, err := uc.worker.Hello(ctx, req)
	if err != nil {
		uc.log.Errorf("relay worker hello failed: %v", err)
		return nil, fmt.Errorf("relay worker hello: %w", err)
	}

	return &workerpb.HelloResponse{Reply: "master relay -> " + resp.GetReply()}, nil
}
