# AGENTS.md - pkg/

<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-15 | Updated: 2026-03-20 -->

## 目录概览

`pkg/` 提供跨服务复用的基础能力。当前一级子目录：`actor`、`audit`、`authn`、`authz`、`bootstrap`、`broker`、`cap`、`ent`、`governance`、`health`、`helpers`、`jwks`、`jwt`、`k8s`、`logger`、`mail`、`mapper`、`openfga`、`pagination`、`redis`、`swagger`、`transport`。

## 模块速览

| 目录 | 用途 |
|------|------|
| `actor/` | 请求 Actor v2（User/Service/System/Anonymous），含 Subject/Roles/Scopes/Attrs |
| `audit/` | 审计事件运行时（Emitter、Recorder、Kratos middleware 骨架与多种 emitter） |
| `authn/` | 基于 JWT 的认证中间件，负责 token 解析、claims 映射与 actor 注入 |
| `authz/` | 基于 OpenFGA 的授权中间件，消费 protoc 生成的 AuthzRule |
| `bootstrap/` | 启动链路、配置加载、服务身份解析与 Runtime 生命周期 |
| `broker/` | 消息代理抽象接口（Broker/Message/Subscriber） |
| `broker/kafka/` | franz-go 实现（producer/consumer/config） |
| `cap/` | 基于 Redis 的内嵌式 Cap PoW 人机验证服务端 |
| `ent/` | Ent 驱动、scope 与 schema mixin |
| `governance/` | 注册发现、配置中心、遥测 |
| `health/` | 健康检查 Handler / Checker / Pinger 组合 |
| `helpers/` | 无状态通用辅助与 bcrypt 哈希 |
| `jwks/` | JWKS 管理、响应生成与端点辅助 |
| `jwt/` | JWT 签发、校验与 claims context 注入 |
| `k8s/` | Kubernetes clientset 与运行时环境探测 |
| `logger/` | Kratos + Zap，含 GORM / Ent 桥接 |
| `mail/` | Mail Sender 抽象与 SMTP 发送实现 |
| `mapper/` | 类型安全映射器、plan / preset / hook / converter |
| `openfga/` | OpenFGA 客户端、check/list/tuple/cache 封装 |
| `pagination/` | PaginationRequest / PaginationResponse 辅助 |
| `redis/` | Redis 客户端与锁 / Cache-aside |
| `swagger/` | Swagger UI 与 OpenAPI 文档挂载 |
| `transport/` | 服务间 transport client / server / middleware |

## 当前事实

- `governance/` 下已经是 `config/`、`registry/`、`telemetry/`，旧的 `configCenter/` 描述已失效
- `logger/` 除 `gorm_log.go` 外还有 `ent_log.go`
- `middleware/whitelist.go` 实现的是 **operation 白名单**，不是 IP 白名单
- `redis/` 当前目录没有单独测试文件

## 开发约定

- 优先保持无状态、可复用、低耦合
- 需要资源释放时返回 `cleanup func()`
- 不在库代码里 `panic` 或 `log.Fatal`
- 依赖生成配置类型时，从 `api/gen/go/servora/conf/v1` 导入

## 常用命令

```bash
go test ./pkg/...
go test ./pkg/logger/...
go test ./pkg/governance/registry/...
go test ./pkg/redis/...
```

## 维护提示

- 若更新共享基础设施能力，优先同步本文件与对应子模块 AGENTS.md
- `helpers` 承担密码哈希等通用辅助
