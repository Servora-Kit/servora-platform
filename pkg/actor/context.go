package actor

import "context"

type contextKey struct{}

func NewContext(ctx context.Context, a Actor) context.Context {
	return context.WithValue(ctx, contextKey{}, a)
}

func FromContext(ctx context.Context) (Actor, bool) {
	a, ok := ctx.Value(contextKey{}).(Actor)
	return a, ok
}

// MustFromContext panics if no actor in context — use only in trusted code paths.
func MustFromContext(ctx context.Context) Actor {
	a, ok := FromContext(ctx)
	if !ok {
		panic("actor: no actor in context")
	}
	return a
}

// ScopeFromContext returns a generic scope value from the actor in context.
// Callers define their own scope key constants (e.g. const ScopeKeyTenantID = "tenant_id").
func ScopeFromContext(ctx context.Context, key string) (string, bool) {
	a, ok := FromContext(ctx)
	if !ok {
		return "", false
	}
	val := a.Scope(key)
	if val == "" {
		return "", false
	}
	return val, true
}
