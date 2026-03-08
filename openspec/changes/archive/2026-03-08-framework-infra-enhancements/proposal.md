## Why

servora 框架当前缺少两项微服务基础设施能力：服务健康探针和 Redis 高级使用模式。Health/Readiness 端点是 K8s 部署刚需但代码层面无实现；Redis 封装仅提供连接管理，缺少分布式锁和缓存辅助模式。

注：客户端熔断（circuitbreaker.Client()）已在 `pkg/transport/client/grpc_conn.go` 中挂载，无需额外处理。

## What Changes

- 新增 `pkg/health/` 组件化健康探针包，提供 `/healthz` 和 `/readyz` 端点能力，服务在 server 层按需引用
- 扩展 `pkg/redis/` 提供分布式锁（基于单节点 SET NX + context 超时）和 Cache-aside helper 能力

## Capabilities

### New Capabilities
- `health-probe`: 组件化的 Health/Readiness 探针，支持自定义 checker 注册，服务按需在 HTTP server 中挂载
- `redis-patterns`: Redis 高级使用模式，包括分布式锁和 Cache-aside 缓存辅助

### Modified Capabilities

（无）

## Impact

- **新增包**: `pkg/health/`
- **扩展文件**: `pkg/redis/` 目录新增文件（lock.go、cache.go）
- **新增依赖**: 无（健康探针基于标准 net/http，Redis 锁基于现有 go-redis）
- **服务影响**: 现有服务不受影响，两项能力均需服务端主动引用才生效
- **K8s 部署**: 服务挂载健康探针后，manifest 中的 livenessProbe/readinessProbe 配置才会生效
