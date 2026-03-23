## MODIFIED Requirements

### Requirement: Authz middleware supports audit recorder injection

`pkg/authz` SHALL provide a `WithDecisionLogger(fn func(ctx context.Context, detail DecisionDetail))` option that accepts a callback function for authorization decision logging. This replaces the previous `WithAuditRecorder(r *audit.Recorder)` option.

The middleware SHALL NOT directly depend on `pkg/audit`. Services bridge the audit system via a closure:

```go
authz.WithDecisionLogger(func(ctx context.Context, d authz.DecisionDetail) {
    recorder.RecordAuthzDecision(ctx, d.Operation, actorFromCtx, audit.AuthzDetail{...})
})
```

When no `DecisionLogger` is configured, the middleware SHALL function normally without any logging.

#### Scenario: Decision logger via closure

- **WHEN** `authz.Server(authorizer, authz.WithDecisionLogger(fn))` is called
- **THEN** the middleware SHALL invoke `fn` after each authorization check

#### Scenario: Nil decision logger is safe

- **WHEN** `authz.Server(authorizer)` is called without `WithDecisionLogger`
- **THEN** the middleware SHALL function normally without invoking any logging callback

### Requirement: Authz middleware emits authz.decision event after Check

After each authorization Check (whether allowed, denied, or errored), the middleware SHALL call the configured `DecisionLogger` with a `DecisionDetail` containing Operation, Subject, Relation, ObjectType, ObjectID, Allowed, Err, and CacheHit.

#### Scenario: Allowed check triggers logger

- **WHEN** a request passes authorization (authorizer returns `true, nil`)
- **THEN** the middleware SHALL call the `DecisionLogger` with `Allowed: true`

#### Scenario: Denied check triggers logger

- **WHEN** a request fails authorization (authorizer returns `false, nil`)
- **THEN** the middleware SHALL call the `DecisionLogger` with `Allowed: false` before returning the permission error

#### Scenario: Check error triggers logger

- **WHEN** the authorizer returns an error
- **THEN** the middleware SHALL call the `DecisionLogger` with `Allowed: false, Err: <error>`

#### Scenario: No logger skips emission

- **WHEN** the middleware has no `DecisionLogger` configured
- **AND** a request is processed
- **THEN** authorization SHALL proceed normally without any logging

## REMOVED Requirements

### Requirement: Authz middleware supports audit recorder injection (old)
**Reason**: Replaced by `WithDecisionLogger` callback pattern to decouple `pkg/authz` from `pkg/audit`
**Migration**: Replace `WithAuditRecorder(recorder)` with `WithDecisionLogger(func(ctx, detail) { recorder.RecordAuthzDecision(...) })`
