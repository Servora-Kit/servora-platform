## Why

`pkg/authn` 和 `pkg/authz` 当前是单体实现，无接口抽象。`authn` 只支持 JWT 模式且 `defaultClaimsMapper` 含 Keycloak 特有字段（`iss→Realm`）；`authz` 直接依赖 `*openfga.Client`、`*audit.Recorder`、`*redis.Client` 和 IAM 服务 proto 枚举 `authzpb.AuthzMode`——框架包反向依赖了服务实现的 proto，违反依赖倒置原则。两个包均不可插拔，无法在不重写中间件的情况下替换认证/授权引擎。

此变更将两个包重构为**接口驱动 + 引擎子目录**架构，参考 `tx7do/kratos-authn` 和 `tx7do/kratos-authz` 的设计精华，同时保留 Servora 特有的 proto-driven rules、actor model 和审计集成。属于主设计文档 Phase 4 的前置基础设施工作（接口化），独立于 Keycloak 集成。

## Non-goals

- 不实现新的认证引擎（OIDC、pre-shared key 等），仅搭建接口和迁移现有 JWT 实现
- 不实现新的授权引擎（Casbin、OPA 等），仅搭建接口和迁移现有 OpenFGA 实现
- 不实现 Phase 4 的网关 header 模式认证（`HeaderAuthenticator`），仅预留接口
- 不变更 `pkg/actor` 接口（Phase 3 已完成 actor v2）
- 不变更 `protoc-gen-servora-authz` / `protoc-gen-servora-audit` 的 proto 注解格式

## What Changes

### pkg/authn
- **BREAKING** 引入 `Authenticator` 接口：`Authenticate(ctx context.Context) (actor.Actor, error)`
- 中间件函数签名变更为 `Server(authenticator Authenticator, opts ...Option) middleware.Middleware`
- 当前 JWT 验证逻辑迁入 `pkg/authn/jwt/` 子目录，实现 `Authenticator` 接口
- Keycloak 特有的 `defaultClaimsMapper`（`iss→Realm`）移入 JWT 引擎的 `KeycloakClaimsMapper` 选项
- 新增 `pkg/authn/noop/` 引擎（总是返回 anonymous actor）

### pkg/authz
- **BREAKING** 引入 `Authorizer` 接口：`IsAuthorized(ctx context.Context, subject, relation, objectType, objectID string) (allowed bool, err error)`
- 中间件函数签名变更为 `Server(authorizer Authorizer, opts ...Option) middleware.Middleware`
- 当前 OpenFGA 检查逻辑迁入 `pkg/authz/openfga/` 子目录，实现 `Authorizer` 接口（含可选 Redis 缓存）
- **审计发射从 authz 引擎中解耦** → 变为中间件层的可选回调 `WithDecisionLogger`
- **BREAKING** `AuthzMode` 枚举从 `app/iam/service/api/protos/` 移至共享 proto `api/protos/servora/authz/v1/`
- 新增 `pkg/authz/noop/` 引擎（总是放行）

### Proto 治理
- `AuthzMode` 枚举和 `AuthzRule` message 从 IAM 服务 proto 移至 `api/protos/servora/authz/v1/authz.proto`
- `protoc-gen-servora-authz` 更新生成代码的 import path

### 服务适配
- `app/iam/service`：更新 server 初始化以使用新的引擎构造方式
- `app/sayhello/service`：同上
- `app/audit/service`：同上（authz 中间件配置）

## Capabilities

### New Capabilities
- `authn-interface`: Authenticator 接口定义、中间件函数签名、引擎子目录架构规范
- `authz-interface`: Authorizer 接口定义、中间件函数签名、引擎子目录架构规范、审计解耦

### Modified Capabilities
- `authz-audit-emit`: 审计发射从 authz 引擎内部解耦为中间件回调模式
- `pkg-despecialization`: ScopeFromHeaders 已删除（无需修改），但 authz 部分需更新为接口化后的描述
- `identity-header-enhancement`: HeaderAuthenticator 成为 authn 引擎之一的规格（Phase 4 时实现，此处仅调整归属）

## Impact

- **Breaking API**: `pkg/authn.Authn()` → `authn.Server(authenticator, opts...)`；`pkg/authz.Authz()` → `authz.Server(authorizer, opts...)`
- **Proto move**: `AuthzMode` 枚举从 IAM 服务 proto 移至共享 proto，所有引用需更新
- **Import paths**: 服务代码中 `pkg/authn` 和 `pkg/authz` 的使用方式变更
- **Codegen**: `protoc-gen-servora-authz` 生成代码的 import path 需适配新的共享 proto 位置
- **Dependencies**: `pkg/authz/authz.go`（中间件层）将不再直接依赖 `pkg/openfga`、`pkg/audit`、`pkg/redis`——这些依赖下沉到 `pkg/authz/openfga/` 引擎
