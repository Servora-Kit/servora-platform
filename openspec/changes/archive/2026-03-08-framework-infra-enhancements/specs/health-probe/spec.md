## ADDED Requirements

### Requirement: 框架必须提供组件化的健康探针 Handler

系统必须提供 `pkg/health` 包，包含 `Handler` 类型，通过 `NewHandler(checkers ...Checker)` 构造。Handler 必须暴露 `LivenessHandler()` 和 `ReadinessHandler()` 两个方法，分别返回 `http.HandlerFunc`。

#### Scenario: 创建无 checker 的 Handler

- **WHEN** 调用 `health.NewHandler()` 且不传入任何 checker
- **THEN** 成功创建 Handler 实例，LivenessHandler 返回 200，ReadinessHandler 返回 200

#### Scenario: 创建带 checker 的 Handler

- **WHEN** 调用 `health.NewHandler(redisChecker, dbChecker)` 传入多个 checker
- **THEN** 成功创建 Handler 实例，ReadinessHandler 执行所有 checker

### Requirement: Checker 接口必须定义 Name 和 Check 方法

系统必须定义 `Checker` 接口，包含 `Name() string` 和 `Check(ctx context.Context) error` 两个方法。Check 返回 nil 表示健康，返回 error 表示不健康。

#### Scenario: 健康的 checker

- **WHEN** 某个 Checker 的 `Check(ctx)` 返回 nil
- **THEN** 该 checker 在 readiness 响应中标记为 healthy

#### Scenario: 不健康的 checker

- **WHEN** 某个 Checker 的 `Check(ctx)` 返回 error
- **THEN** 该 checker 在 readiness 响应中标记为 unhealthy，整体 readiness 返回 503

### Requirement: Liveness 端点必须始终返回 200

系统的 Liveness 端点（`/healthz`）必须始终返回 HTTP 200 状态码和 `{"status": "alive"}` JSON 响应体。Liveness 端点禁止执行任何 checker。

#### Scenario: 进程存活时的 liveness 响应

- **WHEN** 向 `/healthz` 发送 GET 请求
- **THEN** 返回 HTTP 200 和 `{"status": "alive"}`

#### Scenario: 依赖不可用时的 liveness 响应

- **WHEN** 向 `/healthz` 发送 GET 请求，且已注册的某个 checker 会返回 error
- **THEN** 仍然返回 HTTP 200 和 `{"status": "alive"}`（不受 checker 影响）

### Requirement: Readiness 端点必须执行所有注册的 checker

系统的 Readiness 端点（`/readyz`）必须执行所有已注册的 `Checker`，使用带超时的 context（默认 3 秒）。全部通过返回 200，任一失败返回 503。

#### Scenario: 所有 checker 通过

- **WHEN** 向 `/readyz` 发送 GET 请求，且所有 checker 的 `Check(ctx)` 返回 nil
- **THEN** 返回 HTTP 200 和 `{"status": "ready", "checks": {"redis": "ok", "db": "ok"}}`

#### Scenario: 某个 checker 失败

- **WHEN** 向 `/readyz` 发送 GET 请求，且 redis checker 返回 error
- **THEN** 返回 HTTP 503 和 `{"status": "not_ready", "checks": {"redis": "connection refused", "db": "ok"}}`

#### Scenario: checker 执行超时

- **WHEN** 向 `/readyz` 发送 GET 请求，且某个 checker 在 3 秒内未返回
- **THEN** 该 checker 标记为 timeout 错误，整体返回 503

### Requirement: 框架必须提供内置的 Ping checker 工厂

系统必须提供 `PingChecker(name string, pinger Pinger) Checker` 工厂函数。`Pinger` 接口定义 `Ping(ctx context.Context) error` 方法，兼容 `redis.Client` 和 `sql.DB` 的 Ping 签名。

#### Scenario: 使用 Redis 客户端创建 checker

- **WHEN** 调用 `health.PingChecker("redis", redisClient)` 且 redisClient 实现了 `Pinger` 接口
- **THEN** 返回一个 Name 为 "redis" 的 Checker，其 Check 方法调用 redisClient.Ping(ctx)

#### Scenario: Ping 失败

- **WHEN** PingChecker 的底层 Pinger.Ping(ctx) 返回 error
- **THEN** Checker.Check(ctx) 返回相同的 error

### Requirement: HTTP server 必须支持通过 option 挂载健康探针

系统必须在 `pkg/transport/server/http/` 中提供 `WithHealthCheck(*health.Handler)` option。该 option 将 `/healthz` 和 `/readyz` 路由注册到 HTTP server。

#### Scenario: 启用健康探针

- **WHEN** 构建 HTTP server 时传入 `http.WithHealthCheck(healthHandler)`
- **THEN** server 注册 `GET /healthz` 和 `GET /readyz` 路由

#### Scenario: 不启用健康探针

- **WHEN** 构建 HTTP server 时未传入 `WithHealthCheck` option
- **THEN** server 不注册任何健康检查路由
