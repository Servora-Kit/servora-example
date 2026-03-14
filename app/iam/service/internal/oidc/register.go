package oidc

import (
	stdhttp "net/http"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/zitadel/oidc/v3/pkg/op"
)

// OIDC path prefixes delegated to the Provider (zitadel/oidc chi router).
var oidcPrefixes = []string{
	"/.well-known/",
	"/authorize",
	"/oauth/",
	"/userinfo",
	"/keys",
	"/revoke",
	"/end_session",
	"/callback",
}

// Register mounts the OIDC provider and login handlers on the Kratos HTTP server.
// Server remains protocol-agnostic; all OIDC route knowledge lives here.
func Register(s *khttp.Server, provider *op.Provider, lh *LoginHandler, lch *LoginCompleteHandler) {
	for _, prefix := range oidcPrefixes {
		s.Handle(prefix, stdhttp.Handler(provider))
	}
	s.Handle("/login", lh)
	s.Handle("/login/complete", lch)
}
