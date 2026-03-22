package data

import (
	"testing"

	"github.com/zitadel/oidc/v3/pkg/op"

	apppb "github.com/Servora-Kit/servora/api/gen/go/servora/application/service/v1"
)

func newTestApp() *apppb.Application {
	return &apppb.Application{
		Id:              "app-1",
		ClientId:        "client-123",
		Name:            "test-app",
		Type:            "web",
		RedirectUris:    []string{"http://localhost:8080/callback", "http://example.com/cb"},
		Scopes:          []string{"openid", "profile", "email"},
		GrantTypes:      []string{"authorization_code", "refresh_token"},
		ApplicationType: "web",
		AccessTokenType: "jwt",
		IdTokenLifetime: 300,
	}
}

func TestOIDCClient_GetID(t *testing.T) {
	app := newTestApp()
	c := newOIDCClient(app, false)
	if got := c.GetID(); got != app.ClientId {
		t.Errorf("GetID() = %q, want %q", got, app.ClientId)
	}
}

func TestOIDCClient_RedirectURIs(t *testing.T) {
	app := newTestApp()
	c := newOIDCClient(app, false)
	got := c.RedirectURIs()
	if len(got) != len(app.RedirectUris) {
		t.Fatalf("RedirectURIs() length = %d, want %d", len(got), len(app.RedirectUris))
	}
	for i, uri := range got {
		if uri != app.RedirectUris[i] {
			t.Errorf("RedirectURIs()[%d] = %q, want %q", i, uri, app.RedirectUris[i])
		}
	}
}

func TestOIDCClient_ApplicationType(t *testing.T) {
	tests := []struct {
		appType string
		want    op.ApplicationType
	}{
		{"web", op.ApplicationTypeWeb},
		{"native", op.ApplicationTypeNative},
		{"user_agent", op.ApplicationTypeUserAgent},
		{"unknown", op.ApplicationTypeWeb},
	}
	for _, tt := range tests {
		t.Run(tt.appType, func(t *testing.T) {
			app := newTestApp()
			app.ApplicationType = tt.appType
			c := newOIDCClient(app, false)
			if got := c.ApplicationType(); got != tt.want {
				t.Errorf("ApplicationType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOIDCClient_AccessTokenType(t *testing.T) {
	tests := []struct {
		tokenType string
		want      op.AccessTokenType
	}{
		{"jwt", op.AccessTokenTypeJWT},
		{"opaque", op.AccessTokenTypeBearer},
		{"", op.AccessTokenTypeJWT},
	}
	for _, tt := range tests {
		t.Run(tt.tokenType, func(t *testing.T) {
			app := newTestApp()
			app.AccessTokenType = tt.tokenType
			c := newOIDCClient(app, false)
			if got := c.AccessTokenType(); got != tt.want {
				t.Errorf("AccessTokenType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOIDCClient_IsScopeAllowed(t *testing.T) {
	app := newTestApp()
	c := newOIDCClient(app, false)

	if !c.IsScopeAllowed("openid") {
		t.Error("IsScopeAllowed(\"openid\") = false, want true")
	}
	if c.IsScopeAllowed("admin") {
		t.Error("IsScopeAllowed(\"admin\") = true, want false")
	}
}

func TestOIDCClient_LoginURL(t *testing.T) {
	c := newOIDCClient(newTestApp(), false)
	got := c.LoginURL("req-42")
	want := "/login?authRequestID=req-42"
	if got != want {
		t.Errorf("LoginURL() = %q, want %q", got, want)
	}
}

func TestOIDCClient_DevMode(t *testing.T) {
	app := newTestApp()

	c := newOIDCClient(app, true)
	if !c.DevMode() {
		t.Error("DevMode() = false, want true")
	}

	c = newOIDCClient(app, false)
	if c.DevMode() {
		t.Error("DevMode() = true, want false")
	}
}

func TestOIDCClient_IDTokenLifetime(t *testing.T) {
	app := newTestApp()
	c := newOIDCClient(app, false)
	got := c.IDTokenLifetime()
	want := 300 * 1_000_000_000 // 300 seconds in nanoseconds
	if int64(got) != int64(want) {
		t.Errorf("IDTokenLifetime() = %v, want 300s", got)
	}
}
