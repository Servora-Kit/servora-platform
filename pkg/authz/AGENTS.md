# AGENTS.md - pkg/authz/

<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-22 | Updated: 2026-03-22 -->

## 模块目的

提供基于 OpenFGA 的通用授权中间件，消费 protoc 生成的 `AuthzRule`，在请求进入业务层前完成资源级授权判定，并可选产出 `authz.decision` 审计事件。

## 当前文件

- `authz.go`：授权 middleware 主实现（含 `WithAuditRecorder` option 和审计 emit）
- `authz_test.go`：授权链路测试

## 当前实现事实

- 授权依据来自每个 operation 绑定的 `AuthzRule`
- 依赖 `pkg/openfga` 执行关系检查（传完整 principal `"user:" + userID`），可选使用 `pkg/redis` 做缓存
- 适配 `openfga.CachedCheck` 三返回值 `(allowed, cacheHit, err)`
- 通过 `WithAuditRecorder` 注入 `*audit.Recorder`，Check 完成后自动 emit `authz.decision` 事件（allowed/denied/error），包含 `CacheHit` 信息
- 若请求缺少可用规则，默认按 fail-closed 思路处理，而不是放行
- 对象解析依赖 proto 字段/请求消息内容

## 边界约束

- 本包负责授权执行与审计 emit，不负责模型设计、关系写入或 OpenFGA store 运维
- 不在本包定义业务常量、组织树规则或资源生命周期
- 审计 emit 通过可选的 `*audit.Recorder` 实现，nil-safe

## 常见反模式

- 在 middleware 中硬编码业务资源规则，绕过生成的 `AuthzRule`
- 缺少规则时默认放行，导致权限面失控
- 把对象解析、授权决策、业务补偿逻辑揉在一起

## 测试与使用

```bash
go test ./pkg/authz/...
```

## 维护提示

- 若 proto AuthZ 注解有变更，先执行根目录 `make api` 再检查本包调用链
- 若调整缓存策略，需同步确认 `pkg/openfga` 与 `pkg/redis` 的边界仍清晰
- 授权 middleware 现在传完整 principal 给 `openfga.CachedCheck`（`"user:" + a.ID()`）
