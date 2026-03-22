## Why

当前仓库的 proto `package` 缺少统一的顶层命名空间，公共 proto（如 `audit.v1`、`mapper.v1`、`pagination.v1`、`conf.v1`）与业务 proto（如 `iam.service.v1`、`authn.service.v1`）的命名风格也不一致。随着 Servora 继续走向 proto-first、代码生成与框架生态化，这会放大命名冲突、目录治理与多语言生成的认知成本，因此需要在进入更多公共能力扩展前统一收敛。

本变更对应主设计文档 `docs/plans/2026-03-20-keycloak-openfga-audit-design.md` 的框架演进约束补充，不改变业务语义，只治理 proto 命名与目录规范，为后续各阶段持续演进提供稳定基础。

## What Changes

- 统一所有 proto `package` 的顶层命名空间为 `servora.`。
- 统一保留版本后缀；公共 proto、业务 proto、注解 proto 均继续使用 `v1` 命名。
- 目录结构与 proto `package` 严格对齐，以满足 Buf `PACKAGE_DIRECTORY_MATCH` lint。
- 对现有 `.proto` 文件的 `package`、目录层级、`go_package` 与相关引用进行一致性迁移。
- **BREAKING**：变更所有受影响 proto 的 protobuf full name、生成代码包路径与导入路径。
- 补充 proto 包命名规范，明确 `service` 层仅在目录语义存在时保留，且不再允许无顶层命名空间的新 proto 进入仓库。

## Non-goals

- 不修改任何业务 RPC 语义、消息字段语义或审计/授权运行时逻辑。
- 不在本变更中引入新的 proto 能力、注解能力或代码生成器功能。
- 不调整 Buf workspace 模块边界，只做命名与目录治理。
- 不处理版本升级到 `v2` 的兼容策略；本次统一收敛到 `servora.*.v1`。

## Capabilities

### New Capabilities
- `proto-package-governance`: 定义 Servora proto 包命名、目录映射、版本后缀与生成路径一致性的统一规范。

### Modified Capabilities
- `audit-proto`: 审计 proto 的 package namespace、目录与生成路径命名规则发生变更。
- `config-proto-extension`: 配置 proto 的 package namespace、目录与生成路径命名规则发生变更。

## Impact

- 受影响目录：`api/protos/`、`app/iam/service/api/protos/`、`app/sayhello/service/api/protos/`
- 受影响生成产物：`api/gen/go/**`、可能的 `api/gen/ts/**`
- 受影响工具链：Buf lint / breaking、Go 代码生成、可能的 TS 代码生成与业务导入路径
- 受影响引用面：Go imports、proto imports、生成代码中的 package 名称与 protobuf full name
