// demoIdentityMiddleware is a fixture for the demo/audit branch.
//
// It pushes a fixed AuthnDetail into the per-request audit holder so
// audit.Collector can emit an AUTHN_RESULT event without a real authn
// middleware. Replace with `authn.Server(...)` once P0-4 (proto-driven authn)
// lands; the rest of the chain stays unchanged.
package server

import (
	"context"

	auditpb "github.com/Servora-Kit/servora/api/gen/go/servora/audit/v1"
	"github.com/Servora-Kit/servora/obs/audit"
	"github.com/go-kratos/kratos/v2/middleware"
)

func demoIdentityMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			audit.WithAuthnResult(ctx, &auditpb.AuthnDetail{
				Method:  "demo",
				Success: true,
			})
			return handler(ctx, req)
		}
	}
}
