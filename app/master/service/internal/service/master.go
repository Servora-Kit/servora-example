package service

import (
	"context"

	masterpb "github.com/Servora-Kit/servora-example/api/gen/go/servora/master/service/v1"
	workerpb "github.com/Servora-Kit/servora-example/api/gen/go/servora/worker/service/v1"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/biz"
)

type MasterService struct {
	masterpb.UnimplementedMasterServiceServer
	uc *biz.MasterUsecase
}

func NewMasterService(uc *biz.MasterUsecase) *MasterService {
	return &MasterService{
		uc: uc,
	}
}

func (s *MasterService) Hello(ctx context.Context, req *workerpb.HelloRequest) (*workerpb.HelloResponse, error) {
	return s.uc.Hello(ctx, req)
}
