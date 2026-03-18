package data

import (
	"testing"
	"time"

	"github.com/zitadel/oidc/v3/pkg/op"

	"github.com/Servora-Kit/servora/app/iam/service/internal/biz/entity"
)

func newTestApp() *entity.Application {
	return &entity.Application{
		ID:              "app-1",
		ClientID:        "client-123",
		Name:            "test-app",
		RedirectURIs:    []string{"http://localhost:8080/callback", "http://example.com/cb"},
		Scopes:          []string{"openid", "profile", "email"},
		GrantTypes:      []string{"authorization_code", "refresh_token"},
		ApplicationType: "web",
		AccessTokenType: "jwt",
		TenantID:        "tenant-1",
		IDTokenLifetime: 5 * time.Minute,
	}
}

func TestOIDCClient_GetID(t *testing.T) {
	app := newTestApp()
	c := newOIDCClient(app, false)
	if got := c.GetID(); got != app.ClientID {
		t.Errorf("GetID() = %q, want %q", got, app.ClientID)
	}
}

func TestOIDCClient_RedirectURIs(t *testing.T) {
	app := newTestApp()
	c := newOIDCClient(app, false)
	got := c.RedirectURIs()
	if len(got) != len(app.RedirectURIs) {
		t.Fatalf("RedirectURIs() length = %d, want %d", len(got), len(app.RedirectURIs))
	}
	for i, uri := range got {
		if uri != app.RedirectURIs[i] {
			t.Errorf("RedirectURIs()[%d] = %q, want %q", i, uri, app.RedirectURIs[i])
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
