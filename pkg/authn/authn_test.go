package authn

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/Servora-Kit/servora/pkg/actor"
	jwtpkg "github.com/Servora-Kit/servora/pkg/jwt"
	svrmw "github.com/Servora-Kit/servora/pkg/transport/server/middleware"
)

// fakeTransport implements transport.Transporter for test purposes.
type fakeTransport struct {
	headers map[string]string
}

func (f *fakeTransport) Kind() transport.Kind           { return transport.KindHTTP }
func (f *fakeTransport) Endpoint() string               { return "" }
func (f *fakeTransport) Operation() string              { return "" }
func (f *fakeTransport) RequestHeader() transport.Header { return &fakeHeader{f.headers} }
func (f *fakeTransport) ReplyHeader() transport.Header   { return &fakeHeader{} }

type fakeHeader struct {
	m map[string]string
}

func (h *fakeHeader) Get(key string) string      { return h.m[key] }
func (h *fakeHeader) Set(key, value string)      { h.m[key] = value }
func (h *fakeHeader) Add(key, value string)      {}
func (h *fakeHeader) Keys() []string             { return nil }
func (h *fakeHeader) Values(key string) []string { return nil }

func transportCtx(headers map[string]string) context.Context {
	return transport.NewServerContext(context.Background(), &fakeTransport{headers: headers})
}

// testSigner is a minimal JWT signer for test use only (not using pkg/jwt.Signer to avoid PEM encoding).
type testSigner struct {
	key *rsa.PrivateKey
	kid string
}

func (s *testSigner) sign(claims gojwt.MapClaims) (string, error) {
	token := gojwt.NewWithClaims(gojwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.kid
	return token.SignedString(s.key)
}

// setupTest returns a signer, verifier pair for unit tests.
func setupTest(t *testing.T) (*testSigner, *jwtpkg.Verifier) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	const kid = "test-kid"
	verifier := jwtpkg.NewVerifier()
	verifier.AddKey(kid, &key.PublicKey)
	return &testSigner{key: key, kid: kid}, verifier
}

// TestAuthn_NoToken_AnonymousActor checks that a request with no Authorization header
// injects an anonymous actor and calls the handler.
func TestAuthn_NoToken_AnonymousActor(t *testing.T) {
	_, verifier := setupTest(t)
	mw := Authn(WithVerifier(verifier))

	called := false
	handler := mw(func(ctx context.Context, req any) (any, error) {
		called = true
		a, ok := actor.FromContext(ctx)
		if !ok {
			t.Fatal("expected actor in context")
		}
		if a.Type() != actor.TypeAnonymous {
			t.Errorf("expected TypeAnonymous, got %v", a.Type())
		}
		return "ok", nil
	})

	ctx := transportCtx(map[string]string{})
	resp, err := handler(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("handler was not called")
	}
	if resp != "ok" {
		t.Errorf("resp = %v, want ok", resp)
	}
}

// TestAuthn_NoTransport_AnonymousActor checks that without a transport context, an
// anonymous actor is injected.
func TestAuthn_NoTransport_AnonymousActor(t *testing.T) {
	_, verifier := setupTest(t)
	mw := Authn(WithVerifier(verifier))

	handler := mw(func(ctx context.Context, req any) (any, error) {
		a, ok := actor.FromContext(ctx)
		if !ok {
			t.Fatal("expected actor in context")
		}
		if a.Type() != actor.TypeAnonymous {
			t.Errorf("expected TypeAnonymous, got %v", a.Type())
		}
		return nil, nil
	})

	_, err := handler(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestAuthn_InvalidToken_ErrorReturned checks that a malformed token returns an error.
func TestAuthn_InvalidToken_ErrorReturned(t *testing.T) {
	_, verifier := setupTest(t)
	mw := Authn(WithVerifier(verifier))

	handler := mw(func(ctx context.Context, req any) (any, error) {
		t.Fatal("handler should not be called on invalid token")
		return nil, nil
	})

	ctx := transportCtx(map[string]string{"Authorization": "Bearer not.a.valid.jwt"})
	_, err := handler(ctx, nil)
	if err == nil {
		t.Fatal("expected error for invalid JWT")
	}
}

// TestAuthn_ValidJWT_CorrectActorInjected checks that a valid JWT produces the expected actor.
func TestAuthn_ValidJWT_CorrectActorInjected(t *testing.T) {
	signer, verifier := setupTest(t)
	mw := Authn(WithVerifier(verifier))

	tokenStr, err := signer.sign(gojwt.MapClaims{
		"sub":   "user-42",
		"name":  "Alice",
		"email": "alice@example.com",
		"role":  "admin",
	})
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	handler := mw(func(ctx context.Context, req any) (any, error) {
		a, ok := actor.FromContext(ctx)
		if !ok {
			t.Fatal("expected actor in context")
		}
		if a.Type() != actor.TypeUser {
			t.Errorf("type = %v, want TypeUser", a.Type())
		}
		if a.ID() != "user-42" {
			t.Errorf("id = %q, want user-42", a.ID())
		}
		ua, ok := a.(*actor.UserActor)
		if !ok {
			t.Fatal("actor is not *actor.UserActor")
		}
		if ua.Email() != "alice@example.com" {
			t.Errorf("email = %q, want alice@example.com", ua.Email())
		}
		if ua.Meta("role") != "admin" {
			t.Errorf("role = %q, want admin", ua.Meta("role"))
		}
		tok, hasTok := svrmw.TokenFromContext(ctx)
		if !hasTok {
			t.Fatal("expected token in context")
		}
		if tok != tokenStr {
			t.Error("token in context does not match signed token")
		}
		return "ok", nil
	})

	ctx := transportCtx(map[string]string{"Authorization": "Bearer " + tokenStr})
	_, err = handler(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestAuthn_NilVerifier_PassThrough checks that a nil verifier stores the raw token in
// context and injects an anonymous actor.
func TestAuthn_NilVerifier_PassThrough(t *testing.T) {
	mw := Authn() // no WithVerifier → verifier is nil

	const rawToken = "somerawtoken"
	handler := mw(func(ctx context.Context, req any) (any, error) {
		a, ok := actor.FromContext(ctx)
		if !ok {
			t.Fatal("expected actor in context")
		}
		if a.Type() != actor.TypeAnonymous {
			t.Errorf("expected TypeAnonymous, got %v", a.Type())
		}
		tok, hasTok := svrmw.TokenFromContext(ctx)
		if !hasTok {
			t.Fatal("expected raw token in context")
		}
		if tok != rawToken {
			t.Errorf("token = %q, want %q", tok, rawToken)
		}
		return nil, nil
	})

	ctx := transportCtx(map[string]string{"Authorization": "Bearer " + rawToken})
	_, err := handler(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestAuthn_CustomClaimsMapper_CustomActorReturned checks that a custom UserClaimsMapper
// is used to build the actor.
func TestAuthn_CustomClaimsMapper_CustomActorReturned(t *testing.T) {
	signer, verifier := setupTest(t)

	customActor := actor.NewUserActor("custom-id", "Custom", "custom@example.com", nil)
	customMapper := func(_ gojwt.MapClaims) (actor.Actor, error) {
		return customActor, nil
	}

	mw := Authn(WithVerifier(verifier), WithClaimsMapper(customMapper))

	tokenStr, err := signer.sign(gojwt.MapClaims{"sub": "ignored"})
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	handler := mw(func(ctx context.Context, req any) (any, error) {
		a, ok := actor.FromContext(ctx)
		if !ok {
			t.Fatal("expected actor in context")
		}
		if a.ID() != "custom-id" {
			t.Errorf("id = %q, want custom-id", a.ID())
		}
		return nil, nil
	})

	ctx := transportCtx(map[string]string{"Authorization": "Bearer " + tokenStr})
	_, err = handler(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestAuthn_CustomErrorHandler_CalledOnInvalidToken verifies that the custom error
// handler is invoked when token verification fails.
func TestAuthn_CustomErrorHandler_CalledOnInvalidToken(t *testing.T) {
	_, verifier := setupTest(t)

	sentinel := errors.New("custom error")
	mw := Authn(
		WithVerifier(verifier),
		WithErrorHandler(func(_ context.Context, _ error) error { return sentinel }),
	)

	handler := mw(func(ctx context.Context, req any) (any, error) {
		t.Fatal("handler should not be called")
		return nil, nil
	})

	ctx := transportCtx(map[string]string{"Authorization": "Bearer bad.token.here"})
	_, err := handler(ctx, nil)
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, want sentinel error", err)
	}
}

// TestExtractBearerToken checks the exported helper.
func TestExtractBearerToken(t *testing.T) {
	cases := []struct {
		header string
		want   string
	}{
		{"", ""},
		{"Bearer mytoken", "mytoken"},
		{"bearer mytoken", "mytoken"},
		{"BEARER mytoken", "mytoken"},
		{"Basic abc123", ""},
		{"mytoken", ""},
	}
	for _, tc := range cases {
		got := ExtractBearerToken(tc.header)
		if got != tc.want {
			t.Errorf("ExtractBearerToken(%q) = %q, want %q", tc.header, got, tc.want)
		}
	}
}
