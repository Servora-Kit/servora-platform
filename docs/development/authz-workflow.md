# AuthZ 开发运维工作流

本文档描述 servora 项目中 OpenFGA 授权体系的开发、变更和运维流程。

## 架构概览

```
Proto 注解 (.proto)            OpenFGA 模型 (.fga)
       │                              │
       ▼                              ▼
protoc-gen-servora-authz       svr openfga init/apply
       │                              │
       ▼                              ▼
authz_rules.gen.go             OpenFGA 实例 (store + model)
       │                              │
       └──────────┬───────────────────┘
                  ▼
    IAM 服务内 AuthZ 中间件 (运行时)
    app/iam/service/internal/server/middleware/authz.go
         ┌─────────────────┐
         │ 1. 查 rules map │
         │ 2. 解析 object  │
         │ 3. FGA Check()  │
         │ 4. allow/deny   │
         └─────────────────┘
```

## 初始化流程

### 首次设置

```bash
# 1. 安装开发工具（含 protoc-gen-servora-authz 和 svr）
make init

# 2. 启动基础设施（PostgreSQL + Redis + OpenFGA）
make compose.up

# 3. 初始化 OpenFGA store 和上传模型
make openfga.init
# 或
svr openfga init

# 4. 生成代码（含 AuthZ rules）
make gen

# 5. 启动服务
make compose.dev
```

`svr openfga init` 会：
- 连接 OpenFGA API（默认 `http://localhost:8080`）
- 查找或创建名为 `servora` 的 store
- 解析 `manifests/openfga/model/servora.fga` 并上传
- 将 `FGA_API_URL`、`FGA_STORE_ID`、`FGA_MODEL_ID` 写入 `.env`

## 变更流程

### 场景 1：修改权限模型（servora.fga）

当需要新增类型、角色或权限关系时：

```bash
# 1. 编辑模型文件
vim manifests/openfga/model/servora.fga

# 2. 验证语法（可选，需 fga CLI）
make openfga.model.validate

# 3. 上传新版本到 OpenFGA（自动前置 validate → test）
make openfga.model.apply

# 或直接用 svr CLI（仅 DSL 语法验证，不运行 .fga.yaml 测试）
svr openfga model apply

# 4. 重启服务（使新 FGA_MODEL_ID 生效）
```

`make openfga.model.apply` 会按顺序执行：
1. `openfga.model.validate` — 语法检查（需 fga CLI）
2. `openfga.model.test` — 运行 `.fga.yaml` 测试用例（需 fga CLI）
3. 上传新版本到 OpenFGA 并更新 `.env`

`svr openfga model apply` 直接上传，DSL 解析阶段会做语法验证（不依赖 fga CLI），但不运行 `.fga.yaml` 语义测试。

注意事项：
- 每次 `model apply` 会创建新的 model version，`.env` 中的 `FGA_MODEL_ID` 自动更新
- OpenFGA model 升级向后兼容：新增 type/relation 不影响已有 tuple
- 删除 relation 会导致对应 tuple 的 Check 失败，需提前清理

### 场景 2：修改 RPC 权限注解（proto）

当需要调整某个 API 端点的权限要求时：

```bash
# 1. 编辑 proto 文件中的 authz.rule 注解
vim app/iam/service/api/protos/servora/iam/service/v1/i_organization.proto

# 2. 重新生成代码（自动更新 authz_rules.gen.go）
make api

# 3. 重新编译服务
go build ./app/iam/service/...
```

### 场景 3：新增服务类型

当需要新增一个资源类型（如 `team`）时：

1. **OpenFGA 模型**：在 `servora.fga` 中添加 `type team` 及其 relations
2. **Ent Schema**：在 `app/iam/service/internal/data/schema/` 中新增 schema
3. **Proto**：新增 proto 服务定义 + HTTP 聚合 proto + AuthZ 注解
4. **业务代码**：Biz usecase + Data repo + Service layer + OpenFGA tuple 双写
5. **AuthZ 枚举**：如需新增 ObjectType/Relation，更新 `app/iam/service/api/protos/servora/authz/service/v1/authz.proto`
6. **生成与同步**：`make gen` + `make openfga.model.apply`

## AuthZ Proto 注解参考

在 IAM HTTP 聚合 proto（`i_*.proto`）的每个 RPC 方法上声明权限：

```protobuf
import "servora/authz/service/v1/authz.proto";

rpc GetOrganization(...) returns (...) {
  option (google.api.http) = {get: "/v1/organizations/{id}"};
  option (servora.authz.v1.rule) = {
    mode: AUTHZ_MODE_ORGANIZATION
    relation: RELATION_CAN_VIEW
    id_field: "id"
  };
}
```

### AuthzMode 说明

| Mode | 行为 | id_field |
|------|------|----------|
| `AUTHZ_MODE_NONE` | 跳过授权（公开端点） | 不需要 |
| `AUTHZ_MODE_ORGANIZATION` | 检查 `organization:{id}` 上的 relation | 请求消息中的字段名 |
| `AUTHZ_MODE_PROJECT` | 检查 `project:{id}` 上的 relation | 请求消息中的字段名 |
| `AUTHZ_MODE_OBJECT` | 检查 `{object_type}:{id}` 上的 relation | 请求消息中的字段名，或 `"root"` |

### 中间件行为

- **Fail-closed**：未标注 AuthZ 注解的 RPC 方法会被中间件直接拒绝（403）
- **OpenFGA 不可用**：返回 503，不会降级放行
- **字段提取**：通过 proto reflection 从已绑定的请求消息中读取 `id_field` 对应的值

## 自定义 protoc 插件

`cmd/protoc-gen-servora-authz/` 是一个 protoc 插件：
- 读取 proto 文件中的 `(servora.authz.v1.rule)` 方法选项
- 生成 `authz_rules.gen.go` 到 `api/gen/go/servora/iam/service/v1/`
- 输出 `AuthzRules map[string]AuthzRuleEntry`，供 IAM 服务内 AuthZ 中间件（`app/iam/service/internal/server/middleware/authz.go`）查表使用
- 集成在 `make api` 流程中（通过 `buf.go.gen.yaml` 中的 `protoc-gen-servora-authz` 插件）

## OpenFGA Tuple 双写

业务代码在创建/修改资源时需同步写入 OpenFGA tuple：

```go
// 创建组织时，写入 platform 父关系 + owner 角色
fga.WriteTuples(ctx,
    openfga.NewTuple("platform:"+platID, "platform", "organization:"+org.ID),
    openfga.NewTuple("user:"+userID, "owner", "organization:"+org.ID),
)
```

确保在数据库事务成功后再写入 tuple，失败时记录日志但不回滚（最终一致性）。
