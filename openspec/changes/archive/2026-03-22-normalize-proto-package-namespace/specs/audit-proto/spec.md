## MODIFIED Requirements

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

The proto SHALL declare `package servora.audit.v1;` and SHALL live in a directory matching `servora/audit/v1`.

#### Scenario: Proto compiles and generates Go code

- **WHEN** `make api` is run
- **THEN** `api/gen/go/servora/audit/v1/audit.pb.go` SHALL be generated without errors

#### Scenario: Proto generates TypeScript code

- **WHEN** `make api-ts` is run
- **THEN** TypeScript types for AuditEvent SHALL be generated in `api/gen/ts/`

### Requirement: Audit annotations proto defines RPC-level audit rules
`api/protos/servora/audit/v1/annotations.proto` SHALL define:
- `AuditRule` message with fields for `AuditEventType event_type`, `string target_type`, `string target_id_field`
- A method option `google.protobuf.MethodOptions` extension `audit_rule` of type `AuditRule`

This proto is for future codegen consumption (protoc-gen-servora-audit) and SHALL compile in this phase without requiring a generator.

The proto SHALL declare `package servora.audit.v1;` and SHALL live in a directory matching `servora/audit/v1`.

#### Scenario: Annotation proto compiles

- **WHEN** `make api` is run
- **THEN** `annotations.pb.go` SHALL be generated and the `audit_rule` extension SHALL be importable in Go

#### Scenario: Annotation can be applied to RPC

- **WHEN** a service proto imports `servora/audit/v1/annotations.proto` and annotates an RPC with `option (servora.audit.v1.audit_rule) = { event_type: AUDIT_EVENT_TYPE_RESOURCE_MUTATION, target_type: "project", target_id_field: "project_id" };`
- **THEN** the proto SHALL compile without errors
