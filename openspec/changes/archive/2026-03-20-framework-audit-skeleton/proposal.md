## Why

Servora 正在从"包含自建 IAM 的单体框架"转向"面向微服务的脚手架生态"。根据 [master design doc](../../../docs/plans/2026-03-20-keycloak-openfga-audit-design.md) **阶段 1（定骨架）**，需要在引入 Keycloak、Kafka、Audit Service 等具体实现之前，先将框架核心抽象（actor、broker、audit）的骨架与接口定义确立下来。当前代码库缺少 `pkg/broker`、`pkg/audit`、审计相关 proto 定义，且 `pkg/actor` 接口过于精简无法承载 Keycloak claims 映射与多来源身份模型——这些都是后续阶段（审计管线、Keycloak 接入、all-in-proto 审计注解）的前置依赖。

## What Changes

- **BREAKING** — `pkg/logger`: 暴力重构日志包——移除冗余 `Config` struct 改为直接接收 `*conf.App`；移除 `Sync` 字段改为方法；新增 `For(l, module)` 一行创建模块 helper 替代 `NewHelper(l, WithModule("..."))` 冗长写法；新增 `Zap()` getter 暴露底层 zap 实例；修复 nil config panic；消除 prod/default 重复代码；全量更新所有调用方
- **BREAKING** — `pkg/actor`: 扩展 `Actor` interface 和 `UserActor` struct，新增 `Subject`、`ClientID`、`Realm`、`Email`、`Roles`、`Scopes`、`Attrs` 等字段，设计为通用 principal 模型；同步新增 `ServiceActor` 类型；现有调用方需适配新接口签名
- 新增 `pkg/broker`：定义最小消息代理抽象接口（`Broker`、`Publisher`、`Subscriber`、`Message`、`Event`），参考 [kratos-transport](/Users/horonlee/projects/go/kratos-transport) broker 接口风格，但为 Servora 自有生态设计
- 新增 `pkg/broker/kafka`：基于 franz-go 的 Kafka 实现，参考 [Kemate](/Users/horonlee/projects/go/Kemate) `pkg/kafka` 的 producer/consumer 模式
- 新增 `pkg/audit`：审计运行时骨架，包含 `AuditEvent` 模型、`Emitter` 接口、`Recorder` 运行时、审计 middleware
- 新增 `api/protos/servora/audit/v1/audit.proto`：审计事件公共模型 proto（稳定骨架字段 + typed detail）
- 新增 `api/protos/servora/audit/v1/annotations.proto`：审计注解 proto 定义（RPC 级审计规则声明）
- 增强 `pkg/transport/server/middleware/identity.go`：`IdentityFromHeader` 支持读取多个 gateway 注入 header（`X-User-ID`、`X-Client-ID`、`X-Principal-Type`、`X-Realm`、`X-Roles`、`X-Scopes`、`X-Email`、`X-Subject`）
- 基础设施：docker-compose 引入 Kafka（KRaft 模式）+ ClickHouse，参照 [Kemate](/Users/horonlee/projects/go/Kemate) compose 配置
- 工具链：从 Makefile / docker-compose.dev.yaml 中移除 IAM 和 sayhello 的启动与开发条目（服务代码保留作为参考标准）

## Non-goals

- 本阶段不实现 `protoc-gen-servora-audit` 代码生成器（阶段 4）
- 本阶段不创建 `app/audit/service` 审计消费服务（阶段 2）
- 本阶段不接入 Keycloak 或改造 Traefik 认证链路（阶段 3）
- 不删除 `app/iam/service` 或 `app/sayhello/service` 的代码，仅从工具链中解耦
- 不实现 `pkg/authz` 的审计事件采集（阶段 2）
- 不设计 `pkg/task` / `pkg/queue` 任务队列抽象（阶段 5）

## Capabilities

### New Capabilities

- `logger-refactor`: Logger 包暴力重构——简化创建 API、新增 `For` 快捷方法、`Zap()` getter、移除冗余 Config struct、修复 bug、全量迁移调用方
- `actor-v2`: Actor 通用 principal 模型的破坏性升级——扩展接口、新增 ServiceActor、适配所有现有调用方
- `broker-abstraction`: 消息代理抽象层（`pkg/broker`）接口定义与 Kafka 实现
- `audit-runtime`: 审计运行时骨架（`pkg/audit`），包含事件模型、emitter 接口、recorder、middleware
- `audit-proto`: 审计相关 proto 定义——公共事件模型与 RPC 级审计注解
- `identity-header-enhancement`: IdentityFromHeader middleware 增强，支持多 gateway header 映射
- `infra-kafka-clickhouse`: 基础设施引入 Kafka 与 ClickHouse
- `config-proto-extension`: 扩展 `conf.proto` 配置体系，新增 Kafka、ClickHouse、Audit 配置 message；`pkg/logger` 暴露底层 zap 实例以支持 franz-go kzap/kotel 插件集成

### Modified Capabilities

（无现有 spec 需要修改）

## Impact

- **Breaking changes**: `pkg/logger` API 重构（`NewLogger` → `New`、移除 `Config` struct、`Sync` 字段 → 方法），影响 `pkg/bootstrap` 及所有使用 `NewHelper(l, WithModule(...))` 的站点；`pkg/actor.Actor` 接口变更，所有实现该接口的类型和所有调用 `NewUserActor` 的站点需要适配
- **Affected packages**: `pkg/actor`、`pkg/authn`、`pkg/authz`、`pkg/transport/server/middleware`
- **New dependencies**: franz-go（Kafka client）、审计相关 proto 生成代码
- **Build system**: `make gen` / `make api` 需要处理新的 `audit/v1` proto module
- **Infrastructure**: docker-compose.yaml 新增 kafka + clickhouse 服务
- **Toolchain**: Makefile 和 docker-compose.dev.yaml 移除 iam/sayhello 开发条目
