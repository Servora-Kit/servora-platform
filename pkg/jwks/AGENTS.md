# AGENTS.md - pkg/jwks/

<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-22 | Updated: 2026-03-22 -->

## 模块目的

提供 JWKS 管理与发布辅助能力，负责从 RSA 私钥/PEM 构造公钥集合响应，并组织相关端点输出。

## 当前文件

- `manager.go`：JWKS manager 主实现
- `client.go`：JWKS 客户端辅助
- `config.go`：JWKS 配置
- `endpoints.go`：JWKS 端点装配辅助

## 当前实现事实

- `manager.go` 依赖 `pkg/jwt` 的 signer 能力生成对应的 JWKS 响应
- 支持从私钥路径或 PEM 内容创建管理器
- 暴露的是“公钥集合发布”能力，而不是 token 签发本身
- 该包与 JWT 强关联，但职责仍是发布/管理公钥信息

## 边界约束

- 签发、验签、claims 处理属于 `pkg/jwt`，不应回流到 `pkg/jwks`
- 不在这里承载登录态、认证 middleware 或业务会话逻辑
- 不把 JWKS 端点协议细节散落到其他目录重复实现

## 常见反模式

- 在 `pkg/jwks` 中补充 token 签发逻辑，导致与 `pkg/jwt` 职责重叠
- 直接手写 JWK 响应而绕过 manager/signer 统一链路
- 让 JWKS 配置耦合服务内部业务配置结构

## 测试与使用

```bash
go test ./pkg/jwks/...
```

## 维护提示

- 若调整 key 加载方式，优先保证与 `pkg/jwt` 的 KID 计算保持一致
- 若变更端点返回结构，需确认现有验证方和缓存方兼容性
