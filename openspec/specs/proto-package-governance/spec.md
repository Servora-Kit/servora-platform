# Spec: proto-package-governance

## Purpose

Defines requirements for the `proto-package-governance` capability.

## Requirements

### Requirement: Proto packages MUST use Servora namespace and explicit version

All new and migrated proto files SHALL use a package name rooted at `servora.` and SHALL include an explicit version suffix such as `v1`. The package name SHALL reflect the semantic namespace of the proto and SHALL remain stable once published inside the repository.

#### Scenario: New shared proto declares namespaced package

- **WHEN** a new shared proto is added under the repository
- **THEN** its `package` declaration MUST start with `servora.` and end with a version segment such as `.v1`

#### Scenario: Existing short package is migrated

- **WHEN** an existing proto currently uses a short package such as `audit.v1` or `mapper.v1`
- **THEN** it MUST be migrated to a namespaced package such as `servora.audit.v1` or `servora.mapper.v1`

### Requirement: Proto directories MUST match package namespace

The repository SHALL store `.proto` files in directories that match their declared package namespace so that Buf `PACKAGE_DIRECTORY_MATCH` lint passes without exceptions.

#### Scenario: Shared proto path matches package

- **WHEN** a proto declares `package servora.audit.v1;`
- **THEN** the file MUST reside under a path ending with `servora/audit/v1/`

#### Scenario: Service proto path matches package

- **WHEN** a proto declares `package servora.iam.service.v1;`
- **THEN** the file MUST reside under a path ending with `servora/iam/service/v1/`

### Requirement: go_package MUST align with proto namespace and generated path

Each migrated proto SHALL define a `go_package` whose import path aligns with the namespaced proto directory and generated output location, while preserving a readable Go package alias.

#### Scenario: Shared proto go_package is migrated

- **WHEN** `servora.audit.v1` is generated to Go code
- **THEN** its `go_package` MUST point to a path under `api/gen/go/servora/audit/v1` with a stable alias such as `auditv1`

#### Scenario: Service proto go_package is migrated

- **WHEN** `servora.iam.service.v1` is generated to Go code
- **THEN** its `go_package` MUST point to a path under `api/gen/go/servora/iam/service/v1` with a readable alias such as `iampb`
