## MODIFIED Requirements

### Requirement: Kafka configuration in Data message
`api/protos/servora/conf/v1/conf.proto` 的 `Data` message SHALL 包含 `Kafka kafka` 字段，`Data.Kafka` message SHALL 包含以下字段：
- `repeated string brokers` — Kafka broker 地址列表
- `string client_id` — 客户端标识
- `string consumer_group` — 消费者组 ID
- `int32 required_acks` — 生产确认级别
- `int32 retry_max` — 最大重试次数
- `google.protobuf.Duration retry_backoff` — 重试退避间隔
- `google.protobuf.Duration dial_timeout` — 连接超时
- `google.protobuf.Duration read_timeout` — 读超时
- `google.protobuf.Duration write_timeout` — 写超时
- `string compression` — 压缩算法（none/gzip/snappy/lz4/zstd）
- `KafkaSASL sasl` — 可选 SASL 认证配置

`KafkaSASL` message SHALL 包含 `string mechanism`、`string username`、`string password`。

The proto SHALL declare `package servora.conf.v1;` and SHALL live in a directory matching `servora/conf/v1`.

#### Scenario: Kafka config proto compiles

- **WHEN** `make api` is run after adding Kafka config
- **THEN** `conf.Data_Kafka` Go type SHALL be generated and importable

#### Scenario: Kafka broker uses proto config

- **WHEN** `pkg/broker/kafka.NewBroker` is called with `*conf.Data_Kafka`
- **THEN** the Kafka client SHALL be configured with the specified brokers, timeouts, compression, and SASL settings

#### Scenario: Kafka not configured gracefully skips

- **WHEN** `Data.kafka` is nil or has empty `brokers`
- **THEN** `NewBrokerOptional` SHALL return nil and log an info message, without returning an error

### Requirement: ClickHouse configuration in Data message
`Data` message SHALL 包含 `ClickHouse clickhouse` 字段，`Data.ClickHouse` message SHALL 包含：
- `repeated string addrs` — ClickHouse 地址列表
- `string database` — 数据库名
- `string username` — 用户名
- `string password` — 密码
- `google.protobuf.Duration dial_timeout` — 连接超时
- `google.protobuf.Duration read_timeout` — 读超时
- `int32 max_open_conns` — 最大打开连接数
- `int32 max_idle_conns` — 最大空闲连接数
- `google.protobuf.Duration conn_max_lifetime` — 连接最大生命周期

The proto SHALL declare `package servora.conf.v1;` and SHALL live in a directory matching `servora/conf/v1`.

#### Scenario: ClickHouse config proto compiles

- **WHEN** `make api` is run after adding ClickHouse config
- **THEN** `conf.Data_ClickHouse` Go type SHALL be generated and importable

### Requirement: Audit configuration in App message
`App` message SHALL 包含 `Audit audit` 字段，`App.Audit` message SHALL 包含：
- `bool enabled` — 审计功能开关
- `string emitter_type` — emitter 类型（"broker" / "log" / "noop"）
- `string topic` — 审计事件 Kafka topic（默认 "servora.audit.events"）
- `string service_name` — 覆盖 App.name 作为审计事件中的服务标识

The proto SHALL declare `package servora.conf.v1;` and SHALL live in a directory matching `servora/conf/v1`.

#### Scenario: Audit enabled with broker emitter

- **WHEN** config has `audit.enabled: true` and `audit.emitter_type: "broker"`
- **THEN** `NewRecorderOptional` SHALL create a `BrokerEmitter` publishing to the configured topic

#### Scenario: Audit enabled with log emitter

- **WHEN** config has `audit.enabled: true` and `audit.emitter_type: "log"`
- **THEN** `NewRecorderOptional` SHALL create a `LogEmitter` writing to the framework logger

#### Scenario: Audit disabled

- **WHEN** config has `audit.enabled: false` or `audit` is nil
- **THEN** `NewRecorderOptional` SHALL create a `NoopEmitter`
