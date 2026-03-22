## 0. pkg/logger — 暴力重构

- [x] 0.1 重写 `pkg/logger/log.go` 核心：移除 `Config` struct 和 `NewLogger`；新增 `New(app *conf.App) *ZapLogger`，nil-safe，直接读取 `app.Env` 和 `app.GetLog()` 字段
- [x] 0.2 提取 `buildCore(env, level string, writers ...zapcore.WriteSyncer) zapcore.Core`：消除 prod/default 20 行重复代码，统一 encoder 和 level 配置
- [x] 0.3 移除 `ZapLogger.Sync` 字段 → 新增 `Sync() error` 方法，委托至 `l.log.Sync()`
- [x] 0.4 新增 `Zap() *zap.Logger` getter 方法，暴露底层 zap 实例
- [x] 0.5 新增 `For(l log.Logger, module string) *log.Helper` 快捷方法：一行创建带 module 标识的 Helper
- [x] 0.6 简化 `With(l log.Logger, module string) log.Logger` 重载：支持直接传入 module string
- [x] 0.7 更新 `GetGormLogger` / `GormLoggerFrom` / `EntLogFuncFrom`：内部使用 `Zap()` getter 替代直接访问 unexported `log` 字段
- [x] 0.8 迁移 `pkg/bootstrap/bootstrap.go`：`NewLogger(&Config{...})` → `New(bc.App)`，移除手动字段映射
- [x] 0.9 全量迁移 `app/iam/service/**/*.go`（~12 处）：`NewHelper(l, WithModule("x"))` → `For(l, "x")`；`With(l, WithModule("x"))` → `With(l, "x")`；module 命名去掉 `-service` 后缀
- [x] 0.10 全量迁移 `app/sayhello/service/**/*.go`（~1 处）：同上
- [x] 0.11 全量迁移 `pkg/**/*.go`（redis、openfga、transport、jwks、governance）（~6 处）：同上
- [x] 0.12 更新测试：适配 `log_defaults_test.go` 和 `gorm_log_test.go`；新增 `New(nil)` 安全性、`For`、`Zap()`、`Sync()` 方法测试
- [x] 0.13 更新 `pkg/logger/AGENTS.md`：反映新 API 和使用示例
- [x] 0.14 运行 `go build ./...` + `go test ./pkg/logger/...` 验证编译和测试通过

## 1. Actor v2 — 破坏性接口升级

- [x] 1.1 扩展 `pkg/actor/actor.go` 的 `Actor` interface：新增 `Subject()`, `ClientID()`, `Realm()`, `Email()`, `Roles()`, `Scopes()`, `Attrs()` 方法；新增 `TypeService Type = "service"` 常量
- [x] 1.2 重构 `pkg/actor/user.go` 的 `UserActor` struct：新增对应字段和 getter，更新 `NewUserActor` 构造函数签名
- [x] 1.3 新建 `pkg/actor/service.go`：实现 `ServiceActor` struct（Type="service"，携带 ID、ClientID、DisplayName）
- [x] 1.4 适配 `pkg/authn/authn.go`：更新 `defaultClaimsMapper` 以填充 Actor v2 新字段（subject、email、roles 等）；更新 `NewUserActor` 调用签名
- [x] 1.5 适配 `pkg/transport/server/middleware/identity.go`：更新 `NewUserActor` 调用签名以兼容新构造函数
- [x] 1.6 适配 `app/iam/service` 内所有 `actor.NewUserActor` 和 `Actor` interface 使用：确保编译通过
- [x] 1.7 适配 `app/sayhello/service` 内所有 actor 相关调用：确保编译通过
- [x] 1.8 运行 `go build ./...` 验证所有 workspace 模块编译通过

## 2. IdentityFromHeader 增强

- [x] 2.1 扩展 `pkg/transport/server/middleware/identity.go`：支持从多个 gateway header 构建 Actor v2（X-User-ID, X-Subject, X-Client-ID, X-Realm, X-Email, X-Roles, X-Scopes, X-Principal-Type）
- [x] 2.2 新增 `WithHeaderMapping` option：允许自定义 header key → Actor 字段映射
- [x] 2.3 实现 `X-Principal-Type` 分支逻辑：`"service"` 构造 `ServiceActor`，`"user"` 或默认构造 `UserActor`
- [x] 2.4 更新 `pkg/transport/server/middleware/identity_test.go`：覆盖多 header、单 header、无 header、service principal 场景

## 3. Proto 定义 — Audit + 配置扩展

- [x] 3.1 新建 `api/protos/servora/audit/v1/audit.proto`：定义 `AuditEvent`、`AuditEventType` enum、`AuditActor`、`AuditTarget`、`AuditResult`、typed detail messages（AuthnDetail、AuthzDetail、TupleMutationDetail、ResourceMutationDetail）
- [x] 3.2 新建 `api/protos/servora/audit/v1/annotations.proto`：定义 `AuditRule` message 和 `audit_rule` method option extension
- [x] 3.3 扩展 `api/protos/servora/conf/v1/conf.proto` — `Data` message 新增 `Kafka kafka` 字段：定义 `Data.Kafka` message（brokers、client_id、consumer_group、required_acks、retry_max、retry_backoff、dial/read/write_timeout、compression）和 `KafkaSASL` message（mechanism、username、password）
- [x] 3.4 扩展 `api/protos/servora/conf/v1/conf.proto` — `Data` message 新增 `ClickHouse clickhouse` 字段：定义 `Data.ClickHouse` message（addrs、database、username、password、dial_timeout、read_timeout、max_open_conns、max_idle_conns、conn_max_lifetime）
- [x] 3.5 扩展 `api/protos/servora/conf/v1/conf.proto` — `App` message 新增 `Audit audit` 字段：定义 `App.Audit` message（enabled、emitter_type、topic、service_name）
- [x] 3.6 更新 `buf.yaml`：确认新的 `audit/v1` proto 路径被正确纳入 module
- [x] 3.7 运行 `make api`：验证所有 proto 编译通过，`api/gen/go/servora/audit/v1/` 和 `api/gen/go/servora/conf/v1/` 生成产物正确 ⚡ 需要 `make api`

## 4. pkg/broker — 消息代理抽象

- [x] 4.1 新建 `pkg/broker/broker.go`：定义 `Broker` interface（Connect、Disconnect、Publish、Subscribe）
- [x] 4.2 新建 `pkg/broker/message.go`：定义 `Message` struct（Key、Headers、Body）、`Headers` type alias、`Event` interface、`Handler` type、`Subscriber` interface
- [x] 4.3 新建 `pkg/broker/options.go`：定义 `PublishOption`、`SubscribeOption` 及常用 option 函数
- [x] 4.4 新建 `pkg/broker/kafka/` 目录：实现 `kafkaBroker` struct，基于 franz-go，接收 `*conf.Data_Kafka` proto 配置
- [x] 4.5 实现 `pkg/broker/kafka/broker.go`：Connect/Disconnect 生命周期，内部持有 `*kgo.Client`；初始化时注入 `kzap.New(zap)` 日志桥接 + `kotel` OTel hooks
- [x] 4.6 实现 `pkg/broker/kafka/producer.go`：同步 producer，支持 key 分区、header 透传
- [x] 4.7 实现 `pkg/broker/kafka/consumer.go`：consumer group，支持 handler 回调、Ack/Nack
- [x] 4.8 实现 `pkg/broker/kafka/config.go`：`NewBroker(cfg *conf.Data_Kafka, zapLogger *zap.Logger)` 工厂函数 + `NewBrokerOptional(cfg *conf.Data, l logger.Logger)` 可选初始化（遵循 openfga.NewClientOptional 模式）
- [x] 4.9 添加 franz-go 依赖到根 `go.mod`：`go get github.com/twmb/franz-go/pkg/kgo github.com/twmb/franz-go/plugin/kzap github.com/twmb/franz-go/plugin/kotel`

## 5. pkg/audit — 审计运行时骨架

- [x] 5.1 新建 `pkg/audit/event.go`：定义 `AuditEvent` Go struct、`EventType` 常量、`ActorInfo`、`TargetInfo`、`ResultInfo`、typed detail structs
- [x] 5.2 新建 `pkg/audit/emitter.go`：定义 `Emitter` interface（Emit、Close）
- [x] 5.3 新建 `pkg/audit/broker_emitter.go`：实现 `BrokerEmitter`，将 AuditEvent proto-marshal 后通过 `pkg/broker` 发送到可配置的审计 topic
- [x] 5.4 新建 `pkg/audit/log_emitter.go`：实现 `LogEmitter`，将事件 JSON 序列化后写入 logger
- [x] 5.5 新建 `pkg/audit/noop_emitter.go`：实现 `NoopEmitter`，静默丢弃
- [x] 5.6 新建 `pkg/audit/recorder.go`：实现 `Recorder`，提供 `RecordAuthzDecision`、`RecordTupleChange`、`RecordResourceMutation`、`RecordAuthnResult` 方法；自动填充 EventID、OccurredAt、TraceID、RequestID
- [x] 5.7 新建 `pkg/audit/config.go`：实现 `NewRecorderOptional(cfg *conf.App, broker broker.Broker, l logger.Logger)` 可选初始化，根据 `cfg.Audit.EmitterType` 选择 emitter 实现
- [x] 5.8 新建 `pkg/audit/middleware.go`：实现 Kratos audit middleware 骨架（option types、operation rule map、post-handler 记录逻辑）

## 6. 基础设施 — Kafka + ClickHouse

- [x] 6.1 在 `docker-compose.yaml` 中新增 `kafka` 服务：apache/kafka KRaft 模式，参照 Kemate 配置（环境变量、健康检查、volume、网络）
- [x] 6.2 在 `docker-compose.yaml` 中新增 `clickhouse` 服务：clickhouse-server，端口 18123/19000，健康检查，volume
- [x] 6.3 在 `docker-compose.yaml` volumes 段新增 `servora-kafka-data` 和 `servora-clickhouse-data`
- [x] 6.4 验证 `docker compose up kafka clickhouse` 启动正常且健康检查通过

## 7. 工具链解耦 IAM / sayhello

- [x] 7.1 清理 `docker-compose.dev.yaml`：移除 `iam` 和 `sayhello` service 定义
- [x] 7.2 更新 `Makefile`：将 `MICROSERVICES` 变量置空或注释；将 `GO_WORKSPACE_MODULES` 中移除 `app/iam/service` 和 `app/sayhello/service`（lint 不再覆盖废弃服务）；调整 `COMPOSE_STACK_SERVICES` 不再包含 MICROSERVICES
- [x] 7.3 验证 `make compose.dev` 不启动 IAM/sayhello 容器
- [x] 7.4 验证 `app/iam/service` 和 `app/sayhello/service` 代码仍然可独立编译（`go build ./app/iam/service/...`）

## 8. 文档与收尾

- [x] 8.1 更新 `pkg/AGENTS.md`：新增 `broker/`、`audit/` 模块描述；更新 `actor/` 描述反映 v2 变更
- [x] 8.2 更新根 `AGENTS.md` / `CLAUDE.md`：在模块速览中新增 `pkg/broker`、`pkg/audit`；更新基础设施列表添加 Kafka、ClickHouse
- [x] 8.3 运行 `make api` 确认所有 proto 生成正常 ⚡ 需要 `make api`
- [x] 8.4 运行 `go build ./...` 全量编译验证
