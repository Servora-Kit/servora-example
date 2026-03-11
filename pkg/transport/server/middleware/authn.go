package middleware

import (
	"context"
	"fmt"
	"strings"

	gojwt "github.com/golang-jwt/jwt/v5"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/Servora-Kit/servora/pkg/actor"
	jwtpkg "github.com/Servora-Kit/servora/pkg/jwt"
)

// tokenKey stores the raw bearer token in context for downstream propagation.
type tokenKey struct{}

// TokenFromContext retrieves the raw bearer token stored by the authn middleware.
func TokenFromContext(ctx context.Context) (string, bool) {
	t, ok := ctx.Value(tokenKey{}).(string)
	return t, ok
}

// UserClaimsMapper converts parsed JWT MapClaims into an actor.Actor.
type UserClaimsMapper func(claims gojwt.MapClaims) (actor.Actor, error)

// AuthnOption configures the authn middleware.
type AuthnOption func(*authnConfig)

type authnConfig struct {
	verifier     *jwtpkg.Verifier
	claimsMapper UserClaimsMapper
	errorHandler func(ctx context.Context, err error) error
}

func WithVerifier(v *jwtpkg.Verifier) AuthnOption {
	return func(c *authnConfig) { c.verifier = v }
}

func WithClaimsMapper(m UserClaimsMapper) AuthnOption {
	return func(c *authnConfig) { c.claimsMapper = m }
}

func WithAuthnErrorHandler(h func(ctx context.Context, err error) error) AuthnOption {
	return func(c *authnConfig) { c.errorHandler = h }
}

func defaultClaimsMapper(claims gojwt.MapClaims) (actor.Actor, error) {
	id := claimString(claims, "sub")
	if id == "" {
		id = claimString(claims, "id")
	}
	name := claimString(claims, "name")
	email := claimString(claims, "email")

	metadata := make(map[string]string)
	if role := claimString(claims, "role"); role != "" {
		metadata["role"] = role
	}

	return actor.NewUserActor(id, name, email, metadata), nil
}

func claimString(claims gojwt.MapClaims, key string) string {
	v, ok := claims[key]
	if !ok {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%.0f", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Authn creates a Kratos middleware that verifies JWT tokens and injects
// an actor.Actor into the request context.
//
// If no token is present, an AnonymousActor is injected.
// Combine with selector.Server + WhiteList for public route handling.
func Authn(opts ...AuthnOption) middleware.Middleware {
	cfg := &authnConfig{
		claimsMapper: defaultClaimsMapper,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				ctx = actor.NewContext(ctx, actor.NewAnonymousActor())
				return handler(ctx, req)
			}

			tokenString := extractBearerToken(tr.RequestHeader().Get("Authorization"))
			if tokenString == "" {
				ctx = actor.NewContext(ctx, actor.NewAnonymousActor())
				return handler(ctx, req)
			}

			if cfg.verifier == nil {
				ctx = actor.NewContext(ctx, actor.NewAnonymousActor())
				ctx = context.WithValue(ctx, tokenKey{}, tokenString)
				return handler(ctx, req)
			}

			claims := gojwt.MapClaims{}
			if err := cfg.verifier.Verify(tokenString, claims); err != nil {
				if cfg.errorHandler != nil {
					return nil, cfg.errorHandler(ctx, err)
				}
				return nil, err
			}

			a, err := cfg.claimsMapper(claims)
			if err != nil {
				if cfg.errorHandler != nil {
					return nil, cfg.errorHandler(ctx, err)
				}
				return nil, err
			}

			ctx = actor.NewContext(ctx, a)
			ctx = context.WithValue(ctx, tokenKey{}, tokenString)
			return handler(ctx, req)
		}
	}
}

func extractBearerToken(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return parts[1]
}
