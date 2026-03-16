package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"

	"github.com/Servora-Kit/servora/pkg/actor"
)

// requireOrgScope extracts the authenticated user ID and organization scope
// from context. The organization ID is injected by the ScopeFromHeaders
// middleware from the X-Organization-ID header.
func requireOrgScope(ctx context.Context) (userID, orgID string, err error) {
	a, ok := actor.FromContext(ctx)
	if !ok || a.Type() != actor.TypeUser {
		return "", "", errors.Unauthorized("UNAUTHORIZED", "unauthorized")
	}
	ua, ok := a.(*actor.UserActor)
	if !ok {
		return "", "", errors.Unauthorized("UNAUTHORIZED", "unauthorized")
	}
	if ua.OrganizationID() == "" {
		return "", "", errors.BadRequest("MISSING_ORGANIZATION_SCOPE",
			"missing X-Organization-ID header")
	}
	return ua.ID(), ua.OrganizationID(), nil
}
