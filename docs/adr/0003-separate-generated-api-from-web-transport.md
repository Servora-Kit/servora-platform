# Plateau 生成 API 契约与 Web 传输适配分离

> Status: accepted

Plateau Web 应用以可再生成的 `@plateau/api` 作为唯一 Proto HTTP 契约，并由手写的 `@plateau/client` 统一承担 HTTP transport 与框架中立错误适配；认证状态、业务流程和用户文案仍由各应用拥有。这样生成产物可以无条件重建，传输策略可以独立演进，同时避免每个 Web 应用重复实现一套 client。

## Considered Options

- **让 `@plateau/api` 同时拥有生成契约、具体 HTTP runtime 和认证状态**：拒绝。手写 runtime 与应用状态会污染可再生成边界，并把浏览器 Cookie 与 confidential BFF 两种认证模型错误地合并。
- **各 Web 应用自行实现 transport 与错误适配**：拒绝。重试、错误事实、ProtoJSON 请求和流生命周期会在调用方之间分叉。

## Consequences

- `@plateau/api` 只发布生成的类型、操作客户端和路径合同，不拥有具体 HTTP 库或应用状态。
- `@plateau/client` 同时服务生成 API 与无生成代码的手写 HTTP/OIDC 端点，但不持久化 Cookie、OAuth token、client secret 或业务路由状态。
- 具体 HTTP 库、package 构建方式、重试、流式传输、测试和 TypeScript 工具链由对应 OpenSpec spec/design 记录，不在本 ADR 固化。