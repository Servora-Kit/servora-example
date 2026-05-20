package biz

import (
	"context"
	"fmt"

	"log/slog"

	workerpb "github.com/Servora-Kit/servora-example/api/gen/go/worker/service/v1"
)

type WorkerRepo interface {
	Hello(ctx context.Context, req *workerpb.HelloRequest) (*workerpb.HelloResponse, error)
}

type MasterUsecase struct {
	worker WorkerRepo
	log    *slog.Logger
}

func NewMasterUsecase(worker WorkerRepo, l *slog.Logger) *MasterUsecase {
	return &MasterUsecase{
		worker: worker,
		log:    l.With("scope", "master/biz"),
	}
}

func (uc *MasterUsecase) Hello(ctx context.Context, req *workerpb.HelloRequest) (*workerpb.HelloResponse, error) {
	resp, err := uc.worker.Hello(ctx, req)
	if err != nil {
		uc.log.Error("relay worker hello failed", "err", err)
		return nil, fmt.Errorf("relay worker hello: %w", err)
	}

	return &workerpb.HelloResponse{Reply: "master relay -> " + resp.GetReply()}, nil
}
