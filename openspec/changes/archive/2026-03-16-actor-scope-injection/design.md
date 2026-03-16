# 设计：Actor Scope 注入

## 架构概览

```
请求流：
  Authorization: Bearer <jwt>       → Authn 中间件 → UserActor(id, name, email)
  X-Organization-ID: <uuid>         → Scope 中间件 → actor.SetOrganizationID()
  X-Project-ID: <uuid>              → Scope 中间件 → actor.SetProjectID()
                                    → Authz 中间件 → OpenFGA Check
                                    → Handler → Service/Biz（从 Actor 取 scope）
```

Scope 是请求的**唯一来源**（纯 Header，参考 Kemate 的 `X-Workspace-ID` 模式），不走 request field 混合。

## 决策

### D1: Actor 扩展 vs 独立 Viewer 接口

**选择 Actor 扩展**，而不是新建独立 Viewer 接口。

原因：
- servora 已有成熟的 `pkg/actor` 包，`UserActor` 在整个调用链中都通过 context 传递
- 新建 Viewer 会导致 context 中有两个身份/作用域对象，增加混淆
- go-wind-admin 和 Kemate 的 Viewer 模式本质上就是"带作用域的 Actor"，只是名称不同
- 扩展 `UserActor` 加 setter/getter 即可，向后兼容

### D2: 作用域来源——纯 Header

**选择纯 Header（`X-Organization-ID` / `X-Project-ID`）**，参考 Kemate。

两个参考项目：
- **Kemate**：完全依赖 `X-Workspace-ID` header，AuthZ 中间件校验后写入 viewer，Service 层通过 `requireWorkspaceScope(ctx)` 消费
- **go-wind-admin**：完全依赖 JWT claims（`tid` / `ouid`），登录时确定

servora 选择纯 Header 的原因：
- 用户可属于多个 Organization，切换组织不应要求重新签发 JWT → 排除纯 JWT
- 单一来源避免了 "request field vs header 谁优先" 的隐式覆盖复杂性
- Header 在 HTTP 和 gRPC（metadata）中统一工作
- Kemate 已在生产验证这个模式

**Scope vs Resource 的区分：**

需要区分两种不同的 `organization_id` 用法：
- **Scope 参数**：表示"在哪个组织下操作"（如 `ListProjects` — 列哪个组织的项目）→ **从 Header 读**
- **Resource 标识**：表示"操作哪个具体资源"（如 `GetOrganization` — 获取哪个组织）→ **从 URL/request 读**
- **Sub-resource 父级**：表示"操作哪个资源的子项"（如 `AddMember` — 给哪个组织加成员）→ **从 URL/request 读**

只有第一类（Scope 参数）迁移到 Header，后两类保留在 URL path/request field。

**受影响的端点（scope → header）：**

| 端点 | 当前 | 迁移后 |
|------|------|--------|
| `ListProjects` | `GET /v1/organizations/{organization_id}/projects` | `GET /v1/projects` + `X-Organization-ID` |
| `CreateProject` | `POST /v1/projects`（body `organization_id`） | `POST /v1/projects` + `X-Organization-ID` |
| `ListApplications` | `GET /v1/organizations/{organization_id}/applications` | `GET /v1/applications` + `X-Organization-ID` |
| `CreateApplication` | `POST /v1/applications`（body `organization_id`） | `POST /v1/applications` + `X-Organization-ID` |

**不受影响的端点（resource/sub-resource，保留 URL path）：**

| 端点 | URL | 原因 |
|------|-----|------|
| `GetOrganization` | `GET /v1/organizations/{id}` | org 是目标资源 |
| `UpdateOrganization` | `PUT /v1/organizations/{id}` | 同上 |
| `DeleteOrganization` | `DELETE /v1/organizations/{id}` | 同上 |
| `AddMember`（org） | `POST /v1/organizations/{organization_id}/members` | org 是父资源 |
| `RemoveMember`（org） | `DELETE /v1/organizations/{organization_id}/members/{user_id}` | 同上 |
| `ListMembers`（org） | `GET /v1/organizations/{organization_id}/members` | 同上 |
| `GetProject` | `GET /v1/projects/{id}` | project 是目标资源 |
| 所有 project member 操作 | `/v1/projects/{project_id}/members/...` | project 是父资源 |

### D3: Scope 中间件 vs 在 Authz 中间件内处理

**选择独立的 Scope 中间件**，而不是在 Authz 中合并。

原因：
- 职责分离：Scope 负责"解析上下文"，Authz 负责"检查权限"
- 某些接口不需要 Authz 但可能需要 Scope（如 `ListOrganizations` 是 `AUTHZ_MODE_NONE`，但 Service 层可能需要 org 上下文）
- 独立中间件便于复用到其他服务（如 sayhello）

### D4: Ent Schema 保持 Edge，不迁移到 OrgScopeMixin

**保持现有 Edge 定义方式不变**。

三项目对比：
- **go-wind-admin**：纯 `mixin.TenantID[uint32]{}`，无 Edge → 失去关系查询能力
- **Kemate**：显式 field + Edge → 两者兼顾
- **servora 当前**：与 Kemate 相同，通过 `edge.From("organization")...Field("organization_id")` 定义

Edge 的优势对 servora 更重要：
- `org.QueryProjects()`、`org.QueryMembers()` 等关系查询
- 级联删除（`entsql.OnDelete(entsql.Cascade)`）
- Ent 自动管理外键约束和关系完整性

OrgScopeMixin 的适用场景：未来新增不需要反向 edge 的实体（如审计日志、操作记录、资源配额），可使用 Mixin 快速添加 `organization_id` 字段和索引。

### D5: Authz 中间件如何区分 Scope 和 Resource

**约定：`id_field` 为空 → 从 Actor scope 读取；`id_field` 有值 → 从 request field 读取。**

Proto 注解示例：

```proto
// Scope-based：从 Header/Actor 读 organization scope
rpc ListProjects(...) returns (...) {
  option (authz.service.v1.rule) = {
    mode: AUTHZ_MODE_ORGANIZATION
    relation: RELATION_CAN_VIEW
    // id_field 为空 → resolveObject 从 Actor.OrganizationID() 读
  };
}

// Resource-based：从 request field 读 organization ID
rpc GetOrganization(...) returns (...) {
  option (authz.service.v1.rule) = {
    mode: AUTHZ_MODE_ORGANIZATION
    relation: RELATION_CAN_VIEW
    id_field: "id"  // → resolveObject 从 request.id 读
  };
}
```

`resolveObject` 的判断逻辑：

```go
case authzpb.AuthzMode_AUTHZ_MODE_ORGANIZATION:
    objectType = "organization"
    if rule.IDField == "" {
        // Scope-based: 从 Actor scope 读取（Header 注入）
        ua, ok := a.(*actor.UserActor)
        if !ok || ua.OrganizationID() == "" {
            err = fmt.Errorf("missing X-Organization-ID header")
        } else {
            objectID = ua.OrganizationID()
        }
    } else {
        // Resource-based: 从 request field 读取
        objectID, err = extractProtoField(req, rule.IDField)
    }
```

### D6: Scope 的消费者——Authz + Service/Biz 层

纯 Header 模式下，**Authz 和 Service 层都从 Actor scope 读取**（通过 Scope 中间件注入）。

消费链路：
1. **ScopeFromHeaders 中间件**：`X-Organization-ID` → `actor.SetOrganizationID()`
2. **Authz 中间件**：`actor.OrganizationID()` → OpenFGA Check（scope-based 端点）
3. **Service/Biz 层**：`actor.OrganizationIDFromContext(ctx)` → 数据操作 scope

参考 Kemate 的模式：
- AuthZ 中间件读 `X-Workspace-ID` header → 校验权限 → 写入 `UserViewer`
- Service 层通过 `requireWorkspaceScope(ctx)` 从 viewer 取 `userID + workspaceID`
- 业务逻辑用 `workspaceID` 来 scope 数据操作

## 组件设计

### 1. Actor 扩展（`pkg/actor/user.go`）

```go
type UserActor struct {
    id             string
    displayName    string
    email          string
    metadata       map[string]string
    organizationID string  // 当前请求操作的组织（从 X-Organization-ID header 注入）
    projectID      string  // 当前请求操作的项目（从 X-Project-ID header 注入）
}

func (a *UserActor) OrganizationID() string      { return a.organizationID }
func (a *UserActor) ProjectID() string            { return a.projectID }
func (a *UserActor) SetOrganizationID(id string)  { a.organizationID = id }
func (a *UserActor) SetProjectID(id string)       { a.projectID = id }
```

### 2. Context 便捷函数（`pkg/actor/context.go`）

```go
func OrganizationIDFromContext(ctx context.Context) (string, bool) {
    a, ok := FromContext(ctx)
    if !ok {
        return "", false
    }
    ua, ok := a.(*UserActor)
    if !ok || ua.organizationID == "" {
        return "", false
    }
    return ua.organizationID, true
}

func ProjectIDFromContext(ctx context.Context) (string, bool) {
    // 同上逻辑
}
```

### 3. Scope 注入中间件（`pkg/transport/server/middleware/scope.go`）

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
                    return nil, errors.BadRequest("INVALID_ORGANIZATION_ID",
                        "invalid X-Organization-ID header")
                }
                ua.SetOrganizationID(orgID)
            }
            if projID := htr.RequestHeader().Get("X-Project-ID"); projID != "" {
                if _, err := uuid.Parse(projID); err != nil {
                    return nil, errors.BadRequest("INVALID_PROJECT_ID",
                        "invalid X-Project-ID header")
                }
                ua.SetProjectID(projID)
            }

            return handler(ctx, req)
        }
    }
}
```

### 4. Authz 中间件改造（`app/iam/service/internal/server/middleware/authz.go`）

`resolveObject` 需要接收 Actor，增加 "scope vs resource" 分支：

```go
func resolveObject(rule iamv1.AuthzRuleEntry, tenantRootID string,
    req any, a actor.Actor) (objectType, objectID string, err error) {

    switch rule.Mode {
    case authzpb.AuthzMode_AUTHZ_MODE_ORGANIZATION:
        objectType = "organization"
        if rule.IDField == "" {
            objectID, err = scopeFromActor(a, "OrganizationID")
        } else {
            objectID, err = extractProtoField(req, rule.IDField)
        }

    case authzpb.AuthzMode_AUTHZ_MODE_PROJECT:
        objectType = "project"
        if rule.IDField == "" {
            objectID, err = scopeFromActor(a, "ProjectID")
        } else {
            objectID, err = extractProtoField(req, rule.IDField)
        }

    case authzpb.AuthzMode_AUTHZ_MODE_OBJECT:
        objectType = objectTypeToFGA(rule.ObjectType)
        if rule.IDField == "root" && objectType == "tenant" {
            objectID = tenantRootID
        } else {
            objectID, err = extractProtoField(req, rule.IDField)
        }

    default:
        err = fmt.Errorf("unsupported authz mode: %v", rule.Mode)
    }
    return
}

func scopeFromActor(a actor.Actor, field string) (string, error) {
    ua, ok := a.(*actor.UserActor)
    if !ok {
        return "", fmt.Errorf("actor is not a UserActor")
    }
    switch field {
    case "OrganizationID":
        if id := ua.OrganizationID(); id != "" {
            return id, nil
        }
        return "", fmt.Errorf("missing X-Organization-ID header")
    case "ProjectID":
        if id := ua.ProjectID(); id != "" {
            return id, nil
        }
        return "", fmt.Errorf("missing X-Project-ID header")
    default:
        return "", fmt.Errorf("unknown scope field: %s", field)
    }
}
```

### 5. 中间件执行顺序

```
Recovery → Tracing → Logging → RateLimit → Validate → Metrics
  → Authn（JWT → UserActor）
  → ScopeFromHeaders（Header → actor.SetOrganizationID/SetProjectID）
  → Authz（scope-based: Actor scope / resource-based: request field）
  → Handler
```

### 6. Service 层消费模式（参考 Kemate 的 `requireWorkspaceScope`）

```go
func requireOrgScope(ctx context.Context) (userID, orgID string, err error) {
    a, ok := actor.FromContext(ctx)
    if !ok || a.Type() != actor.TypeUser {
        return "", "", errors.Unauthorized("UNAUTHORIZED", "unauthorized")
    }
    ua := a.(*actor.UserActor)
    if ua.OrganizationID() == "" {
        return "", "", errors.BadRequest("MISSING_ORGANIZATION_SCOPE",
            "missing X-Organization-ID")
    }
    return ua.ID(), ua.OrganizationID(), nil
}
```

### 7. Proto 注解变更

受影响端点的 authz 注解需要移除 `id_field`：

```proto
// Before (scope from request field):
rpc ListProjects(...) returns (...) {
  option (google.api.http) = {get: "/v1/organizations/{organization_id}/projects"};
  option (authz.service.v1.rule) = {
    mode: AUTHZ_MODE_ORGANIZATION
    relation: RELATION_CAN_VIEW
    id_field: "organization_id"
  };
}

// After (scope from header):
rpc ListProjects(...) returns (...) {
  option (google.api.http) = {get: "/v1/projects"};
  option (authz.service.v1.rule) = {
    mode: AUTHZ_MODE_ORGANIZATION
    relation: RELATION_CAN_VIEW
    // id_field 省略 → resolveObject 从 Actor scope 读取
  };
}
```

### 8. 前端 Scope 传递（参考 Kemate 模式）

参考 Kemate 的 `contextHeaders` + `useAuthStore`：

1. **State 存储**：全局状态（Zustand/Pinia）存 `currentOrganizationId` / `currentProjectId`
2. **请求拦截器**：HTTP client 的 `contextHeaders` 回调自动附加 `X-Organization-ID`
3. **URL 感知**：URL 结构反映组织上下文（如 `/org/{orgSlug}/projects`），切换组织时更新 state
4. **错误恢复**：收到 `MISSING_ORGANIZATION_SCOPE` 400 错误时，从 URL 恢复 scope 并重试

## 迁移策略

1. Actor 扩展是纯追加，不破坏现有调用
2. Scope 中间件在 server 注册时添加
3. Authz 的 `resolveObject` 改造兼容两种模式（`id_field` 有值 = request，为空 = Actor scope）
4. 4 个受影响端点的 proto 注解和 request message 需更新
5. `protoc-gen-servora-authz` 插件需适配 `id_field` 为空的情况
6. 前端 scope 传递作为独立任务
