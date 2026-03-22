# AGENTS.md - pkg/openfga/

<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-22 | Updated: 2026-03-22 -->

## 模块目的

提供 OpenFGA 客户端与常用操作封装，统一组织配置、关系检查、列表查询、tuple 写入与缓存辅助。

## 当前文件

- `client.go`：客户端构造与基础配置校验
- `config.go`：配置结构
- `check.go`：关系检查封装
- `list.go`：列表查询封装
- `tuples.go`：tuple 写入/管理辅助
- `cache.go`：缓存辅助

## 当前实现事实

- `client.go` 要求 `api_url` 与 `store_id`，可选 `model_id`、`api_token`
- `check.go` 是授权关系判定的主要封装入口
- `list.go`、`tuples.go` 分别聚焦查询与关系写入
- `cache.go` 负责 OpenFGA 查询优化辅助，但不改变授权语义本身

## 边界约束

- 本包只封装 OpenFGA API 与通用调用模式，不负责策略设计与资源规则建模
- 不把业务授权决策散落到 client wrapper 中；业务语义应留在 `pkg/authz` 或上层
- 不在这里承载 Redis 通用能力；缓存仅是 OpenFGA 场景优化

## 常见反模式

- 在 `pkg/openfga` 中硬编码业务资源名、关系名和领域规则
- 把缓存命中逻辑与授权结论语义混为一谈
- 直接在业务层重复拼装 OpenFGA client 而绕过统一 wrapper

## 测试与使用

```bash
go test ./pkg/openfga/...
```

## 维护提示

- 若修改配置字段或 client 初始化要求，需同步检查所有服务配置模板
- 若扩展缓存策略，优先保证缓存失效不会放宽授权边界
