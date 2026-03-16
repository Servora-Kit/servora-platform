# 任务：Actor Scope 注入

## 第一阶段：Actor 扩展

- [x] **T1: 扩展 UserActor 结构体**
  - 文件：`pkg/actor/user.go`
  - 新增 `organizationID`、`projectID` 字段
  - 新增 `OrganizationID()`、`ProjectID()` getter
  - 新增 `SetOrganizationID()`、`SetProjectID()` setter
  - 验证：`go build ./pkg/actor/...`

- [x] **T2: 新增 Actor scope 便捷函数**
  - 文件：`pkg/actor/context.go`
  - 新增 `OrganizationIDFromContext(ctx) (string, bool)` 便捷函数
  - 新增 `ProjectIDFromContext(ctx) (string, bool)` 便捷函数
  - 内部逻辑：取出 Actor → 断言为 `*UserActor` → 返回 scope ID
  - 验证：`go test ./pkg/actor/...`

## 第二阶段：Scope 注入中间件

- [x] **T3: 实现 ScopeFromHeaders 中间件**
  - 文件：`pkg/transport/server/middleware/scope.go`
  - 从 `X-Organization-ID` 和 `X-Project-ID` 请求头提取值
  - UUID 格式校验，非法值返回 400
  - 写入 `UserActor.SetOrganizationID()` / `SetProjectID()`
  - Header 不存在时静默跳过（scope 是可选的）
  - 验证：`go build ./pkg/transport/...`

- [x] **T4: 在 IAM 服务中注册 Scope 中间件**
  - 文件：`app/iam/service/internal/server/http.go`、`grpc.go`
  - 在 Authn 之后、Authz 之前插入 `ScopeFromHeaders()`
  - 验证：`go build ./app/iam/service/...`

## 第三阶段：Authz 中间件改造

- [x] **T5: 改造 resolveObject 支持 scope vs resource**
  - 文件：`app/iam/service/internal/server/middleware/authz.go`
  - `resolveObject` 新增 `actor.Actor` 参数
  - `AUTHZ_MODE_ORGANIZATION` / `AUTHZ_MODE_PROJECT`：`id_field` 为空时从 Actor scope 读取，有值时从 request field 读取
  - 新增 `scopeFromActor` 辅助函数
  - 验证：`go build ./app/iam/service/...`

- [x] **T6: 迁移 4 个 scope-based 端点的 proto 注解**
  - 文件：
    - `app/iam/service/api/protos/iam/service/v1/i_project.proto`（ListProjects、CreateProject）
    - `app/iam/service/api/protos/iam/service/v1/i_application.proto`（ListApplications、CreateApplication）
  - 变更：
    - 移除 scope 端点的 `id_field: "organization_id"` 注解
    - ListProjects URL 从 `/v1/organizations/{organization_id}/projects` 改为 `/v1/projects`
    - ListApplications URL 从 `/v1/organizations/{organization_id}/applications` 改为 `/v1/applications`
  - 对应的 request message 中移除 `organization_id` 字段（scope 来自 header）
  - 验证：`make api` 重新生成

- [x] **T7: 适配 protoc-gen-servora-authz 插件**
  - 文件：`cmd/protoc-gen-servora-authz/`
  - 确保 `id_field` 为空时生成 `IDField: ""`（当前行为应该已经是这样）
  - 验证：`go install ./cmd/protoc-gen-servora-authz && make api`

## 第四阶段：Service 层适配

- [x] **T8: 实现 requireOrgScope 并在 scope-based service 方法中使用**
  - 文件：`app/iam/service/internal/service/scope.go`（新建）
  - 参考 Kemate 的 `requireWorkspaceScope` 模式
  - 实现 `requireOrgScope(ctx) (userID, orgID string, err error)`
  - 在 `ProjectService.CreateProject` 和 `ProjectService.ListProjects` 中使用
  - 在 `ApplicationService.CreateApplication` 和 `ApplicationService.ListApplications` 中使用
  - 验证：`go build ./app/iam/service/...`

## 第五阶段：验证

- [x] **T9: 编译与冒烟测试**
  - `go build ./...` 全项目编译通过
  - `make gen` 确认生成代码与手动改动兼容
  - 本地启动 `make compose.dev` 验证 IAM 服务正常启动
  - 使用 curl 验证：
    - `curl -H "X-Organization-ID: <uuid>" .../v1/projects` → 正常返回
    - 不带 header 的 scope 端点 → 返回 400 MISSING_ORGANIZATION_SCOPE
    - 带非法 UUID header → 返回 400 INVALID_ORGANIZATION_ID
    - resource 端点（如 `GET /v1/organizations/{id}`）不受影响

## 第六阶段：前端 Scope 传递（独立任务，可后续执行）

- [x] **T10: 前端请求拦截器自动附加 scope header**
  - 参考 Kemate 的 `contextHeaders` + `useAuthStore` 模式
  - 在全局状态中存 `currentOrganizationId`
  - HTTP client 拦截器自动附加 `X-Organization-ID` header
  - URL 结构反映组织上下文（如 `/org/{orgSlug}/...`）
  - 收到 `MISSING_ORGANIZATION_SCOPE` 错误时从 URL 恢复 scope
