package data

import (
	"fmt"
	"slices"
	"time"

	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"

	apppb "github.com/Servora-Kit/servora/api/gen/go/servora/application/service/v1"
)

type oidcClient struct {
	app     *apppb.Application
	devMode bool
}

func newOIDCClient(app *apppb.Application, devMode bool) *oidcClient {
	return &oidcClient{app: app, devMode: devMode}
}

func (c *oidcClient) GetID() string                                                       { return c.app.ClientId }
func (c *oidcClient) RedirectURIs() []string                                              { return c.app.RedirectUris }
func (c *oidcClient) PostLogoutRedirectURIs() []string                                    { return nil }
func (c *oidcClient) LoginURL(id string) string                                           { return fmt.Sprintf("/login?authRequestID=%s", id) }
func (c *oidcClient) IDTokenLifetime() time.Duration                                      { return time.Duration(c.app.IdTokenLifetime) * time.Second }
func (c *oidcClient) DevMode() bool                                                       { return c.devMode }
func (c *oidcClient) IDTokenUserinfoClaimsAssertion() bool                                { return false }
func (c *oidcClient) ClockSkew() time.Duration                                            { return 0 }
func (c *oidcClient) RestrictAdditionalIdTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string { return scopes }
}
func (c *oidcClient) RestrictAdditionalAccessTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string { return scopes }
}

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
