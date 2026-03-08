## 1. Health Probe 核心实现

- [x] 1.1 创建 `pkg/health/health.go`：定义 `Checker` 接口（Name + Check）、`Pinger` 接口、`Handler` 结构体、`NewHandler` 构造函数
- [x] 1.2 实现 `Handler.LivenessHandler()`：始终返回 HTTP 200 + `{"status": "alive"}`
- [x] 1.3 实现 `Handler.ReadinessHandler()`：执行所有 checker（带 3 秒超时 context），全部通过返回 200，任一失败返回 503，响应体包含每个 checker 的状态
- [x] 1.4 实现 `PingChecker(name string, pinger Pinger) Checker` 工厂函数

## 2. Health Probe HTTP Server 集成

- [x] 2.1 在 `pkg/transport/server/http/` 中添加 `WithHealthCheck(*health.Handler)` option，注册 `GET /healthz` 和 `GET /readyz` 路由
- [x] 2.2 确保未传入 `WithHealthCheck` 时不注册任何健康检查路由

## 3. Health Probe 测试

- [x] 3.1 编写 `pkg/health/health_test.go`：测试无 checker 的 Handler、全部通过的场景、部分失败的场景、checker 超时场景
- [x] 3.2 编写 `pkg/transport/server/http/server_test.go` 补充测试：验证 WithHealthCheck option 正确注册路由

## 4. Redis 分布式锁实现

- [x] 4.1 创建 `pkg/redis/lock.go`：定义 `Lock` 结构体（持有 client、key、token）、`ErrLockNotAcquired` 和 `ErrLockNotHeld` 错误变量
- [x] 4.2 实现 `Client.TryLock(ctx, key, ttl) (*Lock, error)`：基于 SET NX + 随机 token
- [x] 4.3 实现 `Lock.Unlock(ctx) error`：基于 Lua 脚本原子验证 token 并删除 key

## 5. Redis Cache-aside 实现

- [x] 5.1 创建 `pkg/redis/cache.go`：实现 `GetOrSet[T]` 泛型函数（查缓存 → loader → 写回），Redis 读取失败时降级直接调用 loader
- [x] 5.2 实现 `GetOrSetJSON[T]` 便捷函数：内置 JSON marshal/unmarshal

## 6. Redis 扩展测试

- [x] 6.1 编写 `pkg/redis/lock_test.go`：测试获取锁成功、锁已占用、安全释放、过期后释放、token 不匹配释放
- [x] 6.2 编写 `pkg/redis/cache_test.go`：测试缓存命中、缓存未命中、loader 错误、Redis 故障降级、JSON 便捷函数

## 7. 文档与验证

- [x] 7.1 更新 `pkg/health/AGENTS.md`：记录模块目的、API、使用示例
- [x] 7.2 更新 `pkg/redis/AGENTS.md`：补充分布式锁和 Cache-aside 的 API 和使用示例
- [x] 7.3 运行 `go build ./pkg/health/... && go build ./pkg/redis/...` 确认编译通过
- [x] 7.4 运行 `go test ./pkg/health/... && go test ./pkg/redis/...` 确认测试通过
