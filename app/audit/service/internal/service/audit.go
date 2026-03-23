package service

import (
	"context"

	auditsvcpb "github.com/Servora-Kit/servora/api/gen/go/servora/audit/service/v1"
	"github.com/Servora-Kit/servora/app/audit/service/internal/data"
)

// AuditService implements both AuditQueryService (gRPC) and AuditHTTPService (HTTP).
type AuditService struct {
	auditsvcpb.UnimplementedAuditQueryServiceServer
	auditsvcpb.UnimplementedAuditHTTPServiceServer
	repo *data.AuditRepo
}

// NewAuditService creates a new AuditService.
func NewAuditService(repo *data.AuditRepo) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) ListAuditEvents(ctx context.Context, req *auditsvcpb.ListAuditEventsRequest) (*auditsvcpb.ListAuditEventsResponse, error) {
	items, nextToken, err := s.repo.ListEvents(ctx, req)
	if err != nil {
		return nil, err
	}
	return &auditsvcpb.ListAuditEventsResponse{
		Events:        items,
		NextPageToken: nextToken,
	}, nil
}

func (s *AuditService) CountAuditEvents(ctx context.Context, req *auditsvcpb.CountAuditEventsRequest) (*auditsvcpb.CountAuditEventsResponse, error) {
	count, err := s.repo.CountEvents(ctx, req)
	if err != nil {
		return nil, err
	}
	return &auditsvcpb.CountAuditEventsResponse{TotalCount: count}, nil
}
