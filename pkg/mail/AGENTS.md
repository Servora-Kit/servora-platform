# AGENTS.md - pkg/mail/

<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-22 | Updated: 2026-03-22 -->

## 模块目的

提供邮件发送抽象与 SMTP 实现，统一封装配置驱动的 Sender 创建、默认发件人策略与 noop 回退。

## 当前文件

- `mailer.go`：`Sender` 接口、`NewSender` 与通用装配
- `smtp.go`：SMTP 发送实现

## 当前实现事实

- `Sender` 是上层依赖的稳定抽象
- `NewSender(c *conf.Mail)` 在配置缺失时会返回 noop sender，而不是强制报错
- 支持 `DefaultFrom` 等默认发件人配置
- SMTP 实现基于 `go-mail`，并处理连接参数与 TLS 策略

## 边界约束

- 本包只负责“发送邮件”，不负责模板渲染、营销编排或消息队列重试策略
- 不把业务邮件主题、变量拼装硬编码到共享发送层
- 不让调用方直接依赖底层 SMTP 客户端细节

## 常见反模式

- 在 mail 包里加入业务模板与领域事件判断
- 缺配置时直接让服务启动失败，而不是利用 noop 降级策略
- 绕过 `Sender` 接口直接在业务层 new SMTP 客户端

## 测试与使用

```bash
go test ./pkg/mail/...
```

## 维护提示

- 若调整 noop 回退策略，需同步确认所有依赖方对“未配置邮件”的预期
- 若扩展新后端，优先保持 `Sender` 抽象稳定
