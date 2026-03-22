# 设计文档：Servora 接入 Keycloak 后的认证、授权、审计与框架演进

**日期：** 2026-03-20
**最后更新：** 2026-03-22
**状态：** Phase 1–2a 已完成 · Phase 2b 可启动

---

## 进度总览

| 阶段 | 名称 | 状态 | OpenSpec |
|------|------|------|---------|
| Phase 1 | 框架骨架 (framework-audit-skeleton) | ✅ 已完成 | `openspec/changes/archive/2026-03-20-framework-audit-skeleton/` |
| Phase 2a | 审计 emit 接入（pkg 层） | ✅ 已完成 | `openspec/changes/archive/2026-03-22-audit-emit-integration/` |
| Phase 2b | Audit Service + ClickHouse | 🔜 可启动 | — |
| Phase 3 | Keycloak 接入 | 📋 规划中 | — |
| Phase 4 | all-in-proto 代码生成 | 📋 规划中 | — |
| Phase 5 | Servora 生态扩展 | 📋 规划中 | — |

**已沉淀的框架级 specs（9 个）：** `openspec/specs/` 下的 actor-v2、audit-proto、audit-runtime、broker-abstraction、config-proto-extension、identity-header-enhancement、infra-kafka-clickhouse、logger-refactor、proto-package-governance。

**Phase 2a 新增 specs（3 个）：** `openspec/specs/` 下的 openfga-framework-api、authz-audit-emit、openfga-audit-emit。

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

> 详细设计、spec 与实现索引见 `openspec/changes/archive/2026-03-20-framework-audit-skeleton/`。

| 交付物 | 关键决策 |
|--------|---------|
| pkg/logger v2 ⚡ | 暴力重构：`New(app)` / `For(l,"mod")` / `With(l,"mod")` / `Zap()` / `Sync()` |
| Actor v2 ⚡ | 扩展为完整身份模型（Subject/ClientID/Realm/Roles/Scopes/Attrs），新增 ServiceActor |
| IdentityFromHeader v2 | 8 种 gateway header → Actor v2，支持 WithHeaderMapping |
| audit.proto + annotations.proto | AuditEvent、4 typed detail、AuditRule method option |
| conf.proto 扩展 | Kafka（含 SASL）、ClickHouse、Audit 配置 |
| pkg/broker + kafka | **franz-go**（非 sarama）；Broker/Event/Subscriber/MiddlewareFunc 接口；参考 kratos-transport |
| pkg/audit 骨架 | Emitter → Recorder → Middleware；3 种 emitter（Noop/Log/Broker） |
| 基础设施 | Kafka KRaft + ClickHouse；IAM/sayhello 从工具链移除 |

### 实现约束收敛（Phase 1 + 2a 合并）

以下约束在 Phase 1–2a 实现过程中确立，后续阶段必须遵循：

1. **Optional-init 模式统一**：所有可选基础设施组件（Kafka、ClickHouse、OpenFGA）使用 `NewXxxOptional` 函数，nil 配置返回 nil 而非 panic，调用方 nil-check 后使用
2. **Proto 集中配置**：所有框架级配置（Kafka/ClickHouse/Audit）通过 `api/protos/servora/conf/v1/conf.proto` 统一管理，不做分散的 Go config struct
3. **Logger 桥接模式**：第三方库（franz-go kzap、GORM、Ent）通过 `logger.Zap()` 获取底层 `*zap.Logger`，不直接传递 Kratos `log.Logger`
4. **Module 命名规范**：`logger.For(l, "module")` 中 module 使用 `domain/layer/service` 格式（如 `"user/biz/iam"`），不带 `-service` 后缀
5. **broker 接口扩展点**：新增 broker 实现（NATS、RabbitMQ 等）只需实现 `broker.Broker` interface，不需修改 `pkg/broker` 核心
6. **OpenSpec 主 spec 格式**：必须包含 `## Purpose` section、`## Requirements`（非 `ADDED Requirements`）、每条 requirement 第一行含 SHALL/MUST、至少一个 `#### Scenario`
7. **Proto 包治理规范**：新增或迁移后的 proto 必须使用 `servora.*` package、目录需与 package 命名空间对齐、`go_package` 必须落到 `api/gen/go/servora/**`，对应主 spec 为 `openspec/specs/proto-package-governance/spec.md`
8. **pkg 框架包去特化原则**：`pkg/` 下的框架包不得包含任何业务特化逻辑（如硬编码 `"user:"` 前缀、硬编码业务 model 的 computed relations）。业务特定配置通过 functional option 由调用方注入
9. **ClientOption 模式**：`pkg/openfga.NewClient(cfg, opts...)` 接受 `ClientOption`（`WithAuditRecorder`、`WithComputedRelations`）；`NewClientOptional` 透传 opts。服务层通过 wrapper 函数注入特定 options（如 IAM 的 `NewOpenFGAClient`），再注册到 Wire ProviderSet
10. **core/public 分层模式**：涉及 cross-cutting concern（audit、metrics、tracing）的方法，拆为 unexported core 方法（纯操作）+ 导出 wrapper（组合 cross-cutting 逻辑）。后续新增 cross-cutting 只修改 wrapper，不碰 core
11. **Kafka 双 listener**：docker-compose 中 Kafka 配置 PLAINTEXT (9092, 容器间) + EXTERNAL (29092, 宿主机)，确保开发环境测试与容器间通信均可用

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

## 7. 审计架构（骨架 + authz/openfga 主链已实现，消费端待建）

### 7.1 总体架构

```text
业务服务本地产生审计事件 → pkg/broker (Kafka) → Audit Service → ClickHouse → 查询 API
```

### 7.2 四类事件来源

| 事件类型 | 锚点 | 状态 | 说明 |
|----------|------|------|------|
| `authn.result` | `pkg/authn` / identity adapter | proto 已定义 | Phase 3 接入（随 Keycloak 改造） |
| `authz.decision` | `pkg/authz.Authz` middleware | ✅ Phase 2a 已接入 | `WithAuditRecorder` 直接注入，含 CacheHit |
| `tuple.changed` | `pkg/openfga` tuple write/delete | ✅ Phase 2a 已接入 | 方法内置自动 emit，core/public 分层 |
| `resource.mutation` | 业务服务 handler | proto 已定义 + middleware 骨架 | Phase 4 通过 annotation 自动化 |

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
| `pkg/authz` | ✅ 审计已接入 | `WithAuditRecorder` 直接注入，Check 后自动 emit `authz.decision` |
| `pkg/audit` | ✅ 主链已接入 | Recorder + LogEmitter/BrokerEmitter；e2e Kafka round-trip 已验证 |
| `pkg/broker` | ✅ 接口 + kafka 实现 | franz-go，kzap + kotel |
| `pkg/logger` | ✅ v2 | 暴力重构后的简洁 API |
| `pkg/openfga` | ✅ 框架化完成 | ClientOption 模式、API 去特化、core/public 分层、tuple audit emit |
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
| IAM/sayhello 工具链入口 | 从 Makefile MICROSERVICES/GO_WORKSPACE_MODULES、docker-compose.dev.yaml 移除 | ✅ Phase 1 |
| IAM/sayhello 源代码 | 保留作为新服务参考模板，可独立编译 | ✅ 保留 |
| `pkg/actor` | v2 破坏性升级 | ✅ Phase 1 |
| `pkg/logger` | 暴力重构 | ✅ Phase 1 |
| `pkg/transport/.../identity` | v2 多 header 支持 | ✅ Phase 1 |
| `pkg/openfga` | API 框架化 + ClientOption + core/public 分层 + audit emit | ✅ Phase 2a |
| `pkg/authz` | `WithAuditRecorder` + `authz.decision` emit + CachedCheck 适配 | ✅ Phase 2a |
| `pkg/audit` | 主链接入 + e2e 验证（LogEmitter + BrokerEmitter Kafka） | ✅ Phase 2a |
| `app/iam/service` | 适配 openfga 去特化 API + iamComputedRelations | ✅ Phase 2a |
| Kafka docker-compose | 新增 EXTERNAL listener (port 29092) 支持宿主机连接 | ✅ Phase 2a |

### 9.2 待执行（Phase 2+）

| 组件 | 操作 | 阶段 |
|------|------|------|
| IAM issuer 能力（JWKS/OIDC/登录/注册） | 下线，认证交给 Keycloak | Phase 3 |
| Traefik → IAM /v1/auth/verify 链路 | 改为网关直接对接 Keycloak | Phase 3 |
| `pkg/authn` | 降级为身份适配层（gateway header mode + direct JWT mode） | Phase 3 |
| `app/audit/service` | 新建：Kafka 消费 → ClickHouse 落库 → 查询 API | Phase 2b |
| `cmd/protoc-gen-servora-audit` | 新建：审计注解代码生成器 | Phase 4 |

---

## 10. 分阶段演进计划

### Phase 1：框架骨架 ✅ 已完成

> 交付物见 Section 4 表格。

### Phase 2a：审计 emit 接入（pkg 层） ✅ 已完成

> 详细设计、spec 与实现索引见 `openspec/changes/archive/2026-03-22-audit-emit-integration/`。

| 交付物 | 关键决策 |
|--------|---------|
| pkg/openfga API 框架化 ⚡ | `Check`/`ListObjects`/`CachedCheck` 参数从 `userID` 改为 `user`（完整 principal），移除 `"user:"` 硬编码；`parseTupleComponents` 通用化 |
| pkg/openfga ClientOption 模式 | `NewClient(cfg, opts...)` + `WithAuditRecorder` + `WithComputedRelations`；`NewClientOptional` 透传 |
| pkg/openfga core/public 分层 | `WriteTuples`/`DeleteTuples` 拆为 `writeTuplesCore`/`deleteTuplesCore` + public wrapper，成功后自动 emit `tuple.changed` |
| pkg/openfga CachedCheck 扩展 | 返回值 `(bool, error)` → `(bool, bool, error)`，新增 `cacheHit` |
| pkg/openfga 缓存层去特化 | `affectedRelations` 硬编码移除，改为 `Client.computedRelations`（通过 `WithComputedRelations` 注入）；`InvalidateForTuples` 改为 `Client` 方法 |
| pkg/authz audit 集成 | `WithAuditRecorder(r)` option + Check 后自动 emit `authz.decision`（含 CacheHit，allowed/denied/error 三种 decision） |
| app/iam/service 适配 | 全局适配 openfga 去特化 API；`NewOpenFGAClient` wrapper 注入 `iamComputedRelations` |
| Kafka EXTERNAL listener | docker-compose 新增 port 29092 供宿主机连接 |
| e2e 验证 | `pkg/audit/e2e_test.go`：LogEmitter JSON 输出 + BrokerEmitter Kafka round-trip（含 proto 反序列化） |

### Phase 2b：Audit Service + ClickHouse（可启动）

**目标：** 新建审计微服务，完成 Kafka → ClickHouse → 查询 API 的完整消费链路。

**前置条件：** Phase 2a ✅ + Kafka EXTERNAL listener ✅ + audit event proto ✅

**设计决策：**

1. **服务结构：参考 `app/iam/service` 分层**
   ```
   app/audit/service/
   ├── cmd/server/          # 服务入口
   ├── internal/
   │   ├── biz/             # 业务逻辑（事件消费、查询）
   │   ├── data/            # 数据层（ClickHouse 读写）
   │   ├── server/          # gRPC + HTTP server
   │   └── service/         # Kratos service 实现
   ├── api/protos/          # 私有查询 API proto
   ├── configs/             # 服务配置
   ├── Makefile
   └── go.mod
   ```

2. **ClickHouse 客户端：官方 native driver**
   使用 `github.com/ClickHouse/clickhouse-go/v2`，native 协议（非 database/sql），配合 `PrepareBatch` + `Append` + `Send` 进行高效批量写入。

3. **查询 API proto 路径：服务私有**
   放在 `app/audit/service/api/protos/servora/audit/service/v1/`，与 IAM 服务模式一致。
   注意区分：
   - `api/protos/servora/audit/v1/` — **共享** audit event/annotation proto（Phase 1 已完成）
   - `app/audit/service/api/protos/servora/audit/service/v1/` — audit **微服务**的私有查询 API proto

**核心任务：**
1. ClickHouse schema 设计（audit_events 表、分区策略、TTL、初始化脚本）
2. 新建 `app/audit/service` 服务脚手架（cmd/internal/configs/api）
3. 实现 Kafka consumer：从审计 topic 消费 → 反序列化 → 校验
4. 实现 ClickHouse writer：批量写入 audit_events 表
5. 实现查询 API：基础的按时间/事件类型/actor/service 筛选查询
6. Proto 定义：audit service 查询 API（gRPC + HTTP）
7. docker-compose.dev 集成：audit service 加入开发环境
8. 端到端验证：业务请求 → authz → Kafka → audit service → ClickHouse → 查询 API 可见

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
