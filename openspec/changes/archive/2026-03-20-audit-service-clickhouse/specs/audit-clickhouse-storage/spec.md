## ADDED Requirements

### Requirement: ClickHouse audit_events table DDL is auto-created at startup

The audit service SHALL execute a `CREATE TABLE IF NOT EXISTS audit_events` DDL at startup using the official ClickHouse native driver (`github.com/ClickHouse/clickhouse-go/v2`). The DDL SHALL be idempotent and embedded in Go code (not Docker init scripts).

#### Scenario: First startup creates table

- **WHEN** the audit service starts for the first time against an empty ClickHouse database
- **THEN** the `audit_events` table SHALL be created with the defined schema

#### Scenario: Subsequent startup is idempotent

- **WHEN** the audit service restarts against a ClickHouse database that already has `audit_events`
- **THEN** no error SHALL occur and the existing table SHALL remain unchanged

### Requirement: ClickHouse schema uses pure JSON for detail column

The `audit_events` table SHALL store the `detail` column as a `String` type containing JSON-serialized proto `oneof detail` content. The table SHALL NOT flatten detail fields into separate columns.

The table schema SHALL include:
- `event_id String` — UUID
- `event_type LowCardinality(String)` — enum name
- `event_version String`
- `occurred_at DateTime64(3, 'UTC')`
- `service LowCardinality(String)`
- `operation String`
- `actor_id String`, `actor_type LowCardinality(String)`, `actor_display_name String`
- `target_type LowCardinality(String)`, `target_id String`, `target_name String`
- `success Bool`, `error_code String`, `error_message String`
- `trace_id String`, `request_id String`
- `detail String` — JSON string

#### Scenario: Detail stored as JSON string

- **WHEN** an `AuditEvent` with `AuthzDetail` is written to ClickHouse
- **THEN** the `detail` column SHALL contain a valid JSON string representing the AuthzDetail fields

#### Scenario: Detail column is queryable with JSON functions

- **WHEN** `SELECT JSONExtractString(detail, 'relation') FROM audit_events WHERE event_type = 'AUTHZ_DECISION'` is executed
- **THEN** the query SHALL return the relation value from the AuthzDetail JSON

### Requirement: ClickHouse table uses daily partitioning and configurable TTL

The table SHALL use `PARTITION BY toDate(occurred_at)` for daily partitioning and `TTL occurred_at + INTERVAL N DAY` where N is read from `conf.App.Audit.retention_days` (default 90).

The sort order SHALL be `ORDER BY (service, event_type, occurred_at, event_id)`.

#### Scenario: Default TTL is 90 days

- **WHEN** `retention_days` is not configured or set to 0
- **THEN** the DDL SHALL use `TTL occurred_at + INTERVAL 90 DAY`

#### Scenario: Custom TTL is applied

- **WHEN** `retention_days` is configured as 30
- **THEN** the DDL SHALL use `TTL occurred_at + INTERVAL 30 DAY`

### Requirement: BatchWriter flushes to ClickHouse with configurable thresholds

The BatchWriter SHALL buffer incoming `AuditEvent` records in memory and flush to ClickHouse when either condition is met:
- Buffer size reaches `consumer_batch_size` (default 100)
- Time since last flush exceeds `consumer_flush_interval` (default 1s)

Flush SHALL use ClickHouse native driver's `PrepareBatch` + `Append` + `Send` API.

#### Scenario: Flush on batch size threshold

- **WHEN** 100 events accumulate in the buffer (with default batch_size=100)
- **THEN** the BatchWriter SHALL immediately flush all 100 events to ClickHouse in a single batch

#### Scenario: Flush on time threshold

- **WHEN** 10 events accumulate and 1 second passes (with default flush_interval=1s)
- **THEN** the BatchWriter SHALL flush the 10 events to ClickHouse

#### Scenario: Flush failure triggers Nack

- **WHEN** `Send` fails due to ClickHouse unavailability
- **THEN** the BatchWriter SHALL Nack the corresponding Kafka events to trigger re-delivery

### Requirement: ClickHouse connection uses Optional-init pattern

ClickHouse connection initialization SHALL follow the `NewXxxOptional` pattern: return nil when `conf.Data.ClickHouse` is nil or has empty `addrs`, log info, and not panic.

#### Scenario: ClickHouse not configured

- **WHEN** `conf.Data.ClickHouse` is nil
- **THEN** the factory function SHALL return nil and log "ClickHouse not configured"

#### Scenario: ClickHouse connection failure

- **WHEN** ClickHouse connection fails
- **THEN** the factory function SHALL return nil and log a warning, not panic
