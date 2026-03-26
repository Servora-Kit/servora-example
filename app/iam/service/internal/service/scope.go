package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"

	"github.com/Servora-Kit/servora/core/actor"
)

// requireAuthenticatedUser extracts the authenticated user ID from context.
func requireAuthenticatedUser(ctx context.Context) (userID string, err error) {
	a, ok := actor.FromContext(ctx)
	if !ok || a.Type() != actor.TypeUser {
		return "", errors.Unauthorized("UNAUTHORIZED", "unauthorized")
	}
	return a.ID(), nil
}

