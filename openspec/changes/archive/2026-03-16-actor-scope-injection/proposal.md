# Actor Scope 注入：请求级组织/项目上下文

## 动机

当前 servora IAM 服务已具备 Tenant → Organization → Project 的层级模型和 OpenFGA 授权，但请求级别缺乏统一的**组织/项目作用域上下文**。每个需要 org/project 范围的操作都需要从请求参数中手动提取 ID，且 Actor 只包含用户身份信息，无法承载"当前操作的组织"这一关键上下文。

参考 Kemate 项目的 Viewer 模式（JWT 只放身份，`X-Workspace-ID` 走 Header，中间件注入 viewer 并在 AuthZ + Service 层消费），以及 go-wind-admin 的 UserViewer 模式（JWT claims 携带 `tid`/`ouid`），servora 需要在 Actor 层面引入作用域信息，通过 Header + 中间件自动提取注入。

## 目标

1. 扩展 `pkg/actor` 中的 `UserActor`，使其能够携带请求级的 Organization ID 和 Project ID
2. 新增 Scope 注入中间件，从请求 Header（`X-Organization-ID` / `X-Project-ID`）提取作用域并写入 Actor
3. 改造 Authz 中间件，scope-based 端点从 Actor scope 读取 org/project ID（纯 Header，无 request field 混合）
4. 迁移 4 个 scope-based 端点的 proto 定义和 HTTP URL

## Capabilities

- **actor-scope-extension**：扩展 UserActor 支持组织/项目作用域
- **scope-injection-middleware**：请求级作用域注入中间件
- **authz-scope-resolution**：Authz 中间件支持 scope vs resource 两种 ID 来源

## 非目标

- 不引入 Ent Privacy Policy 自动过滤（后续可选）
- 不修改 JWT Claims 结构（作用域走 Header）
- 不迁移现有 schema 的 Edge 到 OrgScopeMixin（现有 Edge 关系保持不变）
- 不修改 resource/sub-resource 端点的 URL 结构（如 `/v1/organizations/{id}`、`/v1/organizations/{id}/members`）
