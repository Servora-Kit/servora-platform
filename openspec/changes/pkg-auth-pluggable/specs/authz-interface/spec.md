## ADDED Requirements

### Requirement: Authorizer interface defines the authz contract

`pkg/authz` SHALL define an `Authorizer` interface:

```go
type Authorizer interface {
    IsAuthorized(ctx context.Context, subject, relation, objectType, objectID string) (bool, error)
}
```

The interface takes pre-resolved parameters (subject string already constructed by the middleware) and returns a boolean decision. Caching, protocol details, and backend specifics are the engine's internal concern.

#### Scenario: OpenFGA authorizer implements interface

- **WHEN** `pkg/authz/openfga.NewAuthorizer(fgaClient, opts...)` is called
- **THEN** it SHALL return a value implementing `authz.Authorizer`

#### Scenario: Noop authorizer implements interface

- **WHEN** `pkg/authz/noop.NewAuthorizer()` is called
- **THEN** it SHALL return a value implementing `authz.Authorizer`

### Requirement: Server middleware accepts Authorizer interface

`pkg/authz` SHALL provide a `Server(authorizer Authorizer, opts ...Option) middleware.Middleware` function. The middleware SHALL:

1. Extract operation from `transport.Transporter`
2. Look up `AuthzRule` for the operation
3. If `AuthzMode` is `NONE`, pass through
4. Resolve actor from context, reject anonymous
5. Construct subject as `string(actor.Type()) + ":" + actor.ID()`
6. Resolve object type and ID from the rule and request
7. Call `authorizer.IsAuthorized(ctx, subject, relation, objectType, objectID)`
8. Invoke `DecisionLogger` callback (if configured) with the result
9. Allow or deny based on the returned boolean

#### Scenario: Middleware delegates to authorizer

- **WHEN** `Server(fgaAuth, WithRulesFunc(rules))` middleware is applied
- **AND** a request matches a CHECK-mode rule
- **THEN** the middleware SHALL call `fgaAuth.IsAuthorized(ctx, "user:abc", "admin", "platform", "default")`

#### Scenario: No transport in context passes through

- **WHEN** `Server(authorizer)` middleware is applied
- **AND** the context has no `transport.Transporter`
- **THEN** the middleware SHALL call the handler directly

#### Scenario: No rule for operation is fail-closed

- **WHEN** a request operation has no matching rule
- **THEN** the middleware SHALL return 403 `AUTHZ_NO_RULE`

#### Scenario: Nil authorizer returns 503

- **WHEN** `Server(nil)` middleware is applied
- **AND** a request matches a CHECK-mode rule with a valid actor
- **THEN** the middleware SHALL return 503 `AUTHZ_UNAVAILABLE`

### Requirement: DecisionLogger replaces WithAuditRecorder

`pkg/authz` SHALL provide `WithDecisionLogger(fn func(ctx context.Context, detail DecisionDetail))` option. `DecisionDetail` SHALL contain: `Operation`, `Subject`, `Relation`, `ObjectType`, `ObjectID`, `Allowed` (bool), `Err` (error), and `CacheHit` (bool).

The middleware SHALL call the logger function after every authorization check (allowed, denied, or error).

#### Scenario: Decision logger called on allow

- **WHEN** a `DecisionLogger` is configured
- **AND** authorization succeeds
- **THEN** the middleware SHALL call the logger with `Allowed: true`

#### Scenario: Decision logger called on deny

- **WHEN** a `DecisionLogger` is configured
- **AND** authorization is denied
- **THEN** the middleware SHALL call the logger with `Allowed: false, Err: nil`

#### Scenario: Decision logger called on error

- **WHEN** a `DecisionLogger` is configured
- **AND** the authorizer returns an error
- **THEN** the middleware SHALL call the logger with `Allowed: false, Err: <error>`

#### Scenario: No logger configured is silent

- **WHEN** no `DecisionLogger` is configured
- **THEN** the middleware SHALL not attempt any logging callback

### Requirement: AuthzMode is defined in pkg/authz as Go type

`pkg/authz` SHALL define `AuthzMode` as a Go type within `authz.go`, aliased from the shared proto enum `api/gen/go/servora/authz/v1`. The middleware SHALL use this type for rule matching.

#### Scenario: AuthzMode references shared proto

- **WHEN** `pkg/authz/authz.go` is inspected
- **THEN** `AuthzMode` SHALL reference the enum from `servora/authz/v1` (not `servora/authz/service/v1`)

### Requirement: OpenFGA engine lives in pkg/authz/openfga/ subdirectory

The OpenFGA-based authorizer SHALL reside in `pkg/authz/openfga/`. It SHALL encapsulate `*pkgopenfga.Client` and optional `*redis.Client` for caching.

#### Scenario: OpenFGA authorizer with Redis cache

- **WHEN** `openfga.NewAuthorizer(fgaClient, openfga.WithRedisCache(rdb, 60*time.Second))` is created
- **AND** `IsAuthorized` is called
- **THEN** the engine SHALL use `fgaClient.CachedCheck` with the given Redis client and TTL

#### Scenario: OpenFGA authorizer without cache

- **WHEN** `openfga.NewAuthorizer(fgaClient)` is created without Redis option
- **AND** `IsAuthorized` is called
- **THEN** the engine SHALL call `fgaClient.CachedCheck` with nil Redis (degraded mode, no cache)

#### Scenario: OpenFGA authorizer sets CacheHit on DecisionDetail

- **WHEN** the OpenFGA engine obtains `cacheHit` from `CachedCheck`
- **THEN** it SHALL set `DecisionDetail.CacheHit` accordingly for the DecisionLogger to consume

### Requirement: Noop authorizer always allows

`pkg/authz/noop/` SHALL provide a `NewAuthorizer()` function returning an `Authorizer` that always returns `(true, nil)`.

#### Scenario: Noop always allows

- **WHEN** a `noop.Authorizer` processes any authorization request
- **THEN** it SHALL return `(true, nil)`

### Requirement: AuthzRule and resolveObject remain in pkg/authz

`AuthzRule` struct and `resolveObject` function SHALL remain in `pkg/authz/authz.go`. They are middleware-level concerns (rule matching + proto field extraction) and SHALL NOT be pushed into engine implementations.

#### Scenario: AuthzRule struct is in pkg/authz

- **WHEN** `pkg/authz/authz.go` is inspected
- **THEN** it SHALL contain the `AuthzRule` struct with fields `Mode`, `Relation`, `ObjectType`, `IDField`

#### Scenario: resolveObject is used by middleware

- **WHEN** the middleware resolves the object for an authorization check
- **THEN** it SHALL use `resolveObject(rule, req, defaultObjectID)` from `pkg/authz`
