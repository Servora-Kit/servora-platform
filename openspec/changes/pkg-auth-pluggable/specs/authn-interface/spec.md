## ADDED Requirements

### Requirement: Authenticator interface defines the authn contract

`pkg/authn` SHALL define an `Authenticator` interface:

```go
type Authenticator interface {
    Authenticate(ctx context.Context) (actor.Actor, error)
}
```

The interface SHALL accept a `context.Context` (which carries `transport.Transporter` via Kratos) and return an `actor.Actor` or an error. Implementations are responsible for extracting credentials (token, headers, etc.) from the context and converting them to an actor.

#### Scenario: JWT authenticator implements interface

- **WHEN** `pkg/authn/jwt.NewAuthenticator(opts...)` is called
- **THEN** it SHALL return a value implementing `authn.Authenticator`

#### Scenario: Noop authenticator implements interface

- **WHEN** `pkg/authn/noop.NewAuthenticator()` is called
- **THEN** it SHALL return a value implementing `authn.Authenticator`

### Requirement: Server middleware accepts Authenticator interface

`pkg/authn` SHALL provide a `Server(authenticator Authenticator, opts ...Option) middleware.Middleware` function that creates a Kratos middleware using the given authenticator.

The middleware SHALL:
1. Call `authenticator.Authenticate(ctx)` to obtain an actor
2. On success: inject the actor into context via `actor.NewContext(ctx, a)` and call the handler
3. On error: invoke the configured error handler or return the error directly

#### Scenario: Middleware injects actor on success

- **WHEN** `Server(jwtAuth)` middleware is applied
- **AND** `jwtAuth.Authenticate(ctx)` returns a valid `actor.Actor`
- **THEN** the actor SHALL be available via `actor.FromContext(ctx)` in the handler

#### Scenario: Middleware returns error on auth failure

- **WHEN** `Server(jwtAuth)` middleware is applied
- **AND** `jwtAuth.Authenticate(ctx)` returns an error
- **AND** no custom error handler is configured
- **THEN** the middleware SHALL return the error without calling the handler

#### Scenario: Custom error handler is invoked

- **WHEN** `Server(jwtAuth, WithErrorHandler(fn))` middleware is applied
- **AND** `jwtAuth.Authenticate(ctx)` returns an error
- **THEN** the middleware SHALL invoke `fn(ctx, err)` and return its result

### Requirement: JWT engine lives in pkg/authn/jwt/ subdirectory

The JWT-based authenticator implementation SHALL reside in `pkg/authn/jwt/`. It SHALL encapsulate all JWT-specific logic: token extraction from `Authorization: Bearer` header, verification via `pkg/jwt.Verifier`, and claims-to-actor mapping.

#### Scenario: JWT engine extracts bearer token

- **WHEN** a request has header `Authorization: Bearer <token>`
- **AND** a `jwt.Authenticator` processes the request
- **THEN** the engine SHALL extract `<token>` and verify it using the configured `Verifier`

#### Scenario: JWT engine with no token returns anonymous

- **WHEN** a request has no `Authorization` header
- **AND** a `jwt.Authenticator` processes the request
- **THEN** the engine SHALL return `actor.NewAnonymousActor()` without error

#### Scenario: JWT engine with nil verifier passes through

- **WHEN** `jwt.NewAuthenticator()` is created without a `Verifier`
- **AND** a request has a Bearer token
- **THEN** the engine SHALL return `actor.NewAnonymousActor()` and store the raw token in context via `svrmw.NewTokenContext`

### Requirement: ClaimsMapper is a JWT engine configuration

`ClaimsMapper` SHALL be a type `func(claims jwtv5.MapClaims) (actor.Actor, error)` defined in `pkg/authn/jwt/`. It SHALL be configurable via `jwt.WithClaimsMapper(mapper)`.

#### Scenario: Custom claims mapper

- **WHEN** `jwt.NewAuthenticator(jwt.WithClaimsMapper(customMapper))` is created
- **THEN** the engine SHALL use `customMapper` to convert JWT claims to an actor

#### Scenario: Default claims mapper maps standard OIDC claims only

- **WHEN** `jwt.NewAuthenticator()` is created without specifying a mapper
- **THEN** the default mapper SHALL map: `sub→ID`, `name→DisplayName`, `email→Email`, `azp→ClientID`, `scope→Scopes`, `roles→Roles`
- **AND** it SHALL NOT map `iss` to `Realm` (that is Keycloak-specific)

### Requirement: KeycloakClaimsMapper is provided as a built-in option

`pkg/authn/jwt/` SHALL provide a `KeycloakClaimsMapper()` function returning a `ClaimsMapper` that extends the default OIDC mapping with Keycloak-specific fields: `iss→Realm`, and `realm_access.roles` merged into `Roles`.

#### Scenario: Keycloak mapper maps issuer to realm

- **WHEN** `jwt.NewAuthenticator(jwt.WithClaimsMapper(jwt.KeycloakClaimsMapper()))` is created
- **AND** a JWT contains claim `iss: "https://keycloak.example.com/realms/production"`
- **THEN** the resulting actor SHALL have `Realm()` returning `"https://keycloak.example.com/realms/production"`

### Requirement: Noop authenticator always returns anonymous

`pkg/authn/noop/` SHALL provide a `NewAuthenticator()` function returning an `Authenticator` that always returns `actor.NewAnonymousActor()` without error.

#### Scenario: Noop always anonymous

- **WHEN** a `noop.Authenticator` processes any request
- **THEN** it SHALL return `actor.NewAnonymousActor()` and `nil` error
