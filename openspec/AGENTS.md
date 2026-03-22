# AGENTS.md - openspec/

<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-15 | Updated: 2026-03-22 -->

## 目录定位

`openspec/` 保存 OpenSpec 工作流资产，包括进行中的变更、归档变更与长期 specs。
这里的约定来自仓库当前 `openspec/config.yaml`，不是通用模板。

## 结构

```text
openspec/
├── changes/        # 进行中与已归档的变更
│   └── archive/    # 已归档变更
├── specs/          # 长期能力规格（如 actor-v2、logger-refactor 等）
└── config.yaml     # OpenSpec 规则与上下文
```

## config.yaml 关键规则

### proposal
- 必须包含 `Non-goals` 段落
- 必须引用主设计文档 phase 编号
- scope 要严格限制在单个 phase 内

### design
- 需要时加入 ASCII 架构图
- 设计 `pkg/broker` 时参考 `kratos-transport` broker 接口
- 设计 Kafka / audit 集成时参考 `Kemate` 模式
- 明确写出 breaking changes 与 migration notes

### tasks
- 单个任务块控制在 4 小时以内
- 相关任务按逻辑章节分组
- 标记需要 `make gen` / `make api` 的再生成步骤
- 不要遗漏 docker-compose、配置、基础设施等配套任务

## 编写约定

- 变更文档应紧贴当前仓库事实：Go workspace、多模块、Buf v2、Wire、Ent、pnpm workspace
- 当涉及 `app/iam/service`、`app/sayhello/service` 时，注意它们当前更偏参考实现，不要误写成唯一活跃服务
- 引用外部参考项目时，保留 `config.yaml` 中已有的基准项目和用途说明
- 如果设计会影响 Proto / OpenAPI / 自定义插件，任务中应明确生成链路，而不是默认读者会自己想到

## 禁止事项

- 不要把 OpenSpec 文档写成脱离本仓库的泛化 RFC 模板
- 不要遗漏 Non-goals、phase 边界、breaking changes 这类已在 config 中硬性要求的内容
- 不要把需要执行的生成/迁移步骤隐藏在描述里；应在 tasks 中显式列出
