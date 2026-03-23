## Context

`pkg/authn` 和 `pkg/authz` 是 Servora 框架中处理认证和授权的核心包。当前两者均为单体实现：

- `pkg/authn`：直接内嵌 JWT 验证逻辑和 Keycloak 特有的 claims 映射，中间件函数 `Authn()` 即是唯一实现
- `pkg/authz`：直接依赖 `*openfga.Client`、`*audit.Recorder`、`*redis.Client`，并导入 IAM 服务 proto 的 `AuthzMode` 枚举

参考项目 `tx7do/kratos-authn` 和 `tx7do/kratos-authz` 采用了 `interface + engine/` 的可插拔架构，但其 authz 的 `AuthClaims` 前置注入模式和 RBAC 风格接口不适合 Servora 的 ReBAC（OpenFGA）+ proto-driven rules 模式。

## Goals / Non-Goals

**Goals:**
- 将 `pkg/authn` 和 `pkg/authz` 重构为接口驱动架构
- 引擎实现放在子目录中，支持未来扩展
- `AuthzMode` 枚举从 IAM 服务 proto 移至共享 proto
- 审计发射从 authz 引擎解耦为中间件回调
- 所有现有服务（IAM、sayhello、audit）适配新 API

**Non-Goals:**
- 新增认证/授权引擎（OIDC、Casbin、OPA 等）
- 实现网关 header 模式认证（Phase 4 内容）
- 变更 proto 注解格式或 codegen 插件逻辑（仅更新 import path）

## Decisions

### D1: 接口设计——Servora 风格而非 tx7do 风格

**决策**: 采用精简的 Servora 特定接口，而非照搬 tx7do 的宽泛接口。

**Authenticator 接口**:

```go
// pkg/authn/authn.go
type Authenticator interface {
    Authenticate(ctx context.Context) (actor.Actor, error)
}
```

**Authorizer 接口**:

```go
// pkg/authz/authz.go
type Authorizer interface {
    IsAuthorized(ctx context.Context, subject, relation, objectType, objectID string) (allowed bool, err error)
}
```

**Why**:
- tx7do 的 `Authenticator` 包含 `CreateIdentity` / `CreateIdentityWithContext` 等签发功能，Servora 不需要（Keycloak 负责签发）
- tx7do 的 `Authorizer` 接口（`Subjects/Action/Resource/Projects`）面向 RBAC/Chef 风格，不适合 ReBAC（OpenFGA 用 `subject/relation/objectType:objectID`）
- Servora 的 `Authenticate` 直接返回 `actor.Actor`，与框架 actor model 无缝衔接

**Alternatives considered**:
- 照搬 tx7do 接口 → 太宽泛，需要大量适配层
- 不用接口、只用 functional options → 当前方案，不可插拔

### D2: 子目录结构（engine 模式）

**决策**: 每个引擎实现放在独立子目录中。

```
pkg/authn/
  authn.go              → Authenticator 接口 + Server() 中间件 + Option 类型
  jwt/
    jwt.go              → JWTAuthenticator 实现
    claims.go           → ClaimsMapper 类型 + 内置 mappers（DefaultClaimsMapper, KeycloakClaimsMapper）
    options.go          → JWT 引擎的 Option
  noop/
    noop.go             → NoopAuthenticator（返回 anonymous）
  authn_test.go         → 中间件层测试

pkg/authz/
  authz.go              → Authorizer 接口 + AuthzRule + AuthzMode + Server() 中间件 + Option 类型
  openfga/
    openfga.go          → OpenFGAAuthorizer 实现（封装 pkg/openfga.Client + 可选 Redis 缓存）
    options.go          → OpenFGA 引擎的 Option
  noop/
    noop.go             → NoopAuthorizer（总是放行）
  authz_test.go         → 中间件层测试
```

**Why**: 用户明确要求子目录以支持未来多适配器扩展。

**Alternatives considered**:
- 扁平文件（`jwt.go` 直接在 `pkg/authn/`） → 可行但用户偏好子目录
- `engine/jwt/` 深嵌套 → 过度设计

### D3: AuthzMode 定义位置

**决策**: `AuthzMode` 枚举移至 `api/protos/servora/authz/v1/authz.proto`（共享 proto），并同时在 `pkg/authz/authz.go` 中定义 Go 常量作为引用。

```
api/protos/servora/authz/v1/
  authz.proto           → AuthzMode enum + AuthzRule message
```

`protoc-gen-servora-authz` 生成的代码将引用 `api/gen/go/servora/authz/v1` 而非 `api/gen/go/servora/authz/service/v1`。

**Why**: 
- `AuthzMode` 是框架级概念，不应存在于 IAM 服务 proto 中
- 共享 proto 确保所有服务使用统一的枚举定义
- 生成代码可以直接引用共享 proto，保持 type safety

**Alternatives considered**:
- 纯 Go 常量（不用 proto） → 丢失跨语言一致性和 proto 生态集成
- 保留在 IAM proto → 框架包反向依赖服务，违反依赖倒置

### D4: 审计从 authz 引擎解耦

**决策**: 审计发射从 `authzConfig` 内部移至中间件层的回调函数。

```go
// pkg/authz/authz.go

// DecisionDetail 描述一次授权判定的详情。
type DecisionDetail struct {
    Operation  string
    Subject    string
    Relation   string
    ObjectType string
    ObjectID   string
    Allowed    bool
    Err        error
}

// WithDecisionLogger 设置授权判定回调，在每次 Check 后调用。
// 取代原 WithAuditRecorder，使中间件不再依赖 pkg/audit。
func WithDecisionLogger(fn func(ctx context.Context, detail DecisionDetail)) Option
```

服务层通过闭包桥接到 `audit.Recorder`：

```go
authz.Server(authorizer,
    authz.WithDecisionLogger(func(ctx context.Context, d authz.DecisionDetail) {
        recorder.RecordAuthzDecision(ctx, d.Operation, actor, audit.AuthzDetail{...})
    }),
)
```

**Why**: 
- `pkg/authz`（中间件层）不再 import `pkg/audit`
- 审计是可选关注点，不应成为授权引擎的内在依赖
- 回调模式比 interface 更轻量（只有一个函数签名）

**Alternatives considered**:
- 定义 `DecisionLogger` interface → 过度设计，只有一个方法
- 保持 `WithAuditRecorder` → 维持 authz→audit 耦合

### D5: 中间件函数命名

**决策**: 中间件函数统一命名为 `Server()`，与 tx7do 和 Kratos 生态保持一致。

```go
// 使用方式
authn.Server(jwtAuth, authn.WithErrorHandler(...))
authz.Server(fgaAuth, authz.WithRulesFunc(rules), authz.WithDecisionLogger(logger))
```

**Why**: `Server()` 是 Kratos 中间件的惯用命名（`selector.Server()`、`recovery.Server()`），语义清晰。

**Alternatives considered**:
- 保持 `Authn()` / `Authz()` → 不符合 Kratos 惯例，且与包名重复

### D6: OpenFGA Authorizer 内部封装 Redis 缓存

**决策**: Redis 缓存是 OpenFGA 引擎的内部关注点，不暴露给中间件层。

```go
// pkg/authz/openfga/openfga.go
func NewAuthorizer(fgaClient *pkgfga.Client, opts ...Option) authz.Authorizer

// Option
func WithRedisCache(rdb *redis.Client, ttl time.Duration) Option
```

中间件层的 `Authorizer` 接口不含缓存概念，`IsAuthorized` 返回的 `(bool, error)` 不包含 `cacheHit`。

`cacheHit` 信息通过 `DecisionLogger` 回调的扩展字段传递（如果需要）：

```go
type DecisionDetail struct {
    // ...
    CacheHit bool  // 由引擎设置，中间件透传给 DecisionLogger
}
```

**Why**: 
- 缓存是实现细节，不同引擎（OPA、Casbin）不一定有缓存概念
- 保持接口精简

### D7: JWT 引擎的 ClaimsMapper 体系

**决策**: `ClaimsMapper` 是 JWT 引擎的配置项。提供两个内置 mapper：

```go
// pkg/authn/jwt/claims.go

// ClaimsMapper 从 JWT MapClaims 转换为 actor.Actor。
type ClaimsMapper func(claims jwtv5.MapClaims) (actor.Actor, error)

// DefaultClaimsMapper 映射标准 OIDC claims（sub, name, email, azp, scope）。
// 不含任何 IdP 特有字段。
func DefaultClaimsMapper() ClaimsMapper

// KeycloakClaimsMapper 扩展 DefaultClaimsMapper，额外映射 Keycloak 特有字段
//（iss→Realm, realm_access.roles 等）。
func KeycloakClaimsMapper() ClaimsMapper
```

**Why**: 将 IdP 特有逻辑从默认路径移出，框架默认行为不绑定任何特定 IdP。

## Breaking Changes & Migration

### pkg/authn

| Before | After | Migration |
|--------|-------|-----------|
| `authn.Authn(authn.WithVerifier(v))` | `authn.Server(jwt.NewAuthenticator(jwt.WithVerifier(v)))` | 构造 JWT 引擎，传入 Server() |
| `authn.WithClaimsMapper(fn)` | `jwt.WithClaimsMapper(fn)` | ClaimsMapper 移到 JWT 引擎 Option |
| 默认含 Keycloak `iss→Realm` 映射 | 默认只映射标准 OIDC claims | 需要 Keycloak 映射时显式用 `jwt.KeycloakClaimsMapper()` |

### pkg/authz

| Before | After | Migration |
|--------|-------|-----------|
| `authz.Authz(authz.WithFGAClient(c))` | `authz.Server(openfga.NewAuthorizer(c))` | 构造 OpenFGA 引擎，传入 Server() |
| `authz.WithAuthzCache(rdb, ttl)` | `openfga.WithRedisCache(rdb, ttl)` | 缓存配置移到 OpenFGA 引擎 Option |
| `authz.WithAuditRecorder(r)` | `authz.WithDecisionLogger(fn)` | 改用回调函数，服务层桥接 audit.Recorder |
| `authzpb.AuthzMode_AUTHZ_MODE_CHECK` | `authzpb.AuthzMode_AUTHZ_MODE_CHECK`（新 import path） | 更新 import 为 `servora/authz/v1` |

### Proto

| Before | After |
|--------|-------|
| `app/iam/service/api/protos/servora/authz/service/v1/authz.proto` | `api/protos/servora/authz/v1/authz.proto` |
| `import authzpb "...servora/authz/service/v1"` | `import authzpb "...servora/authz/v1"` |

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| **Breaking change 范围大**：authn/authz 是几乎所有服务的中间件 | 服务代码量不大（IAM deprecated, sayhello 示例, audit 新建），一次性迁移可控 |
| **Proto 移动可能影响生成代码引用链** | 先移 proto + regenerate，再改 Go 代码；分步提交 |
| **DecisionLogger 回调丢失类型安全**（相比 WithAuditRecorder） | DecisionDetail 结构体提供足够字段；服务层闭包保持类型安全 |
| **子目录增加 import 深度**（`pkg/authn/jwt` vs `pkg/authn`） | Go 惯例中 2 层嵌套完全可接受；与 Kratos 生态一致 |

## Open Questions

- `pkg/authz/authz.go` 中的 `AuthzRule` 结构体是否也应该移到共享 proto message？当前是纯 Go struct，由 codegen 引用。保持 Go struct 更简洁，但 `AuthzMode` 字段需引用共享 proto 枚举。
