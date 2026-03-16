# 任务：多租户 Actor 扩展与组织作用域

## 第一阶段：Actor 扩展

- [ ] **T1: 扩展 UserActor 结构体**
  - 文件：`pkg/actor/user.go`
  - 新增 `organizationID`、`projectID` 字段
  - 新增 `OrganizationID()`、`ProjectID()` getter
  - 新增 `SetOrganizationID()`、`SetProjectID()` setter
  - 验证：`go build ./pkg/actor/...`

- [ ] **T2: 新增 Actor scope 便捷函数**
  - 文件：`pkg/actor/context.go`
  - 新增 `OrganizationIDFromContext(ctx) (string, bool)` 便捷函数
  - 新增 `ProjectIDFromContext(ctx) (string, bool)` 便捷函数
  - 内部逻辑：取出 Actor → 断言为 `*UserActor` → 返回 scope ID
  - 验证：`go test ./pkg/actor/...`

## 第二阶段：Scope 注入中间件

- [ ] **T3: 实现 ScopeFromHeaders 中间件**
  - 文件：`pkg/transport/server/middleware/scope.go`
  - 从 `X-Organization-ID` 和 `X-Project-ID` 请求头提取值
  - UUID 格式校验，非法值返回 400
  - 写入 `UserActor.SetOrganizationID()` / `SetProjectID()`
  - Header 不存在时静默跳过（scope 是可选的）
  - 验证：`go build ./pkg/transport/...`

- [ ] **T4: 在 IAM 服务中注册 Scope 中间件**
  - 文件：`app/iam/service/internal/server/http.go`、`grpc.go`
  - 在 Authn 之后、Authz 之前插入 `ScopeFromHeaders()`
  - 验证：`go build ./app/iam/service/...`

## 第三阶段：业务层适配（示范性）

- [ ] **T5: 在 OrganizationUsecase 中使用 scope context**
  - 文件：`app/iam/service/internal/biz/organization.go`
  - 在 `Create` 方法中，如果 `actor.OrganizationIDFromContext` 有值且请求未显式指定 tenant，可用 scope 中的值
  - 这是示范性改动，验证 scope 在 biz 层可正确获取
  - 验证：`go build ./app/iam/service/...`

## 第四阶段：验证

- [ ] **T6: 编译与冒烟测试**
  - `go build ./...` 全项目编译通过
  - `make gen` 确认生成代码与手动改动兼容
  - 本地启动 `make compose.dev` 验证 IAM 服务正常启动
  - 使用 curl 验证：带 `X-Organization-ID` header 的请求不报错，不带也不报错
