## Context

当前仓库的 proto package 命名存在三类问题：

1. 缺少统一顶层命名空间，公共 proto 与业务 proto 混用短包名，如 `audit.v1`、`mapper.v1`、`conf.v1`、`iam.service.v1`；
2. package、目录、`go_package` 三者只做了局部对齐，没有形成仓库级规范；
3. 仓库启用了 Buf `STANDARD` lint，其中包含 `PACKAGE_DIRECTORY_MATCH`，因此 package 迁移不能只改文本，必须与目录同步调整。

当前 Buf workspace 由 3 个 proto module 组成：

```text
api/protos
app/iam/service/api/protos
app/sayhello/service/api/protos
```

现有 proto 主要分布如下：

```text
api/protos/audit/v1                -> package audit.v1
api/protos/mapper/v1               -> package mapper.v1
api/protos/pagination/v1           -> package pagination.v1
api/protos/conf/v1                 -> package conf.v1
app/iam/service/api/protos/...     -> package iam.service.v1 / authn.service.v1 / authz.service.v1 / ...
app/sayhello/service/api/protos/...-> package sayhello.service.v1
```

本次变更是命名治理，不改变业务语义，但属于明确的 breaking change，因为它会改动：

- protobuf full name
- proto import 路径
- `go_package` 导入路径
- 生成代码目录
- 下游 Go / TS 引用路径

## Goals / Non-Goals

**Goals:**
- 为所有 proto 建立统一的 `servora.` 顶层命名空间。
- 统一保留版本后缀，收敛到 `servora.*.v1` 体系。
- 让目录层级、proto `package`、`go_package`、生成目录保持一致。
- 为后续新增 proto 提供可执行的命名规范，避免再次出现无命名空间短包名。
- 在不改变业务 schema 语义的前提下完成迁移，保持能力边界不变。

**Non-Goals:**
- 不调整字段定义、RPC 语义、审计事件语义或配置结构语义。
- 不在本变更中引入新 generator、新 annotation 或新运行时逻辑。
- 不改变 Buf workspace module 划分。
- 不处理向 `v2` 演进的多版本并存方案。

## Decisions

### Decision 1: 统一采用 `servora.` 顶层 proto 命名空间

所有 proto package 统一迁移到：

```text
servora.<domain>...<version>
```

示例：

```text
audit.v1                -> servora.audit.v1
mapper.v1               -> servora.mapper.v1
pagination.v1           -> servora.pagination.v1
conf.v1                 -> servora.conf.v1
iam.service.v1          -> servora.iam.service.v1
authn.service.v1        -> servora.authn.service.v1
sayhello.service.v1     -> servora.sayhello.service.v1
```

**Why this over keeping current names?**
- 当前短包名在仓库扩大后更易冲突；
- 顶层命名空间能表达 Servora 作为协议与框架提供者；
- 这与 protobuf 生态中 `google.protobuf`、`google.api` 的命名思路一致，更贴近 proto 自身语境，而不是语言包路径语境。

**Why not use `com.servora.*`?**
- `com.servora.*` 更像 Java 包路径风格；
- 对 proto package 而言，`servora.*` 已足够表达稳定命名空间；
- 当前仓库的核心约束来自 Buf 与多语言 proto 生成，不是 Java 式命名组织。

### Decision 2: 统一保留 `v1`，包括公共 proto 与业务 proto

本变更统一采用带版本 package：

```text
servora.audit.v1
servora.mapper.v1
servora.pagination.v1
servora.conf.v1
servora.iam.service.v1
```

**Why this over only versioning business-domain protos?**
- `audit`、`mapper`、`pagination`、`conf` 也是共享 schema，不只是内部实现；
- 这些 proto 一样会被生成代码、跨模块引用、长期演进；
- 如果首版不带版本，后续引入 breaking change 时会造成“无版本 + v2”并存的不整齐状态。

**Alternative considered:** 仅业务领域保留 `v1`、公共 proto 去掉版本。

**Why rejected?**
- “业务/非业务”不是稳定边界；`audit` 与 `mapper` 虽非业务域，但本质是公共契约；
- 规则会在后续扩展时变得主观，难以治理。

### Decision 3: 目录必须与 package 全量同步迁移

由于 Buf 启用了 `PACKAGE_DIRECTORY_MATCH`，目录与 package 必须严格一致。因此目标目录形态应类似：

```text
api/protos/servora/audit/v1/audit.proto
api/protos/servora/mapper/v1/mapper.proto
api/protos/servora/pagination/v1/pagination.proto
api/protos/servora/conf/v1/conf.proto
app/iam/service/api/protos/servora/authn/service/v1/authn.proto
app/iam/service/api/protos/servora/iam/service/v1/i_authn.proto
app/sayhello/service/api/protos/servora/sayhello/service/v1/sayhello.proto
```

**Why this over changing lint rules?**
- 关闭或绕过 lint 只会把不一致保留下来；
- 目录即命名空间的规则有助于长期治理与检索；
- 对 proto-first 仓库而言，这比“短期少改目录”更值钱。

### Decision 4: `service` 不是默认层级，但若目录语义存在则保留

本次不强行删除已有业务 proto 中的 `service` 层级，而是遵守“目录语义即 package 语义”：

- 若文件位于 `.../authn/service/v1/`，则保留 `servora.authn.service.v1`；
- 若公共 proto 本来不属于服务接口面，则保持 `servora.audit.v1`、`servora.mapper.v1`，不额外补 `service`。

**Why this over flattening all packages now?**
- 当前目标是命名治理，不是信息架构重构；
- 先统一顶层命名空间与版本规则，风险更可控；
- 扁平化/重构领域层级可留给后续专门变更处理。

### Decision 5: `go_package` 与生成目录同步迁移，但 package alias 保持可读

`go_package` 需要随目录迁移而调整，例如：

```text
github.com/Servora-Kit/servora/api/gen/go/servora/audit/v1;auditv1
github.com/Servora-Kit/servora/api/gen/go/servora/iam/service/v1;iampb
```

原则：
- 路径与 proto package / proto 目录保持同形；
- alias 延续已有可读风格（如 `auditv1`、`iampb`、`authnpb`），避免把重命名扩散到无意义的局部变量风格变更。

**Why this over preserving old generated paths?**
- 保留旧路径会让 package、目录、生成物三套坐标不一致；
- 迁移成本虽然更高，但能一次性收敛长期债务。

## Risks / Trade-offs

- [广泛 breaking change] → 先完成映射表与迁移顺序设计，再集中迁移 proto/import/gen，避免半迁移状态。
- [Go / TS 引用路径大面积变动] → 以生成代码为中心统一更新 import，引入一次性全量验证（`make api`、`make build`、`make lint`）。
- [旧文档与设计稿引用过时路径] → 在迁移任务中包含文档引用收敛，但仅修正明确受影响的 proto 路径。
- [服务域层级仍不完全理想] → 本次只治理命名空间与版本，不顺带做更大范围的领域重构。
- [OpenSpec 既有 spec 引用旧路径] → 通过 delta specs 修改“命名与路径规范”要求，而不是重写既有能力语义。

## Mapping Baseline

### Decision Resolution

本次业务 proto **只补 `servora.` 顶层前缀**，**不**同步收敛到 `servora.iam.*` 之类的新领域层级。

这意味着：
- `authn.service.v1` → `servora.authn.service.v1`
- `authz.service.v1` → `servora.authz.service.v1`
- `user.service.v1` → `servora.user.service.v1`
- `application.service.v1` → `servora.application.service.v1`
- `iam.service.v1` → `servora.iam.service.v1`
- `iam.conf.v1` → `servora.iam.conf.v1`
- `sayhello.service.v1` → `servora.sayhello.service.v1`
- `template.service.v1` → `servora.template.service.v1`

**Why this over introducing `servora.iam.<capability>.*` now?**
- 当前 change 目标是 namespace 治理，不是领域重构；
- Buf `PACKAGE_DIRECTORY_MATCH` 已要求目录整体迁移，再叠加领域重排会显著扩大 blast radius；
- 先把 `package` / 目录 / `go_package` / generated imports 收敛到同一坐标系，再考虑二阶段信息架构优化更稳妥。

### Current → Target Mapping Table

| Current proto path | Current package | Current go_package | Target proto path | Target package | Target go_package |
| --- | --- | --- | --- | --- | --- |
| `api/protos/audit/v1/audit.proto` | `audit.v1` | `github.com/Servora-Kit/servora/api/gen/go/audit/v1;auditv1` | `api/protos/servora/audit/v1/audit.proto` | `servora.audit.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/audit/v1;auditv1` |
| `api/protos/audit/v1/annotations.proto` | `audit.v1` | `github.com/Servora-Kit/servora/api/gen/go/audit/v1;auditv1` | `api/protos/servora/audit/v1/annotations.proto` | `servora.audit.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/audit/v1;auditv1` |
| `api/protos/mapper/v1/mapper.proto` | `mapper.v1` | `github.com/Servora-Kit/servora/api/gen/go/mapper/v1;mapperpb` | `api/protos/servora/mapper/v1/mapper.proto` | `servora.mapper.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/mapper/v1;mapperpb` |
| `api/protos/pagination/v1/pagination.proto` | `pagination.v1` | `github.com/Servora-Kit/servora/api/gen/go/pagination/v1;paginationpb` | `api/protos/servora/pagination/v1/pagination.proto` | `servora.pagination.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/pagination/v1;paginationpb` |
| `api/protos/conf/v1/conf.proto` | `conf.v1` | `github.com/Servora-Kit/servora/api/gen/go/conf/v1;conf` | `api/protos/servora/conf/v1/conf.proto` | `servora.conf.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1;conf` |
| `api/protos/template/service/v1/template.proto` | `template.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/template/service/v1;v1` | `api/protos/servora/template/service/v1/template.proto` | `servora.template.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/template/service/v1;templatepb` |
| `api/protos/template/service/v1/template_doc.proto` | `template.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/template/service/v1;v1` | `api/protos/servora/template/service/v1/template_doc.proto` | `servora.template.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/template/service/v1;templatepb` |
| `app/iam/service/api/protos/authn/service/v1/authn.proto` | `authn.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/authn/service/v1;authnpb` | `app/iam/service/api/protos/servora/authn/service/v1/authn.proto` | `servora.authn.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/authn/service/v1;authnpb` |
| `app/iam/service/api/protos/authz/service/v1/authz.proto` | `authz.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/authz/service/v1;authzpb` | `app/iam/service/api/protos/servora/authz/service/v1/authz.proto` | `servora.authz.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/authz/service/v1;authzpb` |
| `app/iam/service/api/protos/user/service/v1/user.proto` | `user.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/user/service/v1;userpb` | `app/iam/service/api/protos/servora/user/service/v1/user.proto` | `servora.user.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/user/service/v1;userpb` |
| `app/iam/service/api/protos/application/service/v1/application.proto` | `application.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/application/service/v1;apppb` | `app/iam/service/api/protos/servora/application/service/v1/application.proto` | `servora.application.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/application/service/v1;apppb` |
| `app/iam/service/api/protos/iam/service/v1/i_authn.proto` | `iam.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/iam/service/v1;iampb` | `app/iam/service/api/protos/servora/iam/service/v1/i_authn.proto` | `servora.iam.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/iam/service/v1;iampb` |
| `app/iam/service/api/protos/iam/service/v1/i_user.proto` | `iam.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/iam/service/v1;iampb` | `app/iam/service/api/protos/servora/iam/service/v1/i_user.proto` | `servora.iam.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/iam/service/v1;iampb` |
| `app/iam/service/api/protos/iam/service/v1/i_application.proto` | `iam.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/iam/service/v1;iampb` | `app/iam/service/api/protos/servora/iam/service/v1/i_application.proto` | `servora.iam.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/iam/service/v1;iampb` |
| `app/iam/service/api/protos/iam/service/v1/iam_doc.proto` | `iam.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/iam/service/v1;iampb` | `app/iam/service/api/protos/servora/iam/service/v1/iam_doc.proto` | `servora.iam.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/iam/service/v1;iampb` |
| `app/iam/service/api/protos/iam/conf/v1/config.proto` | `iam.conf.v1` | `github.com/Servora-Kit/servora/api/gen/go/iam/conf/v1;conf` | `app/iam/service/api/protos/servora/iam/conf/v1/config.proto` | `servora.iam.conf.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/iam/conf/v1;conf` |
| `app/sayhello/service/api/protos/sayhello/service/v1/sayhello.proto` | `sayhello.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/sayhello/service/v1;sayhellopb` | `app/sayhello/service/api/protos/servora/sayhello/service/v1/sayhello.proto` | `servora.sayhello.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/sayhello/service/v1;sayhellopb` |
| `app/sayhello/service/api/protos/sayhello/service/v1/sayhello_doc.proto` | `sayhello.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/sayhello/service/v1;sayhellopb` | `app/sayhello/service/api/protos/servora/sayhello/service/v1/sayhello_doc.proto` | `servora.sayhello.service.v1` | `github.com/Servora-Kit/servora/api/gen/go/servora/sayhello/service/v1;sayhellopb` |

## Migration Plan

```text
Phase A: 固化规范与映射表
    -> 列出当前 package / 目录 / go_package 到目标值的完整映射
Phase B: 迁移公共 proto
    -> audit / mapper / pagination / conf
Phase C: 迁移业务 proto
    -> iam / authn / authz / user / application / sayhello / template
Phase D: 更新 import 与 codegen
    -> proto imports / go_package / generated outputs / Go imports / TS outputs
Phase E: 全量验证
    -> buf lint / make api / make api-ts / make build / 关键测试
```

回滚策略：
- 该变更应以单分支集中迁移完成，不在半迁移状态下合并；
- 若验证失败，直接回退整个变更集，而不是部分保留。

## Open Questions

- `template/service/v1` 是否作为脚手架示例一并迁移到正式命名规范下？
- TypeScript 生成产物的对外 import 前缀是否需要同步暴露 `servora/...` 层级，还是只在物理路径中体现？
