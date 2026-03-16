# Ent 查询 Scope 约定

## 原则

凡涉及 `organization_id` 或 `project_id` 的实体（Project、Application、OrganizationMember、ProjectMember），其 **List / Query / Update / Delete** 方法**必须**在 Where 条件中传入对应 scope（`orgID` 或 `projectID`），禁止无 scope 的全表操作。

## 适用实体

| 实体 | Scope 字段 | 说明 |
|------|-----------|------|
| Organization | — | 根实体，无需 scope |
| Project | `organization_id` | 必须带 orgID |
| Application | `organization_id` | 必须带 orgID |
| OrganizationMember | `organization_id` | 成员方法已按 orgID 传参 |
| ProjectMember | `project_id` | 成员方法已按 projID 传参 |
| User | — | 全局实体，无需 scope |

## Data 层规范

### 必须带 scope 的方法签名

对 Project / Application 等带 `organization_id` 的实体，以下操作的 Repo 接口**必须**包含 `orgID string` 参数：

- `GetByID(ctx, orgID, id)`
- `GetByIDs(ctx, orgID, ids, page, pageSize)`
- `Update(ctx, orgID, entity)`
- `Delete(ctx, orgID, id)`
- `Purge(ctx, orgID, id)`
- `Restore(ctx, orgID, id)`
- `GetByIDIncludingDeleted(ctx, orgID, id)`

### orgID 的处理逻辑

Data 层收到 `orgID` 后：

- **非空且合法 UUID**：在查询中加 `.Where(xxx.OrganizationIDEQ(oid))` 作为 defense-in-depth
- **空字符串**：跳过 scope 过滤（仅用于管理员操作或内部级联）

```go
if oid, err := uuid.Parse(orgID); err == nil {
    query = query.Where(project.OrganizationIDEQ(oid))
}
```

### 无需 scope 的例外

- `PurgeCascade`：内部级联删除，由上层保证正确性
- `GetByClientID`（Application）：OIDC 流程中无 org 上下文
- `ListMembershipsByUserID` / `DeleteMembershipsByUserID`：按 userID 查询，用于用户清理

## Biz 层规范

- 从 `actor.OrganizationIDFromContext(ctx)` 获取 orgID，传入 Data 层
- 若上下文无 orgID（管理员操作），传空字符串 `""`
- List 操作中 orgID 通常来自请求参数，直接透传

## FGA 创建/成员变更回滚

- **创建**（Organization / Project）：`DB Create → AddMember → WriteTuples`；任一步失败回滚前序步骤
- **AddMember**：`DB AddMember → WriteTuples`；FGA 失败则 `RemoveMember` 回滚
- **RemoveMember**：`DB RemoveMember → DeleteTuples`；FGA 失败则 `AddMember` 回滚
- **UpdateMemberRole**：`DB UpdateRole → DeleteTuples(old) → WriteTuples(new)`；FGA 失败则回滚 DB 角色

## Code Review 检查项

新增或修改 Data 层 repo 方法时，Review 必须确认：

1. **是否涉及 org/project 范围的实体**？若是，方法签名是否包含 orgID/projectID
2. **Where 条件是否完整**？单 ID 操作也需带 scope（非空时）
3. **是否有裸查全表**？禁止无 scope 的 List/Query
4. **FGA 操作是否有错误处理**？不允许 `_ = authz.WriteTuples(...)`
5. **FGA 失败是否有回滚**？确认回滚路径正确
