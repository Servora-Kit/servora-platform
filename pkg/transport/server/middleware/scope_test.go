package middleware

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/uuid"

	"github.com/Servora-Kit/servora/pkg/actor"
)

const (
	testOrgUUID  = "550e8400-e29b-41d4-a716-446655440000"
	testProjUUID = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
)

const (
	scopeKeyOrg  = "organization_id"
	scopeKeyProj = "project_id"
)

var (
	orgBinding = ScopeBinding{
		Header:   "X-Organization-ID",
		ScopeKey: scopeKeyOrg,
		Validate: func(v string) error { _, err := uuid.Parse(v); return err },
	}
	projBinding = ScopeBinding{
		Header:   "X-Project-ID",
		ScopeKey: scopeKeyProj,
		Validate: func(v string) error { _, err := uuid.Parse(v); return err },
	}
)

func scopeCtx(headers map[string]string) context.Context {
	ua := actor.NewUserActor(actor.UserActorParams{ID: "user-1", DisplayName: "Test", Email: "test@example.com"})
	ctx := actor.NewContext(context.Background(), ua)
	return transport.NewServerContext(ctx, &fakeTransport{headers: headers})
}

func noopHandler(ctx context.Context, req any) (any, error) { return "ok", nil }

func TestScopeFromHeaders_BothHeaders(t *testing.T) {
	mw := ScopeFromHeaders(orgBinding, projBinding)
	handler := mw(func(ctx context.Context, req any) (any, error) {
		ua := actor.MustFromContext(ctx).(*actor.UserActor)
		if ua.Scope(scopeKeyOrg) != testOrgUUID {
			t.Errorf("org = %q, want %q", ua.Scope(scopeKeyOrg), testOrgUUID)
		}
		if ua.Scope(scopeKeyProj) != testProjUUID {
			t.Errorf("proj = %q, want %q", ua.Scope(scopeKeyProj), testProjUUID)
		}
		return nil, nil
	})

	ctx := scopeCtx(map[string]string{
		"X-Organization-ID": testOrgUUID,
		"X-Project-ID":      testProjUUID,
	})
	_, _ = handler(ctx, nil)
}

func TestScopeFromHeaders_NoHeaders_Silent(t *testing.T) {
	mw := ScopeFromHeaders(orgBinding, projBinding)
	handler := mw(func(ctx context.Context, req any) (any, error) {
		ua := actor.MustFromContext(ctx).(*actor.UserActor)
		if ua.Scope(scopeKeyOrg) != "" {
			t.Errorf("org should be empty, got %q", ua.Scope(scopeKeyOrg))
		}
		if ua.Scope(scopeKeyProj) != "" {
			t.Errorf("proj should be empty, got %q", ua.Scope(scopeKeyProj))
		}
		return nil, nil
	})

	ctx := scopeCtx(map[string]string{})
	_, _ = handler(ctx, nil)
}

func TestScopeFromHeaders_InvalidOrgUUID(t *testing.T) {
	mw := ScopeFromHeaders(orgBinding, projBinding)
	handler := mw(noopHandler)

	ctx := scopeCtx(map[string]string{"X-Organization-ID": "not-a-uuid"})
	_, err := handler(ctx, nil)

	if err == nil {
		t.Fatal("expected error for invalid org UUID")
	}
	se := new(errors.Error)
	if !errors.As(err, &se) || se.Reason != "INVALID_organization_id" {
		t.Errorf("reason = %v, want INVALID_organization_id", err)
	}
}

func TestScopeFromHeaders_InvalidProjectUUID(t *testing.T) {
	mw := ScopeFromHeaders(orgBinding, projBinding)
	handler := mw(noopHandler)

	ctx := scopeCtx(map[string]string{"X-Project-ID": "bad"})
	_, err := handler(ctx, nil)

	if err == nil {
		t.Fatal("expected error for invalid project UUID")
	}
	se := new(errors.Error)
	if !errors.As(err, &se) || se.Reason != "INVALID_project_id" {
		t.Errorf("reason = %v, want INVALID_project_id", err)
	}
}

func TestScopeFromHeaders_NoActor_Passthrough(t *testing.T) {
	mw := ScopeFromHeaders(orgBinding)
	called := false
	handler := mw(func(ctx context.Context, req any) (any, error) {
		called = true
		return nil, nil
	})

	ctx := transport.NewServerContext(context.Background(), &fakeTransport{
		headers: map[string]string{"X-Organization-ID": testOrgUUID},
	})
	_, err := handler(ctx, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestScopeFromHeaders_NoTransport_Passthrough(t *testing.T) {
	mw := ScopeFromHeaders(orgBinding)
	called := false
	handler := mw(func(ctx context.Context, req any) (any, error) {
		called = true
		return nil, nil
	})

	_, err := handler(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestScopeFromHeaders_OrgOnly(t *testing.T) {
	mw := ScopeFromHeaders(orgBinding, projBinding)
	handler := mw(func(ctx context.Context, req any) (any, error) {
		ua := actor.MustFromContext(ctx).(*actor.UserActor)
		if ua.Scope(scopeKeyOrg) != testOrgUUID {
			t.Errorf("org = %q, want %q", ua.Scope(scopeKeyOrg), testOrgUUID)
		}
		if ua.Scope(scopeKeyProj) != "" {
			t.Errorf("proj should be empty, got %q", ua.Scope(scopeKeyProj))
		}
		return nil, nil
	})

	ctx := scopeCtx(map[string]string{"X-Organization-ID": testOrgUUID})
	_, _ = handler(ctx, nil)
}

func TestScopeFromHeaders_NoBindings_Passthrough(t *testing.T) {
	mw := ScopeFromHeaders()
	called := false
	handler := mw(func(ctx context.Context, req any) (any, error) {
		called = true
		return nil, nil
	})

	ctx := scopeCtx(map[string]string{"X-Organization-ID": testOrgUUID})
	_, err := handler(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
