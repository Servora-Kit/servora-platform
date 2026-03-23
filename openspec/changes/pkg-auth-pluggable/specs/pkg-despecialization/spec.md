## REMOVED Requirements

### Requirement: ScopeFromHeaders middleware is configurable
**Reason**: `pkg/transport/server/middleware/scope.go` has been deleted as it had no external callers. Scope header parsing will be handled by service-specific middleware or the future `pkg/authn/header/` engine in Phase 4.
**Migration**: Services needing scope-from-header functionality should define their own middleware or wait for the `HeaderAuthenticator` engine.

## MODIFIED Requirements

### Requirement: Authz middleware supports multi-actor-type principal construction

`pkg/authz` middleware SHALL dynamically construct the OpenFGA principal string based on `actor.Type()` and `actor.ID()`, using the pattern `string(a.Type()) + ":" + a.ID()`. It SHALL NOT hardcode `"user:"` prefix.

This logic SHALL reside in `pkg/authz/authz.go` (middleware layer), not in the `Authorizer` engine. The engine receives the fully constructed subject string.

#### Scenario: User actor principal

- **WHEN** a request from a user actor with Type `"user"` and ID `"alice"` is authorized
- **THEN** the middleware SHALL construct principal `"user:alice"` and pass it to `authorizer.IsAuthorized`

#### Scenario: Service actor principal

- **WHEN** a request from a service actor with Type `"service"` and ID `"order-svc"` is authorized
- **THEN** the middleware SHALL construct principal `"service:order-svc"` and pass it to `authorizer.IsAuthorized`

### Requirement: Authz middleware allows configurable non-checkable actor types

`pkg/authz` middleware SHALL NOT hardcode which actor types are rejected. By default, `anonymous` actors SHALL be rejected (no identity), but `user` and `service` actors SHALL both be allowed through to the `authorizer.IsAuthorized` call.

#### Scenario: Service actor passes authz check

- **WHEN** a service actor with ID `"order-svc"` makes a request to a CHECK operation
- **AND** the `Authorizer.IsAuthorized` returns `true`
- **THEN** the middleware SHALL allow the request

#### Scenario: Anonymous actor is rejected

- **WHEN** an anonymous actor makes a request to a CHECK operation
- **THEN** the middleware SHALL return 403 AUTHZ_DENIED without calling the `Authorizer`

### Requirement: Authz default object ID is configurable

`pkg/authz` middleware SHALL use `"default"` as the fallback object ID when `IDField` is empty, but SHALL allow overriding this via `WithDefaultObjectID(id string)` option.

#### Scenario: Default fallback ID

- **WHEN** a rule has empty `IDField` and no `WithDefaultObjectID` is set
- **THEN** the object ID SHALL be `"default"`

#### Scenario: Custom fallback ID

- **WHEN** `WithDefaultObjectID("singleton")` is set
- **AND** a rule has empty `IDField`
- **THEN** the object ID SHALL be `"singleton"`
