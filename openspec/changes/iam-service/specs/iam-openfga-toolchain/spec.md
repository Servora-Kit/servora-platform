# IAM OpenFGA Toolchain Specification

## ADDED Requirements

### Requirement: The system MUST validate OpenFGA models

系统 MUST 提供工具验证 OpenFGA 模型文件的语法正确性。

#### Scenario: Validate model syntax

- **WHEN** 开发者执行 `make openfga.model.validate`
- **THEN** 系统使用 `fga model validate` 命令验证 `manifests/openfga/model/iam.fga` 文件、返回验证结果

#### Scenario: Validation fails with syntax error

- **WHEN** 模型文件包含语法错误
- **THEN** 系统返回错误信息（行号、错误描述）、退出码非零

#### Scenario: Validation succeeds

- **WHEN** 模型文件语法正确
- **THEN** 系统返回成功消息、退出码为零

### Requirement: The system MUST test OpenFGA models

系统 MUST 提供工具测试 OpenFGA 模型的权限逻辑正确性。

#### Scenario: Run model tests

- **WHEN** 开发者执行 `make openfga.model.test`
- **THEN** 系统使用 `fga model test` 命令运行 `manifests/openfga/tests/iam.fga.yaml` 中的测试用例、返回测试结果

#### Scenario: Test case passes

- **WHEN** 测试用例的预期结果与实际结果一致
- **THEN** 系统标记测试为通过、继续执行下一个测试

#### Scenario: Test case fails

- **WHEN** 测试用例的预期结果与实际结果不一致
- **THEN** 系统标记测试为失败、显示预期值和实际值、退出码非零

#### Scenario: All tests pass

- **WHEN** 所有测试用例通过
- **THEN** 系统返回成功消息、退出码为零

### Requirement: The system MUST initialize OpenFGA store

系统 MUST 提供工具初始化 OpenFGA store 和上传权限模型。

#### Scenario: Initialize OpenFGA store

- **WHEN** 开发者执行 `make openfga.init`
- **THEN** 系统创建 OpenFGA store（如果不存在）、上传 `manifests/openfga/model/iam.fga` 模型、返回 store_id 和 authorization_model_id

#### Scenario: Store already exists

- **WHEN** OpenFGA store 已存在
- **THEN** 系统跳过创建、上传新版本的模型、返回新的 authorization_model_id

#### Scenario: Initialization fails

- **WHEN** OpenFGA 服务不可用
- **THEN** 系统返回错误 "Failed to connect to OpenFGA"、退出码非零

### Requirement: The system MUST deploy and version OpenFGA models

系统 MUST 提供工具将模型部署到运行中的 OpenFGA 实例。

#### Scenario: Apply model to running instance

- **WHEN** 开发者执行 `make openfga.model.apply`
- **THEN** 系统上传 `manifests/openfga/model/iam.fga` 到运行中的 OpenFGA、返回新的 authorization_model_id

#### Scenario: Model version tracking

- **WHEN** 模型被部署
- **THEN** 系统记录 authorization_model_id 到配置文件或环境变量

#### Scenario: Rollback to previous model

- **WHEN** 新模型导致问题需要回滚
- **THEN** 系统支持切换到之前的 authorization_model_id

### Requirement: The system MUST support JWKS key rotation

系统 MUST 提供工具执行 JWKS 密钥轮换，采用三阶段流程（分发 → 切换 → 清理）。

#### Scenario: Phase 1 - Distribute new key

- **WHEN** 管理员执行 `./scripts/rotate-jwks-key.sh distribute`
- **THEN** 系统生成新密钥对、将新公钥添加到 JWKS、保持使用旧密钥签发 Token

#### Scenario: Phase 2 - Switch to new key

- **WHEN** 管理员执行 `./scripts/rotate-jwks-key.sh switch`（在分发后至少 15 分钟）
- **THEN** 系统切换到使用新密钥签发 Token、旧密钥仍保留在 JWKS 中用于验证

#### Scenario: Phase 3 - Cleanup old key

- **WHEN** 管理员执行 `./scripts/rotate-jwks-key.sh cleanup`（在切换后至少 15 分钟）
- **THEN** 系统从 JWKS 中移除旧密钥

#### Scenario: Rotation status check

- **WHEN** 管理员执行 `./scripts/rotate-jwks-key.sh status`
- **THEN** 系统显示当前密钥状态（活跃密钥、待清理密钥、上次轮换时间）

### Requirement: The system MUST provide OpenFGA model documentation

系统 MUST 提供清晰的 OpenFGA 模型文档，说明权限模型的设计和使用。

#### Scenario: Model documentation includes entity types

- **WHEN** 开发者查看 `docs/iam/openfga.md`
- **THEN** 文档必须列出所有实体类型（platform、tenant、workspace、user）及其关系

#### Scenario: Model documentation includes relation definitions

- **WHEN** 开发者查看模型文档
- **THEN** 文档必须说明每个关系的含义（owner、admin、member、viewer）和继承规则

#### Scenario: Model documentation includes usage examples

- **WHEN** 开发者查看模型文档
- **THEN** 文档必须提供权限检查的示例代码（如何调用 Check API、如何使用 ListObjects）

### Requirement: The system MUST provide OpenFGA debugging tools

系统 MUST 提供工具帮助调试 OpenFGA 权限问题。

#### Scenario: Query relation tuples

- **WHEN** 开发者需要查看用户的权限关系
- **THEN** 系统提供命令查询指定用户的所有关系元组

#### Scenario: Explain permission decision

- **WHEN** 开发者需要理解为什么用户有或没有某个权限
- **THEN** 系统提供命令显示权限检查的推理路径（如 OpenFGA Expand API）

#### Scenario: List all users with permission

- **WHEN** 开发者需要查看哪些用户有访问某个资源的权限
- **THEN** 系统提供命令列出所有有权限的用户

### Requirement: The system MUST provide OpenFGA performance monitoring

系统 MUST 提供工具监控 OpenFGA 的性能指标。

#### Scenario: Monitor Check API latency

- **WHEN** 系统运行时
- **THEN** 监控系统记录 OpenFGA Check API 的响应时间（P50、P95、P99）

#### Scenario: Alert on high latency

- **WHEN** OpenFGA Check API 的 P99 响应时间超过 100ms
- **THEN** 系统发送告警通知

#### Scenario: Monitor ListObjects API latency

- **WHEN** 系统运行时
- **THEN** 监控系统记录 OpenFGA ListObjects API 的响应时间

#### Scenario: Track cache hit rate

- **WHEN** 系统使用 Redis 缓存权限检查结果
- **THEN** 监控系统记录缓存命中率

### Requirement: The system MUST provide OpenFGA backup and restore

系统 MUST 提供工具备份和恢复 OpenFGA 数据。

#### Scenario: Backup relation tuples

- **WHEN** 管理员执行备份命令
- **THEN** 系统导出所有关系元组到 JSON 文件

#### Scenario: Restore relation tuples

- **WHEN** 管理员执行恢复命令并提供备份文件
- **THEN** 系统导入关系元组到 OpenFGA

#### Scenario: Backup includes model version

- **WHEN** 备份被创建
- **THEN** 备份文件必须包含 authorization_model_id 和模型定义

### Requirement: The system MUST support migration shadow validation

系统 MUST 支持迁移期间的影子校验（shadow mode），用于比较旧系统与 IAM/OpenFGA 判定结果一致性。

#### Scenario: Shadow mode compares decisions

- **WHEN** 迁移处于双写/影子阶段
- **THEN** 系统同时记录旧链路与新链路的授权结果并输出差异报告

#### Scenario: Shadow mismatch blocks cutover

- **WHEN** 影子校验差异率超过预设阈值
- **THEN** 系统阻止切换并标记迁移失败
