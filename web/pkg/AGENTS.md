# AGENTS.md - web/pkg/

<!-- Parent: ../../AGENTS.md -->
<!-- Generated: 2026-03-15 | Updated: 2026-03-22 -->

## 目录定位

`web/pkg/` 是前端共享工具库，包名 `@servora/web-pkg`。
这里放 **跨前端应用复用** 的请求处理、Token 管理、Kratos 错误解析等基础能力，服务于 `@servora/api-client` 生成客户端的实际调用。

## 当前内容

| 文件 | 作用 |
|------|------|
| `request.ts` | `createRequestHandler()`、`ApiError`、Token 刷新、统一请求头/超时/全局错误处理 |
| `errors.ts` | 解析 Kratos 标准错误体，提供 `parseKratosError()` / `isKratosReason()` / `kratosMessage()` |
| `package.json` | workspace 包定义，当前依赖 `ofetch` |

## 修改约定

- 优先把这里当作 **proto client 适配层**，而不是业务逻辑目录
- 新增能力前先判断它是否能被多个前端应用复用；如果只服务单个应用，优先放回对应 `web/<service>/`
- 保持 API 设计小而稳定：这里导出的类型/函数会成为多个应用的共享契约
- 错误处理需兼容 Kratos 返回格式：`{ code, reason, message, metadata? }`
- 如果修改了 Token 刷新、请求头注入、`onError` 触发时机，必须检查是否会影响现有调用方的重试/提示行为

## 与生成客户端的关系

- 不要在这里手写业务接口类型；请求/响应类型来自 `@servora/api-client`
- `request.ts` 的职责是为生成客户端提供通用 `RequestHandler`，而不是替代生成客户端本身
- 如需变更调用约定，优先保持与 `web/iam/` 当前接入方式兼容

## 禁止事项

- 不要把页面状态、业务 store、路由相关逻辑放进来
- 不要把某个服务专属的 toast / 文案 / 页面跳转策略硬编码到共享层
- 不要在这里复制 `api/gen/ts/` 里的生成类型或客户端实现

## 验证建议

- 修改后至少检查 `web/iam/` 的调用是否仍能消费这些导出
- 如果涉及 proto 变更，先在仓库根执行 `make api-ts`，不要试图手修生成产物
