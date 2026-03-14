package data

import (
	"fmt"
	"slices"
	"time"

	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"

	"github.com/Servora-Kit/servora/app/iam/service/internal/biz/entity"
)

type oidcClient struct {
	app     *entity.Application
	devMode bool
}

func newOIDCClient(app *entity.Application, devMode bool) *oidcClient {
	return &oidcClient{app: app, devMode: devMode}
}

func (c *oidcClient) GetID() string                                                       { return c.app.ClientID }
func (c *oidcClient) RedirectURIs() []string                                              { return c.app.RedirectURIs }
func (c *oidcClient) PostLogoutRedirectURIs() []string                                    { return nil }
func (c *oidcClient) LoginURL(id string) string                                           { return fmt.Sprintf("/login?authRequestID=%s", id) }
func (c *oidcClient) IDTokenLifetime() time.Duration                                      { return c.app.IDTokenLifetime }
func (c *oidcClient) DevMode() bool                                                       { return c.devMode }
func (c *oidcClient) IDTokenUserinfoClaimsAssertion() bool                                { return false }
func (c *oidcClient) ClockSkew() time.Duration                                            { return 0 }
func (c *oidcClient) RestrictAdditionalIdTokenScopes() func(scopes []string) []string     { return nil }
func (c *oidcClient) RestrictAdditionalAccessTokenScopes() func(scopes []string) []string { return nil }

func (c *oidcClient) ApplicationType() op.ApplicationType {
	switch c.app.ApplicationType {
	case "native":
		return op.ApplicationTypeNative
	case "user_agent":
		return op.ApplicationTypeUserAgent
	default:
		return op.ApplicationTypeWeb
	}
}

func (c *oidcClient) AuthMethod() oidc.AuthMethod {
	return oidc.AuthMethodBasic
}

func (c *oidcClient) ResponseTypes() []oidc.ResponseType {
	return []oidc.ResponseType{oidc.ResponseTypeCode}
}

func (c *oidcClient) GrantTypes() []oidc.GrantType {
	types := make([]oidc.GrantType, 0, len(c.app.GrantTypes))
	for _, gt := range c.app.GrantTypes {
		types = append(types, oidc.GrantType(gt))
	}
	return types
}

func (c *oidcClient) AccessTokenType() op.AccessTokenType {
	if c.app.AccessTokenType == "opaque" {
		return op.AccessTokenTypeBearer
	}
	return op.AccessTokenTypeJWT
}

func (c *oidcClient) IsScopeAllowed(scope string) bool {
	return slices.Contains(c.app.Scopes, scope)
}
