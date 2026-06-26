package stubauth

import (
	"context"

	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
)

// PassthroughAuthHeaders returns a Kratos client-side middleware that
// copies inbound auth-related headers from the server-side transport
// (the request the master is currently handling) onto the outbound
// transport (the gRPC call master is about to make to worker).
//
// Without this, master's downstream calls land at worker with no
// credentials, and worker's MODE_REQUIRED authn rejects them — even
// though master's own inbound authn already passed.
//
// Demo-only: a real production stack should pick a deliberate
// service-to-service identity strategy (mTLS / S2S apikey / signed
// JWT issued by master itself) rather than blindly forwarding the
// caller's credentials. This middleware is the simplest way to
// demonstrate the multi-scheme flow end-to-end.
//
// Headers forwarded: "Authorization", "X-API-Key".
func PassthroughAuthHeaders() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			inbound, hasInbound := transport.FromServerContext(ctx)
			outbound, hasOutbound := transport.FromClientContext(ctx)
			if hasInbound && hasOutbound {
				for _, name := range []string{"Authorization", "X-API-Key"} {
					if v := inbound.RequestHeader().Get(name); v != "" {
						outbound.RequestHeader().Set(name, v)
					}
				}
			}
			return handler(ctx, req)
		}
	}
}
