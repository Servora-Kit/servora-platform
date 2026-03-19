package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"

	pkgauthn "github.com/Servora-Kit/servora/pkg/authn"
	jwtpkg "github.com/Servora-Kit/servora/pkg/jwt"
)

// UserClaimsMapper is re-exported from pkg/authn for backward compatibility.
type UserClaimsMapper = pkgauthn.UserClaimsMapper

// AuthnOption is re-exported from pkg/authn for backward compatibility.
type AuthnOption = pkgauthn.Option

// WithVerifier sets the JWT verifier on the Authn middleware.
func WithVerifier(v *jwtpkg.Verifier) AuthnOption { return pkgauthn.WithVerifier(v) }

// WithClaimsMapper sets a custom claims-to-actor mapper on the Authn middleware.
func WithClaimsMapper(m UserClaimsMapper) AuthnOption { return pkgauthn.WithClaimsMapper(m) }

// WithAuthnErrorHandler sets a custom error handler on the Authn middleware.
func WithAuthnErrorHandler(h func(ctx context.Context, err error) error) AuthnOption {
	return pkgauthn.WithErrorHandler(h)
}

// Authn delegates to pkg/authn.Authn, injecting an actor into the request context.
func Authn(opts ...AuthnOption) middleware.Middleware { return pkgauthn.Authn(opts...) }
