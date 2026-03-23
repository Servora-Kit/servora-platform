package middleware

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/Servora-Kit/servora/pkg/actor"
)

// ScopeBinding maps an HTTP request header to a UserActor scope key.
// Validate is an optional validator function; a non-nil error causes a 400 response.
type ScopeBinding struct {
	Header   string            // HTTP header name (e.g. "X-Tenant-ID")
	ScopeKey string            // actor scope key (e.g. "tenant_id")
	Validate func(string) error // optional validator (e.g. uuid.Parse)
}

// ScopeFromHeaders creates a Kratos middleware that reads scope values from
// configured HTTP request headers and injects them into the UserActor.
//
// Requires an authenticated UserActor in context (i.e. must run after Authn).
// Headers are optional — absent headers are silently skipped.
// If a Validate function is provided and returns an error, the request is rejected with 400.
//
// Example:
//
//	const ScopeKeyTenantID = "tenant_id"
//
//	ScopeFromHeaders(
//	    middleware.ScopeBinding{
//	        Header:   "X-Tenant-ID",
//	        ScopeKey: ScopeKeyTenantID,
//	        Validate: func(v string) error { _, err := uuid.Parse(v); return err },
//	    },
//	)
func ScopeFromHeaders(bindings ...ScopeBinding) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}

			a, ok := actor.FromContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			ua, ok := a.(*actor.UserActor)
			if !ok {
				return handler(ctx, req)
			}

			for _, b := range bindings {
				val := tr.RequestHeader().Get(b.Header)
				if val == "" {
					continue
				}
				if b.Validate != nil {
					if err := b.Validate(val); err != nil {
						return nil, errors.BadRequest(
							fmt.Sprintf("INVALID_%s", b.ScopeKey),
							fmt.Sprintf("invalid %s header: %v", b.Header, err),
						)
					}
				}
				ua.SetScope(b.ScopeKey, val)
			}

			return handler(ctx, req)
		}
	}
}
