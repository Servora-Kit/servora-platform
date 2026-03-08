## Context

servora 是基于 Go Kratos v2 的微服务快开框架。当前 `pkg/` 层提供了日志、Redis 连接、中间件链、transport client 等共享能力，但缺少：

1. **健康探针**：K8s manifest 中已配置 livenessProbe/readinessProbe 路径，但服务代码层面无对应端点实现。当前 HTTP server 工厂（`pkg/transport/server/http/`）不提供健康检查路由注册能力。
2. **Redis 高级模式**：`pkg/redis/` 仅封装基础 CRUD，业务中常见的分布式锁和缓存模式需要各服务自行实现，导致重复代码和不一致的错误处理。

框架的核心约束：
- `pkg/` 保持无状态、可复用、低耦合
- 服务通过 Wire 注入依赖，新能力须兼容 ProviderSet 模式
- 现有 HTTP server 使用 `pkg/transport/server/http/` 的 option 模式构建

## Goals / Non-Goals

**Goals:**
- 提供即插即用的 `/healthz` 和 `/readyz` 端点组件，服务在 HTTP server 构建时一行代码引入
- 健康探针支持注册自定义 checker（如检查数据库连接、Redis 连通性），Readiness 与 Liveness 分离
- 提供基于 go-redis 的分布式锁实现，支持自动续期和 context 取消
- 提供 Cache-aside 泛型 helper，封装「查缓存 → 未命中查数据源 → 写回缓存」模式

**Non-Goals:**
- 不实现 gRPC 健康检查协议（grpc.health.v1）——当前框架重点在 HTTP 探针，gRPC 健康检查可后续独立添加
- 不实现 Redlock（多节点分布式锁）——单节点 Redis 锁已满足当前部署模式，多节点场景属于进阶需求
- 不修改现有 `pkg/redis/redis.go` 的 Client 结构体——新能力以独立文件和函数形式提供，不破坏现有 API
- 不处理 Redis Cluster 或 Sentinel 场景——当前框架仅支持单实例 Redis

## Decisions

### 1. 健康探针：独立 `pkg/health/` 包 + HTTP server option 注入

**选择**: 新建 `pkg/health/` 包，定义 `Checker` 接口和 `Handler`，通过 `http.WithHealthCheck()` option 注入 HTTP server。

**而非**: 直接在 `pkg/transport/server/http/server.go` 内硬编码健康路由。

**原因**: 
- 保持 HTTP server 工厂的职责单一
- 允许服务选择性启用（不需要探针的内部服务可以不引入）
- Checker 接口允许服务注册自定义检查逻辑（DB、Redis、外部依赖）

**API 设计**:
```go
// pkg/health/health.go
type Checker interface {
    Name() string
    Check(ctx context.Context) error
}

type Handler struct { ... }

func NewHandler(checkers ...Checker) *Handler
func (h *Handler) LivenessHandler() http.HandlerFunc   // /healthz
func (h *Handler) ReadinessHandler() http.HandlerFunc   // /readyz
```

**服务端使用**:
```go
// 服务的 server/http.go 中
healthHandler := health.NewHandler(
    health.PingChecker("redis", redisClient),  // 内置的 Redis checker
    health.PingChecker("db", dbPinger),        // 内置的 DB checker
)
srv := http.NewServer(
    http.WithHealthCheck(healthHandler),
    // ... 其他 options
)
```

### 2. Liveness vs Readiness 的语义分离

**选择**: Liveness 始终返回 200（进程存活即可），Readiness 执行所有注册的 checker。

**而非**: 两者都执行 checker。

**原因**: 遵循 K8s 最佳实践——Liveness 探针不应依赖外部服务（避免 DB 短暂不可用导致 Pod 重启循环），Readiness 才检查依赖就绪状态。

### 3. Redis 分布式锁：基于 SET NX + Lua 释放

**选择**: 单节点 SET NX + 随机 token 获取锁，Lua 脚本原子释放，支持 context 取消。

**而非**: 引入 `github.com/bsm/redislock` 等第三方库。

**原因**:
- 实现简洁（~80 行），不引入新依赖
- 框架已依赖 `github.com/redis/go-redis/v9`，直接利用其 Eval 能力
- 单节点场景下 SET NX 已足够可靠

**API 设计**:
```go
// pkg/redis/lock.go
type Lock struct { ... }

func (c *Client) TryLock(ctx context.Context, key string, ttl time.Duration) (*Lock, error)
func (l *Lock) Unlock(ctx context.Context) error
```

### 4. Cache-aside helper：泛型函数

**选择**: 提供泛型函数 `GetOrSet[T]`，接受 loader 回调。

**而非**: 在 Client 上定义方法。

**原因**:
- 泛型函数避免 Client 结构体膨胀
- 调用方自行控制序列化/反序列化策略
- 支持自定义 TTL 和 nil 值处理（防缓存穿透）

**API 设计**:
```go
// pkg/redis/cache.go
func GetOrSet[T any](ctx context.Context, c *Client, key string, ttl time.Duration, 
    loader func(ctx context.Context) (T, error),
    marshal func(T) (string, error),
    unmarshal func(string) (T, error),
) (T, error)
```

## Risks / Trade-offs

| 风险 | 缓解措施 |
|---|---|
| 单节点 Redis 锁在 Redis 主从切换时可能失效 | 文档明确标注适用场景，后续可扩展 Redlock |
| Cache-aside 泛型函数签名较长（marshal/unmarshal 参数） | 提供 `GetOrSetJSON[T]` 便捷函数作为 JSON 场景的快捷方式 |
| 健康探针路由（`/healthz`、`/readyz`）可能与业务路由冲突 | 路径固定为约定俗成的 K8s 标准路径，冲突概率极低 |
| Checker 执行超时可能拖慢 Readiness 响应 | checker 执行使用带超时的 context，默认 3 秒 |
