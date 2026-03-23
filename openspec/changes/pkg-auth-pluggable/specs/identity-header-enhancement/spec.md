## MODIFIED Requirements

### Requirement: IdentityFromHeader reads multiple gateway headers

This capability SHALL be implemented as a `pkg/authn/header/` engine implementing the `Authenticator` interface in Phase 4. The middleware SHALL support reading the following headers and mapping them to `actor.Actor` fields:
- `X-User-ID` → `Actor.ID()`
- `X-Subject` → `Actor.Subject()`
- `X-Client-ID` → `Actor.ClientID()`
- `X-Principal-Type` → `Actor.Type()`
- `X-Realm` → `Actor.Realm()`
- `X-Email` → `Actor.Email()`
- `X-Roles` → `Actor.Roles()` (comma-separated string → `[]string`)
- `X-Scopes` → `Actor.Scopes()` (space-separated string → `[]string`)

The engine SHALL implement `authn.Authenticator` interface:

```go
func (a *HeaderAuthenticator) Authenticate(ctx context.Context) (actor.Actor, error)
```

#### Scenario: All headers present

- **WHEN** a request arrives with headers `X-User-ID: u1`, `X-Subject: sub-1`, `X-Client-ID: web-app`, `X-Realm: production`, `X-Email: a@b.com`, `X-Roles: admin,viewer`, `X-Scopes: openid profile`
- **THEN** the `HeaderAuthenticator` SHALL return an actor with ID()="u1", Subject()="sub-1", ClientID()="web-app", Realm()="production", Email()="a@b.com", Roles()=["admin","viewer"], Scopes()=["openid","profile"]

#### Scenario: Only X-User-ID present (backward compatible)

- **WHEN** a request arrives with only `X-User-ID: u1` and no other identity headers
- **THEN** the `HeaderAuthenticator` SHALL return an actor with ID()="u1" and all other identity fields as zero values

#### Scenario: No identity headers

- **WHEN** a request arrives with no identity headers at all
- **THEN** the `HeaderAuthenticator` SHALL return `actor.NewAnonymousActor()`

### Requirement: Header key mapping is configurable

The `HeaderAuthenticator` SHALL accept `WithHeaderMapping(mapping HeaderMapping)` option to override the default header key → Actor field mapping.

#### Scenario: Custom header key for user ID

- **WHEN** `header.NewAuthenticator(header.WithHeaderMapping(map[string]string{"id": "X-Custom-User"}))` is configured
- **AND** a request arrives with `X-Custom-User: custom-1`
- **THEN** the resulting actor SHALL have ID()="custom-1"

### Requirement: X-Principal-Type determines actor type

The `HeaderAuthenticator` SHALL branch on `X-Principal-Type` header value:
- `"service"` → construct `ServiceActor`
- `"user"` or absent/empty → construct `UserActor` if `X-User-ID` is present, otherwise `AnonymousActor`

#### Scenario: Service principal from header

- **WHEN** a request arrives with `X-Principal-Type: service`, `X-User-ID: order-svc`, `X-Client-ID: order-client`
- **THEN** the `HeaderAuthenticator` SHALL return a `ServiceActor` with Type()="service", ID()="order-svc", ClientID()="order-client"
