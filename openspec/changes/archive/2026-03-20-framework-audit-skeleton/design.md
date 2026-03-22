## Context

Servora 当前具备可运行的 `pkg/actor`（精简 interface）、`pkg/authz`（OpenFGA middleware）、`pkg/authn`（JWT 验证）、`pkg/openfga`（客户端适配），以及 `IdentityFromHeader` middleware（仅读取 `X-User-ID`）。但要支撑后续的 Keycloak 接入、分布式审计管线和 all-in-proto 审计注解，需要先将框架核心抽象升级到位。

参考来源：
- [`kratos-transport`](/Users/horonlee/projects/go/kratos-transport) `/broker/`：Broker、Event、Message、Subscriber 接口设计
- [`Kemate`](/Users/horonlee/projects/go/Kemate) `/pkg/kafka/`：Sarama 同步 producer 封装、Config 映射
- [`Kemate`](/Users/horonlee/projects/go/Kemate) `/app/audit/service/`：consumer → data → service 分层结构
- Master design doc 阶段 1 目标

本阶段覆盖 master design doc **阶段 1（定骨架）+ 阶段 2 基础设施引入**，为后续阶段提供所有接口与基础设施前置依赖。

## Goals / Non-Goals

**Goals:**

- 将 `Actor` interface 升级为通用 principal 模型，一次性完成破坏性变更
- 建立 `pkg/broker` 最小消息代理抽象 + `pkg/broker/kafka` Kafka 实现
- 建立 `pkg/audit` 审计运行时骨架（事件模型、emitter、recorder、middleware）
- 定义 `audit/v1` proto（公共事件模型 + 审计注解）
- 增强 `IdentityFromHeader` 以支持多 gateway header
- 引入 Kafka + ClickHouse 到 docker-compose 基础设施
- 从工具链中解耦 IAM / sayhello 的开发与启动条目

**Non-Goals:**

- 不实现 `protoc-gen-servora-audit` 代码生成器
- 不创建 `app/audit/service`
- 不接入 Keycloak 或改造 Traefik 认证链路
- 不在 `pkg/authz` 中接入审计事件采集
- 不实现 broker 的 NATS / RabbitMQ / Redis Streams 等其他实现

## Decisions

### D0: pkg/logger — 暴力重构

**决策**: 彻底重构 `pkg/logger`，消除所有 API 痛点，作为本阶段第一个任务（其余组件依赖 logger）。

**当前痛点分析**:

```
痛点                                    频次    严重性
────────────────────────────────────────────────────────
logger.NewHelper(l, logger.WithModule("x/y/z"))  ~20处  极度冗长
logger.Config 手动映射 conf.App_Log              1处    冗余转换
Sync func() error 字段 (exported, 未使用)        0调用  死代码+封装泄漏
NewLogger(nil) → panic                          潜在    bug
prod/default 分支 20 行重复代码                   1处    维护负担
GORM/Ent bridge 需要类型断言拿 *zap.Logger       2处    不优雅
```

**重构后 API 对比**:

```
Before                                          After
──────────────────────────────────────────────────────────────────
logger.NewLogger(&logger.Config{                logger.New(bc.App)
    Env: bc.App.Env,
    Level: appLog.GetLevel(),
    Filename: appLog.GetFilename(),
    MaxSize: appLog.GetMaxSize(),
    MaxBackups: appLog.GetMaxBackups(),
    MaxAge: appLog.GetMaxAge(),
    Compress: appLog.GetCompress(),
})

logger.NewHelper(l,                             logger.For(l, "user/data/iam")
    logger.WithModule("user/data/iam-service"))

logger.With(l,                                  logger.With(l, "http/server/iam")
    logger.WithModule("http/server/iam-service"))

zapLogger.Sync (field, never called)            zapLogger.Sync() (method)
                                                zapLogger.Zap() *zap.Logger (NEW)
```

**破坏性变更清单**:

| 变更 | 影响范围 | 迁移方式 |
|------|---------|---------|
| 移除 `logger.Config` struct | `pkg/bootstrap` (1 处) | 改用 `logger.New(app)` |
| `NewLogger` → `New` | `pkg/bootstrap` (1 处) | 重命名 |
| `New` 返回 `*ZapLogger` 而非 `log.Logger` | `pkg/bootstrap` (1 处) | 类型明确化，更利于调用 `Zap()`/`Sync()` |
| 移除 `Sync` 字段 → `Sync()` 方法 | 无调用方 | 无迁移 |
| `WithModule(m)` → `For(l, m)` (shorthand) | ~20 处 `NewHelper` 调用 | 全量替换为 `For` |
| `With(l, WithModule(m))` 简化 | ~5 处 | 改为 `With(l, m)` 重载 |

**新 API 设计**:

```go
// ==================== 创建 ====================

// New 从 proto 配置创建 ZapLogger。
// 直接接收 *conf.App（读取 .Env 和 .Log），nil-safe。
// 返回 *ZapLogger（而非 log.Logger），便于调用 Zap()/Sync()。
func New(app *conf.App) *ZapLogger

// ==================== 模块 Helper（最常用）====================

// For 创建带 module 标识的 Helper —— 一行替代 NewHelper+WithModule。
//   Before: logger.NewHelper(l, logger.WithModule("user/data/iam-service"))
//   After:  logger.For(l, "user/data/iam")
func For(l Logger, module string) *Helper

// ==================== 结构化字段 ====================

// With 添加结构化字段到 logger。支持两种调用风格：
//   logger.With(l, "http/server/iam")              — 快捷 module
//   logger.With(l, WithModule("x"), WithField(...)) — 原始 Option
func With(l Logger, opts ...Option) Logger

// WithModule, WithField — 保留，供 With 的 Option 风格使用
func WithModule(module string) Option
func WithField(key string, value any) Option

// NewHelper — 保留，供需要额外 Option 的场景
func NewHelper(l Logger, opts ...Option) *Helper

// ==================== ZapLogger 方法 ====================

func (l *ZapLogger) Zap() *zap.Logger    // 暴露底层 zap（供 kzap、GORM bridge 等）
func (l *ZapLogger) Sync() error         // 方法替代字段
func (l *ZapLogger) Log(level, keyvals...) error  // 不变
func (l *ZapLogger) GetGormLogger(module string) GormLogger  // 不变
```

**内部实现优化**:

1. **nil config 安全**: `New(nil)` 返回 dev 模式 console logger（不 panic）
2. **提取 `buildCore`**: prod/default 共享 core 构建逻辑，消除 20 行重复
3. **GORM/Ent bridge 使用 `Zap()` getter**: 消除内部对 unexported `log` 字段的直接访问
4. **module 命名简化约定**: 去掉 `-service` 后缀（如 `"user/data/iam-service"` → `"user/data/iam"`），更简洁

### D1: Actor v2 — 破坏性接口扩展

**决策**: 直接扩展 `Actor` interface，新增方法而非通过 `Scope(key)` 透传。

**理由**:

- 框架优先：Actor 是 Servora 框架的核心类型，应有明确的类型安全字段
- `Subject()`、`Email()`、`Roles()`、`Scopes()` 是 actor 的一等公民属性，不应降级为 string key 查找
- 一次性做到位，避免后续再次 break

**替代方案**: 保持现有 interface 不变，新字段走 `Scope(key)` — 被否决，因为 Scope 语义是"请求级维度"（tenant/org/project），不适合承载身份固有属性

**新 Actor interface 设计**:

```go
type Actor interface {
    ID() string
    Type() Type
    DisplayName() string
    Email() string
    Subject() string       // 外部 IdP subject (e.g. Keycloak sub)
    ClientID() string      // OAuth2 client_id
    Realm() string         // IdP realm / tenant namespace
    Roles() []string       // 角色列表
    Scopes() []string      // OAuth2 scopes
    Attrs() map[string]string  // 扩展属性 (开放 bag)
    Scope(key string) string   // 保留：请求级维度 (tenant/org/project)
}
```

**新增 actor 类型**:

- `ServiceActor`：service-to-service 调用身份（Type = "service"）

**Breaking change 迁移**:

```
影响范围:
├── pkg/actor/           → interface 签名变更、UserActor struct 扩展、新增 ServiceActor
├── pkg/authn/           → defaultClaimsMapper 适配新 Actor 字段
├── pkg/authz/           → 仅依赖 Actor.ID() 和 Actor.Type()，无需改动
├── pkg/transport/       → IdentityFromHeader 适配多 header → 新 Actor 构造
├── app/iam/service/     → 保留代码但不再在工具链中活跃，需要编译通过
└── app/sayhello/service → 同上
```

### D2: pkg/broker — Servora 自有消息代理抽象

**决策**: 参考 [kratos-transport](/Users/horonlee/projects/go/kratos-transport) broker 接口风格，在 Servora 内部建立自有 `pkg/broker` 生态，不直接依赖 kratos-transport。

**理由**:

- Servora 需要控制自己的核心接口演进节奏
- kratos-transport 设计偏重，包含了 Request/Response、Binder 等 Servora 暂不需要的语义
- 自有接口可以更好地与 `pkg/audit` 集成

**借鉴 [kratos-transport](/Users/horonlee/projects/go/kratos-transport) 的部分**:

- `Broker` interface 的 `Connect/Disconnect` 生命周期
- `Message` 的 `Headers` + `Body` 结构
- `Event` 的 `Topic() + Message() + Ack()` 模式
- `Subscriber` 的 `Topic() + Unsubscribe()` 模式
- Option 函数式配置风格

**简化的部分**:

- 移除 `Binder`（类型绑定）— Servora 使用 proto 编解码，不需要运行时类型绑定
- 移除 `Request`（请求/响应）— broker 定位为事件总线，不做 RPC
- 移除 `Name()` / `Address()` — 使用更精简的接口
- `Any` 改为 `[]byte` — body 统一为序列化字节，编解码由上层负责

**Servora broker 最小接口**:

```go
// pkg/broker/broker.go
type Broker interface {
    Connect(ctx context.Context) error
    Disconnect(ctx context.Context) error
    Publish(ctx context.Context, topic string, msg *Message, opts ...PublishOption) error
    Subscribe(ctx context.Context, topic string, handler Handler, opts ...SubscribeOption) (Subscriber, error)
}

type Message struct {
    Key     string
    Headers Headers
    Body    []byte
}

type Headers map[string]string

type Event interface {
    Topic() string
    Message() *Message
    Ack() error
    Nack() error
}

type Handler func(ctx context.Context, event Event) error

type Subscriber interface {
    Topic() string
    Options() SubscribeOptions
    Unsubscribe(removeFromManager bool) error
}
```

**Kafka 实现**: `pkg/broker/kafka/`

- 使用 **franz-go** 而非 Sarama（Sarama 已标记维护模式，franz-go 更现代）
- 参考 [Kemate](/Users/horonlee/projects/go/Kemate) `pkg/kafka` 的 Config 结构，但适配到 Servora broker 抽象上
- 支持 producer + consumer group
- 接收 `conf.Data_Kafka` proto 配置（与 OpenFGA 接收 `conf.App_OpenFGA` 模式一致）

**franz-go 插件集成**:

franz-go 提供 [plugin 生态](https://github.com/twmb/franz-go/tree/master/plugin)，Servora 将使用以下两个：

- **kzap**（`github.com/twmb/franz-go/plugin/kzap`）：将 Kafka 客户端内部日志桥接到 Servora 的 zap logger
- **kotel**（`github.com/twmb/franz-go/plugin/kotel`）：将 Kafka produce/consume 操作自动注入 OTel tracing + metrics

集成方式：

```go
import (
    "github.com/twmb/franz-go/pkg/kgo"
    "github.com/twmb/franz-go/plugin/kzap"
    "github.com/twmb/franz-go/plugin/kotel"
)

func newKafkaClient(cfg *conf.Data_Kafka, zapLogger *zap.Logger) (*kgo.Client, error) {
    ktracer := kotel.NewTracer()
    kmeter  := kotel.NewMeter()
    kotelSvc := kotel.NewKotel(kotel.WithTracer(ktracer), kotel.WithMeter(kmeter))

    opts := []kgo.Opt{
        kgo.SeedBrokers(cfg.Brokers...),
        kgo.WithLogger(kzap.New(zapLogger)),
        kgo.WithHooks(kotelSvc.Hooks()...),
    }
    return kgo.NewClient(opts...)
}
```

这要求 `pkg/logger.ZapLogger` 暴露底层 `*zap.Logger`：新增 `Zap() *zap.Logger` getter 方法。
此 getter 同时服务于未来任何需要原始 zap 实例的框架组件（GORM bridge 已有类似模式）。

### D3: pkg/audit — 审计运行时骨架

**决策**: 审计运行时与 broker 解耦，通过 `Emitter` 接口抽象事件发送。

**架构**:

```
┌──────────────────────────────────────────────────────────────────┐
│                        业务服务 / middleware                       │
│                                                                  │
│   ┌──────────┐    ┌───────────┐    ┌──────────────┐             │
│   │ authz    │    │ tuple     │    │ resource     │             │
│   │ middleware│    │ write/del │    │ mutation     │             │
│   └────┬─────┘    └─────┬─────┘    └──────┬───────┘             │
│        │                │                  │                     │
│        └────────────────┼──────────────────┘                     │
│                         ▼                                        │
│              ┌─────────────────────┐                             │
│              │   pkg/audit.Recorder│  ← 汇聚、构建、发送          │
│              │   (with Emitter)    │                             │
│              └──────────┬──────────┘                             │
│                         ▼                                        │
│              ┌─────────────────────┐                             │
│              │   Emitter interface │  ← 抽象发送接口              │
│              │   ├── BrokerEmitter │  ← 生产实现（→ Kafka topic） │
│              │   ├── LogEmitter    │  ← 开发/调试用               │
│              │   └── NoopEmitter   │  ← 测试/禁用审计             │
│              └─────────────────────┘                             │
└──────────────────────────────────────────────────────────────────┘
```

**审计事件模型** (Go runtime, 对应 proto 定义):

```go
type AuditEvent struct {
    EventID    string
    EventType  EventType       // authn.result | authz.decision | authz.tuple.changed | resource.mutation
    Version    string          // 事件版本
    OccurredAt time.Time
    Service    string
    Operation  string
    Actor      ActorInfo       // 快照，非引用
    Target     TargetInfo
    Result     ResultInfo
    TraceID    string
    RequestID  string
    Detail     any             // typed: AuthnDetail | AuthzDetail | TupleMutationDetail | ResourceMutationDetail
}
```

**Emitter 接口**:

```go
type Emitter interface {
    Emit(ctx context.Context, event *AuditEvent) error
    Close() error
}
```

**BrokerEmitter**: 将 `AuditEvent` proto-marshal 后通过 `pkg/broker` 发送到审计 topic。

### D4: audit.proto — 审计事件公共模型

**决策**: proto 定义放在 `api/protos/servora/audit/v1/`，包含两个文件。

**`audit.proto`** — 公共事件模型:

```
┌──────────────────────────────────────────────┐
│ AuditEvent                                   │
├──────────────────────────────────────────────┤
│ event_id:     string (UUID)                  │
│ event_type:   AuditEventType (enum)          │
│ event_version:string                         │
│ occurred_at:  google.protobuf.Timestamp      │
│ service:      string                         │
│ operation:    string                         │
│ actor:        AuditActor                     │
│ target:       AuditTarget                    │
│ result:       AuditResult                    │
│ trace_id:     string                         │
│ request_id:   string                         │
│ detail:       oneof {                        │
│   authn_detail:           AuthnDetail        │
│   authz_detail:           AuthzDetail        │
│   tuple_mutation_detail:  TupleMutDetail     │
│   resource_mutation_detail:ResMutDetail      │
│ }                                            │
└──────────────────────────────────────────────┘
```

**`annotations.proto`** — RPC 级审计注解:

- 定义 `AuditRule` message 与 `audit_rule` method option
- 声明每个 RPC 是否产生审计事件、事件类型、target 提取规则
- 本阶段只做 proto 定义，不做 codegen 消费

### D5: IdentityFromHeader 增强

**决策**: 保持框架网关无关性，不绑定 Traefik 或任何特定网关。

**理由**:

- 作为框架，Servora 不应假定用户使用哪个网关
- `IdentityFromHeader` 的职责是"从 header 构建 Actor"，网关类型无关
- ForwardAuth vs OIDC middleware 是网关侧决策，不影响框架接口

**增强方案**:

- 支持配置多个 header key → Actor 字段的映射
- 默认映射: `X-User-ID` → ID, `X-Subject` → Subject, `X-Client-ID` → ClientID, `X-Realm` → Realm, `X-Email` → Email, `X-Roles` → Roles (逗号分隔), `X-Scopes` → Scopes (空格分隔), `X-Principal-Type` → Type

### D6: 基础设施 — Kafka + ClickHouse

**决策**: 参照 [Kemate](/Users/horonlee/projects/go/Kemate) docker-compose 模式引入。

**Kafka**: 使用 KRaft 模式（无 ZooKeeper），与 [Kemate](/Users/horonlee/projects/go/Kemate) 配置一致。

**ClickHouse**: 本阶段仅引入容器，不创建表结构（表结构随 `app/audit/service` 在阶段 2 落地）。

### D7: 工具链解耦 IAM/sayhello

**决策**: 保留服务代码，仅从工具链（Makefile、docker-compose.dev.yaml）中移除开发/启动条目。

**理由**:

- IAM 和 sayhello 作为完整的微服务参考标准，供后续新服务（如 audit）参考
- 代码保留确保 Actor v2 breaking change 后编译仍然通过（需适配）
- 从工具链移除避免开发者误启动已废弃服务

### D8: 配置层扩展 — conf.proto + pkg 集成

**决策**: 在 `api/protos/servora/conf/v1/conf.proto` 的 `Data` message 中新增 `Kafka` 和 `ClickHouse` 子 message，在 `App` message 中新增 `Audit` 子 message。所有新增 pkg 组件通过 proto 配置初始化。

**理由**:

- 与现有模式一致：`Data.Redis`、`Data.Database`、`App.OpenFGA`、`App.Log` 均使用 proto 配置
- 配置通过 proto → YAML → 代码生成类型，保证类型安全
- 让每个基础设施组件（Kafka、ClickHouse、Audit）都可通过配置文件开启/关闭/调参

**新增配置结构**:

```
Data (existing)
├── Database (existing)
├── Redis (existing)
├── Client (existing)
├── Kafka (NEW)
│   ├── brokers: repeated string
│   ├── client_id: string
│   ├── consumer_group: string
│   ├── required_acks: int32
│   ├── retry_max: int32
│   ├── retry_backoff: Duration
│   ├── dial_timeout: Duration
│   ├── read_timeout: Duration
│   ├── write_timeout: Duration
│   ├── compression: string (none/gzip/snappy/lz4/zstd)
│   └── sasl: KafkaSASL (optional)
│       ├── mechanism: string (PLAIN/SCRAM-SHA-256/SCRAM-SHA-512)
│       ├── username: string
│       └── password: string
└── ClickHouse (NEW)
    ├── addrs: repeated string
    ├── database: string
    ├── username: string
    ├── password: string
    ├── dial_timeout: Duration
    ├── read_timeout: Duration
    ├── max_open_conns: int32
    ├── max_idle_conns: int32
    └── conn_max_lifetime: Duration

App (existing)
├── ... (existing fields)
└── Audit (NEW)
    ├── enabled: bool
    ├── emitter_type: string (broker/log/noop)
    ├── topic: string (default: "servora.audit.events")
    └── service_name: string (override App.name if needed)
```

**pkg 初始化模式** — 与 `pkg/openfga` 的 `NewClientOptional` 模式一致：

```go
// pkg/broker/kafka/config.go
func NewBrokerOptional(cfg *conf.Data, l logger.Logger) broker.Broker {
    if cfg.Kafka == nil || len(cfg.Kafka.Brokers) == 0 {
        logger.For(l, "broker/kafka").Info("Kafka not configured, broker disabled")
        return nil
    }
    zapL := l.(*logger.ZapLogger).Zap()
    b, err := NewBroker(cfg.Kafka, zapL)
    if err != nil {
        logger.For(l, "broker/kafka").Warnf("failed to create Kafka broker: %v", err)
        return nil
    }
    return b
}

// pkg/audit/config.go
func NewRecorderOptional(cfg *conf.App, broker broker.Broker, l logger.Logger) *Recorder {
    if cfg.Audit == nil || !cfg.Audit.Enabled {
        return NewRecorder(NewNoopEmitter())
    }
    switch cfg.Audit.EmitterType {
    case "broker":
        return NewRecorder(NewBrokerEmitter(broker, cfg.Audit.Topic))
    case "log":
        return NewRecorder(NewLogEmitter(l))
    default:
        return NewRecorder(NewNoopEmitter())
    }
}
```

**Logger Zap getter**:

`pkg/logger/log.go` 新增一行：

```go
func (l *ZapLogger) Zap() *zap.Logger { return l.log }
```

此方法让任何需要原始 `*zap.Logger` 的框架组件都能获取，而不需要绕过封装。当前使用方：
- `pkg/broker/kafka`：通过 `kzap.New(zap)` 桥接 franz-go 日志
- `GetGormLogger`：已有类似模式（通过 `l.log.With(...)` 直接使用 zap）

## Risks / Trade-offs

**[Actor interface breaking change]** → 所有 Actor 实现需要适配新方法。
*Mitigation*: 本阶段一并更新 `pkg/authn`、`IdentityFromHeader`、IAM/sayhello 中的调用方，确保全部编译通过。

**[franz-go vs Sarama]** → franz-go 在 Servora 中是首次引入的新依赖。
*Mitigation*: franz-go 是 Go 社区推荐的现代 Kafka 客户端，API 更简洁，性能更好。[Kemate](/Users/horonlee/projects/go/Kemate) 用 Sarama 但 Sarama 已处于维护模式。

**[审计骨架无消费端]** → 本阶段 `pkg/audit` 只有 emitter 无 consumer，无法端到端验证。
*Mitigation*: 提供 `LogEmitter` 和 `NoopEmitter` 用于本地验证；完整管线在阶段 2 落地。

**[proto 注解定义先行]** → `annotations.proto` 定义了注解但无 codegen 消费。
*Mitigation*: 注解 proto 可独立定义和迭代，codegen 在阶段 4 实现。本阶段确保注解设计合理即可。

**[ClickHouse 无表结构]** → 仅容器运行，无实际用途。
*Mitigation*: 为阶段 2 `app/audit/service` 预置基础设施，降低后续启动成本。

**[暴露 Zap() getter]** → 打破了 ZapLogger 对底层 zap 实例的封装。
*Mitigation*: 这是必要的逃逸口——GORM bridge 已有类似模式（直接操作 `l.log`）。getter 方法让组件（如 franz-go kzap）可以注入同一 zap 实例，保持日志输出统一。

**[franz-go plugin 依赖增加]** → kzap + kotel 各自引入额外 go module。
*Mitigation*: 这些是 franz-go 官方维护的轻量 plugin（无额外外部依赖），版本与 franz-go 主库同步。

## Open Questions

1. **franz-go consumer group 配置**: 是否在本阶段的 `pkg/broker/kafka` 中完整实现 consumer group 逻辑，还是仅实现 producer + 基础 consumer？
   *倾向*: 实现完整的 producer + consumer group，因为阶段 2 马上需要。

2. **审计 topic 命名约定**: 是否所有事件类型走同一个 topic（`servora.audit.events`），还是按类型分 topic（`servora.audit.authz.decision` 等）？
   *倾向*: 统一 topic + event_type 字段区分，简化消费端逻辑。

3. **Actor.Roles() 返回类型**: `[]string` vs 自定义 `Role` 类型？
   *倾向*: `[]string`，保持简单。角色的复杂语义（层级、权限等）由上层业务或 OpenFGA 处理。
