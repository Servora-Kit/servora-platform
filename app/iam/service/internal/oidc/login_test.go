package oidc

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/log"

	confpb "github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
)

func newTestLoginHandler() *LoginHandler {
	app := &confpb.App{
		Oidc: &confpb.App_Oidc{
			LoginBaseUrl: "http://localhost:3000",
		},
	}
	return NewLoginHandler(app, nil, log.DefaultLogger)
}

func TestLoginHandler_GET_MissingAuthRequestID(t *testing.T) {
	h := newTestLoginHandler()

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "missing authRequestID") {
		t.Fatalf("expected body to contain 'missing authRequestID', got %q", rec.Body.String())
	}
}

func TestLoginHandler_GET_RedirectToSPA(t *testing.T) {
	h := newTestLoginHandler()

	req := httptest.NewRequest(http.MethodGet, "/login?authRequestID=test-123", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected status %d (redirect), got %d", http.StatusFound, rec.Code)
	}
	location := rec.Header().Get("Location")
	if !strings.Contains(location, "authRequestID") {
		t.Fatalf("expected Location to contain authRequestID, got %q", location)
	}
	if !strings.Contains(location, "test-123") {
		t.Fatalf("expected Location to contain 'test-123', got %q", location)
	}
}

func TestLoginCompleteHandler_BadJSON(t *testing.T) {
	lh := newTestLoginHandler()
	h := NewLoginCompleteHandler(lh, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/login/complete", strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestLoginHandler_MethodNotAllowed(t *testing.T) {
	h := newTestLoginHandler()

	req := httptest.NewRequest(http.MethodDelete, "/login", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestLoginCompleteHandler_MethodNotAllowed(t *testing.T) {
	lh := newTestLoginHandler()
	h := NewLoginCompleteHandler(lh, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/login/complete", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestLoginCompleteHandler_MissingFields(t *testing.T) {
	t.Skip("requires Redis for full authenticate flow")
}

func TestLoginHandler_POST_AuthenticateFlow(t *testing.T) {
	t.Skip("requires Redis for full authenticate flow")
}
