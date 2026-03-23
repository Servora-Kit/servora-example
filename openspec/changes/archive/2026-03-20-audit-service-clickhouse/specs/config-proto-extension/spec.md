## MODIFIED Requirements

### Requirement: Audit configuration in App message

`App` message SHALL 包含 `Audit audit` 字段，`App.Audit` message SHALL 包含：
- `bool enabled` — 审计功能开关
- `string emitter_type` — emitter 类型（"broker" / "log" / "noop"）
- `string topic` — 审计事件 Kafka topic（默认 "servora.audit.events"）
- `string service_name` — 覆盖 App.name 作为审计事件中的服务标识
- `int32 consumer_batch_size` — consumer 批量写入大小（默认 100）
- `google.protobuf.Duration consumer_flush_interval` — consumer 批量刷新间隔（默认 1s）
- `int32 retention_days` — ClickHouse 数据保留天数（默认 90）

#### Scenario: Audit enabled with broker emitter

- **WHEN** config has `audit.enabled: true` and `audit.emitter_type: "broker"`
- **THEN** `NewRecorderOptional` SHALL create a `BrokerEmitter` publishing to the configured topic

#### Scenario: Audit enabled with log emitter

- **WHEN** config has `audit.enabled: true` and `audit.emitter_type: "log"`
- **THEN** `NewRecorderOptional` SHALL create a `LogEmitter` writing to the framework logger

#### Scenario: Audit disabled

- **WHEN** config has `audit.enabled: false` or `audit` is nil
- **THEN** `NewRecorderOptional` SHALL create a `NoopEmitter`

#### Scenario: Consumer batch config applies to audit service

- **WHEN** config has `audit.consumer_batch_size: 200` and `audit.consumer_flush_interval: 2s`
- **THEN** the audit service BatchWriter SHALL use batch size 200 and flush interval 2s

#### Scenario: Retention days config applies to DDL

- **WHEN** config has `audit.retention_days: 30`
- **THEN** the ClickHouse DDL SHALL use `TTL occurred_at + INTERVAL 30 DAY`

#### Scenario: Default values when consumer fields unset

- **WHEN** `consumer_batch_size`, `consumer_flush_interval`, and `retention_days` are not set (zero values)
- **THEN** the audit service SHALL use defaults: batch_size=100, flush_interval=1s, retention_days=90
