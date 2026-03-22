# AGENTS.md - pkg/governance/

<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-15 | Updated: 2026-03-15 -->

## 模块目的

提供服务治理相关基础能力，当前按 `registry/`、`config/`、`telemetry/` 三个一级专题子目录组织注册发现、配置中心与遥测辅助能力。

## 当前结构

```text
pkg/governance/
├── config/
├── registry/
└── telemetry/
```

## 当前实现事实

- `registry/` 支持 `consul`、`etcd`、`nacos`、`kubernetes`
- `registry/registry.go` 提供统一入口；目录内还有 `etcd_watcher.go`
- `config/` 承载 Consul / Etcd / Nacos 配置源实现
- `telemetry/` 当前包含 `metrics.go` 与 `tracing.go`
- 本级目录主要表达治理能力边界，不递归描述各子目录实现细节

## 边界约束

- 本目录负责治理能力分层与公共约定，不承载服务私有注册、配置装配或观测策略
- 不把具体业务模块接入方式硬编码到共享治理目录
- 更细粒度规则应在 `registry/`、`config/`、`telemetry/` 各自子目录维护

## 常见反模式

- 在 `pkg/governance` 根目录继续堆积 provider 细节，而不是放入对应子目录
- 把服务私有依赖发现逻辑写进共享治理层
- 将遥测初始化与业务指标语义揉在一起

## 使用位置

- `app/iam/service/internal/server/` 通过 `registry.NewRegistrar` 与 `telemetry.NewMetrics` 接入
- `app/iam/service/internal/data/data.go` 通过 `registry.NewDiscovery` 注入服务发现

## 测试

```bash
go test ./pkg/governance/registry/...
go test ./pkg/governance/config/...
```

## 维护提示

- 若新增治理子专题，优先保持一级目录划分清晰，并同步更新父级 `pkg/AGENTS.md`
- 若 provider 支持矩阵变化，先同步本文件“当前实现事实”
