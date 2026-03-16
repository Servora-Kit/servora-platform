# 设计：多租户 Actor 扩展与组织作用域

## 架构概览

```
请求流：
  Authorization: Bearer <jwt>       → Authn 中间件 → UserActor(id, name, email)
  X-Organization-ID: <uuid>         → Scope 中间件 → actor.SetOrganizationID()
  X-Project-ID: <uuid>              → Scope 中间件 → actor.SetProjectID()
                                    → Authz 中间件 → OpenFGA Check
                                    → Handler → biz → data
```

## 决策

### D1: Actor 扩展 vs 独立 Viewer 接口

**选择 Actor 扩展**，而不是新建独立 Viewer 接口。

原因：
- servora 已有成熟的 `pkg/actor` 包，`UserActor` 在整个调用链中都通过 context 传递
- 新建 Viewer 会导致 context 中有两个身份/作用域对象，增加混淆
- go-wind-admin 和 Kemate 的 Viewer 模式本质上就是"带作用域的 Actor"，只是名称不同
- 扩展 `UserActor` 加 setter/getter 即可，向后兼容

### D2: 作用域来源——Header vs JWT vs URL

**选择 Header（`X-Organization-ID` / `X-Project-ID`）**。

原因：
- 用户可属于多个 Organization，切换组织不应要求重新签发 JWT
- JWT 内容不可变，而作用域是请求级别可变的
- Header 方案被 Kemate（`X-Workspace-ID`）、go-wind-admin（JWT 内 `tid`，但 servora 场景不适合）验证有效
- URL 路径参数在 RESTful 中已携带 `organization_id`，但 gRPC 没有 URL 路径；Header 统一两种协议

### D3: Scope 中间件 vs 在 Authz 中间件内处理

**选择独立的 Scope 中间件**，而不是在 Authz 中合并。

原因：
- 职责分离：Scope 负责"解析上下文"，Authz 负责"检查权限"
- 某些接口不需要 Authz 但可能需要 Scope（如 ListOrganizations 当前是 AUTHZ_MODE_NONE，但未来可能需要 org 上下文）
- 独立中间件便于复用到其他服务（如 sayhello）

### D4: OrgScopeMixin 统一方式

**保持现有 edge 定义方式不变**，暂不迁移到 OrgScopeMixin。

原因：
- 当前 `organization_id` 是通过 edge 定义的（`edge.From("organization", Organization.Type)`），Ent 会自动管理外键和关系查询
- 如果改用 mixin 的 plain field，会失去 edge 关系的便利性（如 `org.QueryProjects()`）
- OrgScopeMixin 更适合未来新增的、不需要反向 edge 的实体（如审计日志、资源配额等）
- 对现有 schema，添加 Privacy Policy 时直接基于已有的 `organization_id` 字段操作即可

## 组件设计

### 1. Actor 扩展（`pkg/actor/user.go`）

```go
type UserActor struct {
    id             string
    displayName    string
    email          string
    metadata       map[string]string
    organizationID string  // 新增：当前请求操作的组织
    projectID      string  // 新增：当前请求操作的项目
}

func (a *UserActor) OrganizationID() string          { return a.organizationID }
func (a *UserActor) ProjectID() string                { return a.projectID }
func (a *UserActor) SetOrganizationID(id string)      { a.organizationID = id }
func (a *UserActor) SetProjectID(id string)            { a.projectID = id }
```

### 2. Scope 注入中间件（`pkg/transport/server/middleware/scope.go`）

```go
func ScopeFromHeaders() middleware.Middleware {
    return func(handler middleware.Handler) middleware.Handler {
        return func(ctx context.Context, req any) (any, error) {
            tr, ok := transport.FromServerContext(ctx)
            if !ok {
                return handler(ctx, req)
            }
            htr, ok := tr.(interface{ RequestHeader() transport.Header })
            if !ok {
                return handler(ctx, req)
            }

            a, ok := actor.FromContext(ctx)
            if !ok {
                return handler(ctx, req)
            }
            ua, ok := a.(*actor.UserActor)
            if !ok {
                return handler(ctx, req)
            }

            if orgID := htr.RequestHeader().Get("X-Organization-ID"); orgID != "" {
                if _, err := uuid.Parse(orgID); err != nil {
                    return nil, errors.BadRequest("INVALID_ORGANIZATION_ID", "invalid X-Organization-ID header")
                }
                ua.SetOrganizationID(orgID)
            }
            if projID := htr.RequestHeader().Get("X-Project-ID"); projID != "" {
                if _, err := uuid.Parse(projID); err != nil {
                    return nil, errors.BadRequest("INVALID_PROJECT_ID", "invalid X-Project-ID header")
                }
                ua.SetProjectID(projID)
            }

            return handler(ctx, req)
        }
    }
}
```

### 3. 中间件执行顺序

```
Recovery → Tracing → Logging → RateLimit → Validate → Metrics
  → Authn（JWT → UserActor）
  → ScopeFromHeaders（Header → actor.SetOrganizationID/SetProjectID）
  → Authz（OpenFGA Check）
  → Handler
```

Scope 中间件放在 Authn 之后（需要 UserActor 存在）、Authz 之前（Authz 可能需要读取 scope）。

### 4. 对现有 Authz 的影响

当前 Authz 从 proto request 的字段中提取 `organization_id`/`project_id`。引入 Scope 后，Authz 可以**优先从 Actor 读取 scope**，回退到 request field。这是可选优化，不是必须的。

## 迁移策略

1. Actor 扩展是纯追加，不破坏现有 API
2. Scope 中间件是 opt-in，需要在 server 注册时添加
3. OrgScopeMixin 保持现状，新增实体时推荐使用
4. 前端需要在请求中添加 `X-Organization-ID` header（独立任务）
