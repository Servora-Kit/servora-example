package oidc

import (
	stdhttp "net/http"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/zitadel/oidc/v3/pkg/op"
)

// oidcPrefixes lists path prefixes routed to the zitadel/oidc Provider.
// These are registered via HandlePrefix (gorilla/mux PathPrefix) so any
// sub-path (e.g. /.well-known/openid-configuration) is matched correctly.
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
		s.HandlePrefix(prefix, stdhttp.Handler(provider))
	}
	s.HandlePrefix("/login/complete", lch)
	s.HandlePrefix("/login", lh)
}
