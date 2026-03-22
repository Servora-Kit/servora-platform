## ADDED Requirements

### Requirement: AuditEvent proto defines stable event schema

`api/protos/servora/audit/v1/audit.proto` SHALL define an `AuditEvent` message with fields:
- `string event_id` — UUID
- `AuditEventType event_type` — enum
- `string event_version`
- `google.protobuf.Timestamp occurred_at`
- `string service`
- `string operation`
- `AuditActor actor`
- `AuditTarget target`
- `AuditResult result`
- `string trace_id`
- `string request_id`
- `oneof detail` containing `AuthnDetail`, `AuthzDetail`, `TupleMutationDetail`, `ResourceMutationDetail`

#### Scenario: Proto compiles and generates Go code

- **WHEN** `make api` is run
- **THEN** `api/gen/go/servora/audit/v1/audit.pb.go` SHALL be generated without errors

#### Scenario: Proto generates TypeScript code

- **WHEN** `make api-ts` is run
- **THEN** TypeScript types for AuditEvent SHALL be generated in `api/gen/ts/`

### Requirement: AuditEventType enum covers four event categories

`AuditEventType` enum SHALL include:
- `AUDIT_EVENT_TYPE_UNSPECIFIED = 0`
- `AUDIT_EVENT_TYPE_AUTHN_RESULT = 1`
- `AUDIT_EVENT_TYPE_AUTHZ_DECISION = 2`
- `AUDIT_EVENT_TYPE_TUPLE_CHANGED = 3`
- `AUDIT_EVENT_TYPE_RESOURCE_MUTATION = 4`

#### Scenario: Enum values are distinct

- **WHEN** the proto is compiled
- **THEN** each enum value SHALL have a unique integer and be usable in Go and TypeScript

### Requirement: AuditActor message captures actor snapshot

`AuditActor` message SHALL include:
- `string id`
- `string type` — user / service / system / anonymous
- `string display_name`
- `string email`
- `string subject`
- `string client_id`
- `string realm`

#### Scenario: AuditActor from UserActor

- **WHEN** an AuditEvent is created from a request with a UserActor
- **THEN** the `AuditActor` fields SHALL be populated from the Actor's getter methods

### Requirement: Typed detail messages for each event category

The proto SHALL define:
- `AuthnDetail` — with `string method`, `bool success`, `string failure_reason`
- `AuthzDetail` — with `string relation`, `string object_type`, `string object_id`, `AuthzDecision decision` (enum: ALLOWED/DENIED/NO_RULE/ERROR), `bool cache_hit`, `string error_reason`
- `TupleMutationDetail` — with `TupleMutationType mutation_type` (enum: WRITE/DELETE), `repeated TupleChange tuples`
- `ResourceMutationDetail` — with `ResourceMutationType mutation_type` (enum: CREATE/UPDATE/DELETE), `string resource_type`, `string resource_id`

#### Scenario: AuthzDetail captures full decision context

- **WHEN** an authz decision event is created for user "user-1" checking "viewer" on "project:proj-1" with result=allowed
- **THEN** the `AuthzDetail` SHALL contain relation="viewer", object_type="project", object_id="proj-1", decision=ALLOWED

### Requirement: Audit annotations proto defines RPC-level audit rules

`api/protos/servora/audit/v1/annotations.proto` SHALL define:
- `AuditRule` message with fields for `AuditEventType event_type`, `string target_type`, `string target_id_field`
- A method option `google.protobuf.MethodOptions` extension `audit_rule` of type `AuditRule`

This proto is for future codegen consumption (protoc-gen-servora-audit) and SHALL compile in this phase without requiring a generator.

#### Scenario: Annotation proto compiles

- **WHEN** `make api` is run
- **THEN** `annotations.pb.go` SHALL be generated and the `audit_rule` extension SHALL be importable in Go

#### Scenario: Annotation can be applied to RPC

- **WHEN** a service proto imports `servora/audit/v1/annotations.proto` and annotates an RPC with `option (servora.audit.v1.audit_rule) = { event_type: AUDIT_EVENT_TYPE_RESOURCE_MUTATION, target_type: "project", target_id_field: "project_id" };`
- **THEN** the proto SHALL compile without errors
