package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	svrmw "github.com/Servora-Kit/servora/pkg/transport/server/middleware"
)

// TokenPropagation creates a client middleware that forwards the bearer token
// from the incoming request context to outgoing service-to-service calls.
//
// It reads the token stored by the server-side Authn middleware and sets it
// as the Authorization header on outgoing requests.
func TokenPropagation() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			token, ok := svrmw.TokenFromContext(ctx)
			if !ok || token == "" {
				return handler(ctx, req)
			}

			tr, ok := transport.FromClientContext(ctx)
			if !ok {
				return handler(ctx, req)
			}

			tr.RequestHeader().Set("Authorization", "Bearer "+token)
			return handler(ctx, req)
		}
	}
}
