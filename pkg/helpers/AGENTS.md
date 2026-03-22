# AGENTS.md - pkg/helpers/

<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-22 | Updated: 2026-03-22 -->

## 模块目的

提供无状态、可复用的通用辅助函数，当前包含时间/字符串辅助与密码哈希能力。

## 当前文件

- `helpers.go`：`MicrosecondsStr`、`Slugify` 等通用 helper
- `hash.go`：密码哈希与校验相关 helper

## 当前实现事实

- 当前 helper 主要集中在轻量、纯函数式工具
- `hash.go` 承担 bcrypt 类密码哈希能力，是少数与安全相关的通用辅助
- 该目录没有状态管理，也不应依赖服务级业务上下文

## 边界约束

- 这里只能放跨服务、通用、低耦合 helper
- 不把业务规则、配置装配、数据库访问或 transport 逻辑塞进 helpers
- 不把“暂时没地方放”的代码丢进本目录作为兜底

## 常见反模式

- 把业务特定字符串拼装或 DTO 转换塞进 helpers
- 在 helper 中隐式读取环境变量、全局状态或外部资源
- 让密码哈希 API 与某个服务账户模型耦合

## 测试与使用

```bash
go test ./pkg/helpers/...
```

## 维护提示

- 新增 helper 前先确认它是否真的可跨服务复用
- 涉及密码哈希策略变更时，需评估存量数据兼容与安全影响
