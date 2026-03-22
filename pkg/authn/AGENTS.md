# AGENTS.md - pkg/authn/

<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-22 | Updated: 2026-03-22 -->

## 模块目的

提供基于 JWT 的认证中间件，负责从请求中提取 Bearer Token、完成校验、把 claims 映射为 `actor.Actor`，并写回 transport 上下文。

## 当前文件

- `authn.go`：认证 middleware 与相关选项
- `authn_test.go`：认证中间件测试

## 当前实现事实

- middleware 会尝试从请求头提取 Bearer token
- 成功校验后将 claims 映射为 `actor.Actor`
- 未携带 token 时可注入匿名 Actor，保证下游读取身份时有稳定语义
- verifier 为空时允许透传，便于在某些场景下按需关闭校验
- token 还会写入 transport token context，供后续链路复用

## 边界约束

- 本包是 middleware 层 glue code，不是 JWT 基础库；签发/验签细节在 `pkg/jwt`
- 本包不承载组织、项目、租户等业务 claims 解释规则
- 本包不直接做资源级授权；授权决策应留给 `pkg/authz` 或业务层

## 常见反模式

- 在 `pkg/authn` 中堆积业务 claims 解释和领域规则
- 把匿名身份、缺 token、验签失败三种状态混成一种处理
- 绕过 `actor` / transport context，直接在业务层重复解析 token

## 测试与使用

```bash
go test ./pkg/authn/...
```

## 维护提示

- 若调整 claims 到 Actor 的映射字段，需同步检查 `pkg/actor` 接口契约
- 若调整 token 提取策略，优先确认与 `pkg/transport/server/middleware` 的兼容性
