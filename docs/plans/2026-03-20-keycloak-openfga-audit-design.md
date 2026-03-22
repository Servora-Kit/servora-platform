# 设计文档：Servora 接入 Keycloak 后的认证、授权、审计与框架演进

**日期：** 2026-03-20
**最后更新：** 2026-03-20
**状态：** Phase 1 已完成 · Phase 2 规划中

---

## 进度总览

| 阶段 | 名称 | 状态 | OpenSpec |
|------|------|------|---------|
| Phase 1 | 框架骨架 (framework-audit-skeleton) | ✅ 已完成 | `openspec/changes/archive/2026-03-20-framework-audit-skeleton/` |
| Phase 2 | 审计主链 + authz 集成 | 🔜 待启动 | — |
| Phase 3 | Keycloak 接入 | 📋 规划中 | — |
| Phase 4 | all-in-proto 代码生成 | 📋 规划中 | — |
| Phase 5 | Servora 生态扩展 | 📋 规划中 | — |

**已沉淀的框架级 specs（8 个）：** `openspec/specs/` 下的 actor-v2、audit-proto、audit-runtime、broker-abstraction、config-proto-extension、identity-header-enhancement、infra-kafka-clickhouse、logger-refactor。

---

## 1. 背景与目标

Servora 当前仓库同时承载了框架能力（`pkg/`、`cmd/`、`api/`）与一个早期自建 IAM 服务实现。现阶段已经明确：

- 未来希望将 **Servora 打造成面向微服务快速开发的脚手架与框架生态**；
- `pkg/` 中的能力会逐步框架化、通用化，并最终作为 **Servora 生态 Go 包** 对外发布；
- 认证引入 **Keycloak**，授权继续采用 **OpenFGA**，审计采用 **Kafka + ClickHouse**；
- 当前 IAM 和 sayhello 服务已从工具链（Makefile、docker-compose.dev）中移除，保留代码作为未来新服务的参考模板。

---

## 2. 核心决策（不变）

| 决策点 | 结论 |
|---|---|
| 认证中心 | 使用 **Keycloak** |
| 网关认证策略 | 由 **Traefik / Gateway 统一验 token** |
| 业务服务是否重复验 JWT | 默认 **不重复验**，优先信任网关注入的 principal header |
| 授权底座 | 继续使用 **OpenFGA** |
| 授权执行位置 | **各业务服务本地** 接入 `pkg/authz` |
| 是否保留中央 IAM/AuthZ 在线代理 | **不保留**；最多保留薄的管理/后台能力 |
| 审计架构 | **中心化 Audit Service + 非中心化 authz/audit emit** |
| 审计总线 | 先支持 **Kafka**（franz-go），后续框架化支持更多 broker |
| actor 模型 | **通用 principal 模型**，不直接镜像 Keycloak claims |
| 审计规则配置方式 | 采用 **all-in-proto + 注解 + 代码生成 + middleware** |
| broker / transport 演进方向 | 在 Servora 内部建设自有 `pkg` 生态，参考外部项目但不以其为核心依赖 |

---

## 3. 职责分工（不变）

### 3.1 Keycloak
负责用户认证、OIDC/OAuth2 标准流程、token 签发、JWKS/discovery、client/realm/role 管理。
不负责业务资源级授权、OpenFGA 关系建模、审计存储。

### 3.2 网关（Traefik）
负责统一入口、对接 Keycloak、验证 token、将 principal 注入上游请求头、粗粒度入口控制。
不负责细粒度授权判断、业务资源审计。

### 3.3 各业务服务
从 gateway header 构建 actor → 本地 `pkg/authz` → 直接调用 OpenFGA → 产出审计事件。

### 3.4 OpenFGA
关系模型存储、Check/ListObjects/tuple write/delete。

### 3.5 Audit Service（待建）
消费审计 topic → 校验反序列化 → 落库（ClickHouse） → 提供查询统计能力。

---

## 4. Phase 1 已完成：框架骨架

> 详细设计与 spec 见 `openspec/changes/archive/2026-03-20-framework-audit-skeleton/`。
> 以下仅记录关键决策和实现索引。

### 4.1 pkg/logger 重构 ⚡ 破坏性变更

**决策：** 原 API 过于繁琐（Config struct + NewLogger + Sync 字段 + 冗长 helper 创建），全面重构为调用方友好的简洁 API。

**新 API：**
- `New(app *conf.App)` → nil-safe 构造函数，直接读 proto config
- `For(l, "module")` → 一行创建带 module 的 Helper
- `With(l, "module")` → 字符串简写，兼容 Option 风格
- `Zap()` getter → 暴露 `*zap.Logger`（供 kzap、GORM bridge 等使用）
- `Sync()` → 从字段改为方法

**迁移范围：** bootstrap、app/iam、app/sayhello、pkg/redis、pkg/openfga、pkg/transport、pkg/jwks、pkg/governance。Module 命名去掉 `-service` 后缀。

### 4.2 Actor v2 ⚡ 破坏性变更

**决策：** Actor 从纯 ID/Type/DisplayName 扩展为完整身份模型，支持多种 principal 来源。

**新增：** Subject、ClientID、Realm、Email、Roles、Scopes、Attrs 方法；`TypeService` 常量；`ServiceActor` struct；`UserActorParams` 构造模式。

### 4.3 IdentityFromHeader v2

**决策：** 支持从网关注入的多个 header 构建 Actor v2。

**支持的 header：** X-User-ID、X-Subject、X-Client-ID、X-Realm、X-Email、X-Roles、X-Scopes、X-Principal-Type。X-Principal-Type=service 构造 ServiceActor。WithHeaderMapping option 支持自定义映射。

### 4.4 Proto 定义

- `api/protos/servora/audit/v1/audit.proto`：AuditEvent、4 种 typed detail（Authn/Authz/TupleMutation/ResourceMutation）
- `api/protos/servora/audit/v1/annotations.proto`：AuditRule message + audit_rule method option
- `api/protos/servora/conf/v1/conf.proto`：Data 新增 Kafka（含 SASL）、ClickHouse 配置；App 新增 Audit 配置

### 4.5 pkg/broker — 消息代理抽象

**选型决策：** Kafka 实现使用 **franz-go**（而非 sarama）。原因：更现代的 API、原生支持 KRaft、内置 kzap/kotel 插件、更好的性能。

> 参考源：`/Users/horonlee/projects/go/kratos-transport`（broker 接口设计）、`/Users/horonlee/projects/go/Kemate`（docker-compose 配置模式）。
> **注意：** Kemate 实际使用 sarama，仅参考其基础设施配置和 optional-init 模式，不参考其 Go Kafka 库选择。

**接口设计（参考 kratos-transport 并增强）：**
- `Broker` interface：Connect / Disconnect / Publish / Subscribe
- `Event` interface：Message + **RawMessage**（底层原始消息）+ **Error**（fetch 级错误）+ Ack / Nack
- `MiddlewareFunc` + `Chain()`：handler 中间件链
- `Subscriber`：Topic + **Options**() + Unsubscribe(**removeFromManager**)

**初始化模式：** `NewBrokerOptional(ctx, cfg, logger)` — nil-safe，未配置时返回 nil + Info 日志，创建/连接失败时 warn + nil（不 panic）。遵循 `openfga.NewClientOptional` 模式。

### 4.6 pkg/audit — 审计运行时骨架

**架构：** Emitter interface（Emit/Close）→ 3 种实现（NoopEmitter、LogEmitter、BrokerEmitter）→ Recorder（自动填充 EventID/OccurredAt/TraceID）→ Kratos Middleware 骨架。

**初始化模式：** `NewRecorderOptional(cfg, broker, logger)` 根据 `App.Audit.EmitterType` 选择 emitter。

### 4.7 基础设施

- **Kafka：** apache/kafka KRaft 模式（无 ZooKeeper），固定 KAFKA_CLUSTER_ID，CONTROLLER_QUORUM_VOTERS=1@localhost:9093，health check retries=20/start_period=20s
- **ClickHouse：** clickhouse-server:25.1-alpine，端口 18123(HTTP)/19000(Native)
- **IAM/sayhello：** 从 docker-compose.dev.yaml 和 Makefile（MICROSERVICES、GO_WORKSPACE_MODULES）中移除，代码保留作为参考

### 4.8 实现约束收敛

以下约束在 Phase 1 实现过程中确立，后续阶段必须遵循：

1. **Optional-init 模式统一**：所有可选基础设施组件（Kafka、ClickHouse、OpenFGA）使用 `NewXxxOptional` 函数，nil 配置返回 nil 而非 panic，调用方 nil-check 后使用
2. **Proto 集中配置**：所有框架级配置（Kafka/ClickHouse/Audit）通过 `api/protos/servora/conf/v1/conf.proto` 统一管理，不做分散的 Go config struct
3. **Logger 桥接模式**：第三方库（franz-go kzap、GORM、Ent）通过 `logger.Zap()` 获取底层 `*zap.Logger`，不直接传递 Kratos `log.Logger`
4. **Module 命名规范**：`logger.For(l, "module")` 中 module 使用 `domain/layer/service` 格式（如 `"user/biz/iam"`），不带 `-service` 后缀
5. **broker 接口扩展点**：新增 broker 实现（NATS、RabbitMQ 等）只需实现 `broker.Broker` interface，不需修改 `pkg/broker` 核心
6. **OpenSpec 主 spec 格式**：必须包含 `## Purpose` section、`## Requirements`（非 `ADDED Requirements`）、每条 requirement 第一行含 SHALL/MUST、至少一个 `#### Scenario`

---

## 5. 不保留中央 IAM/AuthZ 在线代理（不变）

核心理由保持不变：
1. `pkg/authz` 已具备通用执行能力，无需再套代理
2. OpenFGA 自身已是独立基础设施
3. 减少网络跳数与故障面
4. 授权决策本地执行更容易获取业务上下文

允许保留薄中心能力：OpenFGA model/store 管理、后台 tuple 管理、审计查询、运维控制台。

---

## 6. actor 模型（Phase 1 已实现）

actor 不直接等于 Keycloak claims。采用：

```text
Keycloak claims / gateway headers → adapter → actor.Actor
```

actor 字段：Type（user|service|anonymous|system）、ID、Subject、ClientID、Realm、DisplayName、Email、Roles、Scopes、Attrs。

`pkg/authz`、`pkg/audit`、业务服务只依赖 actor，不依赖 Keycloak 原始 claims 结构。

Keycloak 主集成方式：OIDC discovery、JWKS、token/introspection/userinfo、Admin REST（非 gRPC）。

---

## 7. 审计架构（骨架已实现，主链待接入）

### 7.1 总体架构

```text
业务服务本地产生审计事件 → pkg/broker (Kafka) → Audit Service → ClickHouse → 查询 API
```

### 7.2 四类事件来源

| 事件类型 | 锚点 | Phase 1 状态 | Phase 2 计划 |
|----------|------|-------------|-------------|
| `authn.result` | `pkg/authn` / identity adapter | proto 已定义 | 接入 emit |
| `authz.decision` | `pkg/authz.Authz` middleware | proto 已定义 | **P0 优先接入** |
| `authz.tuple.changed` | `pkg/openfga` tuple write/delete | proto 已定义 | P1 接入 |
| `resource.mutation` | 业务服务 handler | proto 已定义 + middleware 骨架 | 通过 annotation 自动化 |

### 7.3 all-in-proto 路线

```text
proto 注解 → protoc-gen-servora-audit → middleware 自动执行
```

结构：
- `api/protos/servora/audit/v1/audit.proto` — ✅ 已定义
- `api/protos/servora/audit/v1/annotations.proto` — ✅ 已定义（AuditRule + audit_rule method option）
- `cmd/protoc-gen-servora-audit` — Phase 4 实现
- `pkg/audit` runtime — ✅ 骨架已实现

---

## 8. 框架化演进方向

### 8.1 pkg 生态当前状态

| 包 | 状态 | 说明 |
|----|------|------|
| `pkg/actor` | ✅ v2 | 通用 principal 模型，4 种 actor type |
| `pkg/authn` | 🔄 待降级 | Phase 3 改造为身份适配层 |
| `pkg/authz` | ✅ 可用 | Phase 2 接入审计 emit |
| `pkg/audit` | ✅ 骨架 | Phase 2 接入 authz 主链 |
| `pkg/broker` | ✅ 接口 + kafka 实现 | franz-go，kzap + kotel |
| `pkg/logger` | ✅ v2 | 暴力重构后的简洁 API |
| `pkg/openfga` | ✅ 可用 | Phase 2 tuple 审计接入 |
| `pkg/transport` | ✅ 可用 | IdentityFromHeader v2 已升级 |

### 8.2 关于参考项目

| 项目 | 本地路径 | 参考内容 | 不参考内容 |
|------|---------|---------|-----------|
| kratos-transport | `/Users/horonlee/projects/go/kratos-transport` | broker 接口设计、Event/Subscriber/Handler 类型签名、option 组织、middleware 模式 | 整套外部抽象边界、直接作为依赖 |
| Kemate | `/Users/horonlee/projects/go/Kemate` | docker-compose 配置（Kafka KRaft）、optional-init 模式 | sarama 选型、Kafka Go 库代码 |

### 8.3 目录边界

- `pkg/transport`：请求/响应型能力（HTTP/gRPC/SSE/WebSocket），middleware，metadata 透传
- `pkg/broker`：消息型、事件型能力，broker interface，producer/consumer lifecycle
- `pkg/task` 或 `pkg/queue`（未来）：任务队列（Asynq 等），不强塞进 broker

---

## 9. 该删什么、留什么、换什么

### 9.1 已执行的变更

| 组件 | 操作 | 状态 |
|------|------|------|
| IAM/sayhello 工具链入口 | 从 Makefile MICROSERVICES/GO_WORKSPACE_MODULES、docker-compose.dev.yaml 移除 | ✅ 已完成 |
| IAM/sayhello 源代码 | 保留作为新服务参考模板，可独立编译 | ✅ 保留 |
| `pkg/actor` | v2 破坏性升级 | ✅ 已完成 |
| `pkg/logger` | 暴力重构 | ✅ 已完成 |
| `pkg/transport/.../identity` | v2 多 header 支持 | ✅ 已完成 |

### 9.2 待执行（Phase 2+）

| 组件 | 操作 | 阶段 |
|------|------|------|
| IAM issuer 能力（JWKS/OIDC/登录/注册） | 下线，认证交给 Keycloak | Phase 3 |
| Traefik → IAM /v1/auth/verify 链路 | 改为网关直接对接 Keycloak | Phase 3 |
| `pkg/authn` | 降级为身份适配层（gateway header mode + direct JWT mode） | Phase 3 |
| `pkg/authz` | 接入审计 emit（authz.decision 事件） | Phase 2 |
| `pkg/openfga` | tuple 变更审计接入 | Phase 2 |
| `app/audit/service` | 新建：消费 Kafka → ClickHouse 落库 → 查询 API | Phase 2 |
| `cmd/protoc-gen-servora-audit` | 新建：审计注解代码生成器 | Phase 4 |

---

## 10. 分阶段演进计划

### Phase 1：框架骨架 ✅ 已完成

**交付物：**
1. ~~actor v2~~ ✅
2. ~~pkg/logger 重构~~ ✅
3. ~~audit.proto + annotations.proto~~ ✅
4. ~~conf.proto 扩展（Kafka/ClickHouse/Audit 配置）~~ ✅
5. ~~pkg/broker 接口 + kafka 实现~~ ✅
6. ~~pkg/audit 骨架~~ ✅
7. ~~IdentityFromHeader v2~~ ✅
8. ~~Docker Compose Kafka + ClickHouse~~ ✅
9. ~~IAM/sayhello 工具链解耦~~ ✅

### Phase 2：审计主链 + authz 集成（待启动）

**目标：** 将审计骨架连接成可运行的端到端审计链路。

**核心任务：**
1. `pkg/authz` middleware 接入 `pkg/audit.Recorder.RecordAuthzDecision`，每次 Check 产出 `authz.decision` 事件
2. `pkg/openfga` 的 WriteTuples/DeleteTuples 接入 `RecordTupleChange`，产出 `authz.tuple.changed` 事件
3. 新建 `app/audit/service`：
   - 从 Kafka 消费审计事件
   - 反序列化 + 校验
   - 写入 ClickHouse
   - 提供基础查询 API（gRPC + HTTP）
4. ClickHouse schema 设计（audit_events 表、分区策略、TTL）
5. 端到端验证：业务请求 → authz 判定 → audit event → Kafka → audit service → ClickHouse → 查询可见

**前置条件：** Phase 1 ✅ + 基础设施（Kafka、ClickHouse）运行正常 ✅

### Phase 3：Keycloak 接入

**目标：** 完成认证链路切换，下线自建 IAM issuer 能力。

**核心任务：**
1. 部署 Keycloak（docker-compose 新增 keycloak 服务）
2. 配置 Traefik 对接 Keycloak（ForwardAuth 或 OIDC middleware）
3. `pkg/authn` 降级重构：
   - Gateway identity mode（默认）：从 header 构造 actor
   - Direct JWT verification mode：极少数绕过网关的场景
4. 清理 IAM 中的 issuer/verify/JWKS/OIDC/登录注册能力
5. 前端对接 Keycloak 登录流程

### Phase 4：all-in-proto 代码生成

**目标：** 审计走向声明式，减少手写 emit 逻辑。

**核心任务：**
1. 实现 `cmd/protoc-gen-servora-audit`
2. 生成 operation → audit rule map
3. 生成字段提取 helper + detail builder
4. middleware 自动按 proto 规则执行审计
5. 集成到 `make api` 生成链路（`buf.audit.gen.yaml`）

### Phase 5：Servora 生态扩展

**目标：** 框架能力泛化，为对外发布做准备。

**方向：**
1. `pkg/broker` 补更多实现（NATS / RabbitMQ / Redis Streams）
2. 设计 `pkg/task` / `pkg/queue`（Asynq 等任务队列）
3. 统一框架级 observability、eventbus、identity、audit、authz 能力
4. 将 Servora 逐步沉淀为对外发布的微服务框架生态

---

## 11. 最终结论

本次设计的核心不是"替换一个认证服务"，而是为 Servora 确立一套长期有效的边界：

- 认证交给 **Keycloak**
- 网关负责 **统一认证与 principal 注入**
- 授权由 **各业务服务本地执行 `pkg/authz` + OpenFGA**
- 审计采用 **本地 emit + Kafka + 中心 Audit Service**
- actor 设计为 **通用 principal 模型**
- 审计与授权逐步走向 **all-in-proto + 注解 + 代码生成 + middleware**
- broker / transport / audit / authz / actor 构成 **Servora 的 pkg 框架生态**

Servora 未来围绕明确的基础设施边界、清晰的框架能力分层、通用的 proto 驱动与代码生成能力、面向微服务脚手架的长期 pkg 生态持续演进。
