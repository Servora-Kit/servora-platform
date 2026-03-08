## ADDED Requirements

### Requirement: Redis 客户端必须提供分布式锁能力

系统必须在 `pkg/redis/` 中提供 `TryLock(ctx context.Context, key string, ttl time.Duration) (*Lock, error)` 方法。锁基于 Redis SET NX 命令实现，使用随机 token 标识持有者。

#### Scenario: 成功获取锁

- **WHEN** 调用 `client.TryLock(ctx, "order:123:lock", 10*time.Second)` 且该 key 不存在
- **THEN** 返回 `*Lock` 实例和 nil error，Redis 中该 key 设置了 10 秒 TTL

#### Scenario: 锁已被其他持有者占用

- **WHEN** 调用 `client.TryLock(ctx, "order:123:lock", 10*time.Second)` 且该 key 已存在
- **THEN** 返回 nil Lock 和 `ErrLockNotAcquired` error

#### Scenario: context 已取消

- **WHEN** 调用 `client.TryLock(ctx, key, ttl)` 且 ctx 已被取消
- **THEN** 返回 nil Lock 和 context 错误

### Requirement: Lock 必须支持安全释放

系统必须提供 `Lock.Unlock(ctx context.Context) error` 方法，使用 Lua 脚本原子性地验证 token 并删除 key，防止误释放他人持有的锁。

#### Scenario: 持有者释放自己的锁

- **WHEN** 锁持有者调用 `lock.Unlock(ctx)`
- **THEN** Redis 中该 key 被删除，返回 nil error

#### Scenario: 锁已过期后尝试释放

- **WHEN** 锁的 TTL 已过期，持有者调用 `lock.Unlock(ctx)`
- **THEN** 返回 `ErrLockNotHeld` error（key 已不存在或 token 不匹配）

#### Scenario: 他人持有的锁尝试释放

- **WHEN** 锁被 A 持有，B 使用不同的 Lock 实例调用 `Unlock(ctx)`
- **THEN** 返回 `ErrLockNotHeld` error（token 不匹配，不会删除 key）

### Requirement: 框架必须提供 Cache-aside 泛型 helper

系统必须在 `pkg/redis/` 中提供 `GetOrSet[T]` 泛型函数，实现「查缓存 → 未命中调用 loader → 写回缓存」模式。

#### Scenario: 缓存命中

- **WHEN** 调用 `GetOrSet[User](ctx, client, "user:1", ttl, loader, marshal, unmarshal)` 且 key 存在于 Redis
- **THEN** 从 Redis 读取值并通过 unmarshal 反序列化返回，不调用 loader

#### Scenario: 缓存未命中

- **WHEN** 调用 `GetOrSet[User](ctx, client, "user:1", ttl, loader, marshal, unmarshal)` 且 key 不存在于 Redis
- **THEN** 调用 loader 获取数据，通过 marshal 序列化后写入 Redis（TTL 为指定值），返回 loader 的结果

#### Scenario: loader 返回 error

- **WHEN** 缓存未命中且 loader 返回 error
- **THEN** 不写入缓存，直接返回 loader 的 error

#### Scenario: Redis 读取失败时降级

- **WHEN** 调用 GetOrSet 但 Redis 读取操作失败（如网络错误）
- **THEN** 降级为直接调用 loader 获取数据，返回 loader 的结果（缓存失败不应阻塞业务）

### Requirement: 框架必须提供 JSON 场景的 Cache-aside 便捷函数

系统必须提供 `GetOrSetJSON[T]` 函数，内置 JSON 序列化/反序列化，简化最常见的使用场景。

#### Scenario: JSON 缓存命中

- **WHEN** 调用 `GetOrSetJSON[User](ctx, client, "user:1", ttl, loader)` 且 key 存在
- **THEN** 从 Redis 读取 JSON 字符串并反序列化为 `T` 类型返回

#### Scenario: JSON 缓存写入

- **WHEN** 调用 `GetOrSetJSON[User](ctx, client, "user:1", ttl, loader)` 且 key 不存在
- **THEN** 调用 loader 获取数据，以 JSON 格式写入 Redis，返回 loader 的结果
