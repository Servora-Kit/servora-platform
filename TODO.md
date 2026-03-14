# TODO

## 级联删除原子性与僵尸数据

- **现状**：`PurgeUser` 跨 FGA / Redis / Postgres 三步执行，无分布式事务；DB 内 `PurgeCascade` 已用 `InTx` 保证原子。
- **风险**：服务在 FGA/Redis/DB 任两步之间崩溃会导致跨系统不一致（如 DB 已删用户、FGA 仍留 tuple）。
- **遗漏**：`PurgeCascade` 只删 `OrganizationMember`、`ProjectMember`、`User`，未删该用户创建的默认 `Organization` / `Project`，会留下孤儿行。
- **待办**：
  1. 将 PurgeUser 顺序改为先 DB（PurgeCascade）再 FGA 再 Redis，或为 FGA/Redis 做补偿/重试。
  2. 在 PurgeCascade 同一事务内按依赖顺序删除该用户拥有的 Organization、Project（或通过 schema 外键 CASCADE 由 DB 级联）。
  3. 可选：PurgeUser 打点/日志，便于排查中断点；提供按 user_id 清理 FGA/Redis 残留的补偿脚本或管理接口。
