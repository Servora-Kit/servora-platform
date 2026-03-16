# 多租户 Actor 扩展与组织作用域

## 动机

当前 servora IAM 服务已具备 Tenant → Organization → Project 的层级模型和 OpenFGA 授权，但请求级别缺乏统一的**组织/项目作用域上下文**。每个需要 org/project 范围的操作都需要从请求参数中手动提取 ID，且 Actor 只包含用户身份信息，无法承载"当前操作的组织"这一关键上下文。

参考 Kemate 项目的 Viewer 模式（JWT 只放身份，`X-Workspace-ID` 走 Header，中间件注入 viewer），以及 go-wind-admin 的 UserViewer 模式，servora 需要在 Actor 层面引入作用域信息，并通过中间件自动提取和注入。

## 目标

1. 扩展 `pkg/actor` 中的 `UserActor`，使其能够携带请求级的 Organization ID 和 Project ID
2. 新增 Scope 注入中间件，从请求 Header（`X-Organization-ID` / `X-Project-ID`）提取作用域并校验权限
3. 统一使用 `OrgScopeMixin` 替代各 schema 手动定义 `organization_id`，为后续自动过滤打基础

## Capabilities

- **actor-scope-extension**：扩展 UserActor 支持组织/项目作用域
- **scope-injection-middleware**：请求级作用域注入中间件
- **org-scope-mixin-unification**：统一 OrgScopeMixin 使用

## 非目标

- 不引入 Ent Privacy Policy 自动过滤（后续可选）
- 不修改 JWT Claims 结构（作用域走 Header）
- 不修改前端（前端适配是独立任务）
