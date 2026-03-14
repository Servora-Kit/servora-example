package oidc

import "github.com/google/wire"

// ProviderSet provides OIDC provider and login handlers for HTTP registration.
var ProviderSet = wire.NewSet(NewProvider, NewLoginHandler, NewLoginCompleteHandler)
