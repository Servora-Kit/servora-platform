# IAM Web 静态部署与 OIDC Provider 边界

> Status: accepted

`iam/web` 是 IAM OIDC Provider 的同源静态交互界面，采用 Next.js static export 和浏览器 CSR；IAM Go service 负责 IAM API、HttpOnly 登录会话、CAP 与 OIDC 协议。Web 静态资源作为独立部署单元提供，公开入口由网关保持单一 origin 并按路径转发页面资源与 Go 接口，避免把前端发布和静态资源生命周期耦合到 Go service。

## Considered Options

- **在 IAM Go service 中嵌入 Web 资源**：不采用。独立静态部署可以支持前端本地热重载和独立发布，Go service 专注 API、CAP 与 OIDC。
- **让 IAM Web 使用 Next.js 请求时服务端运行时**：不采用。认证、会话和协议状态已经由 Go service 管理，静态页面加同源 API 足够完成 IAM 交互。
- **让 IAM Web 直接成为 OAuth client**：不采用。IAM Web 只承载 Provider 自有登录交互，不接触 OAuth client 密钥和 token。

## Consequences

- 页面资源和 Go 接口可以独立部署，但公开入口必须保持同一个 origin，以继续使用 HttpOnly Cookie 和同源请求防护。
- IAM OIDC 协议由 Go 测试验证，Web smoke 只验证页面交互和 Provider 登录导航。
- 未来业务应用是否使用 SSR、服务端运行时或 BFF，按具体需求另行决定。
