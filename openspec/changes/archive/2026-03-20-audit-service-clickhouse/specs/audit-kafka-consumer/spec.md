## ADDED Requirements

### Requirement: Audit consumer subscribes via pkg/broker interface

The audit consumer SHALL use `pkg/broker.Broker.Subscribe` to consume the audit topic (configured via `conf.App.Audit.topic`, default `"servora.audit.events"`). It SHALL NOT directly depend on franz-go or any broker-specific package.

#### Scenario: Consumer subscribes to configured topic

- **WHEN** the audit service starts with `audit.topic = "servora.audit.events"`
- **THEN** the consumer SHALL call `broker.Subscribe(ctx, "servora.audit.events", handler, opts...)` with the configured consumer group

#### Scenario: Consumer uses broker abstraction only

- **WHEN** the audit consumer import list is inspected
- **THEN** it SHALL import `pkg/broker` but SHALL NOT import `pkg/broker/kafka` or `github.com/twmb/franz-go`

### Requirement: Consumer handler deserializes and validates AuditEvent

The consumer handler SHALL:
1. Unmarshal `event.Message().Body` using `proto.Unmarshal` into `auditv1.AuditEvent`
2. Validate required fields: `event_id`, `event_type`, `occurred_at`, `service`
3. On successful validation, submit the event to the BatchWriter buffer
4. On deserialization or validation failure, log the error and Ack the event (skip bad messages)

#### Scenario: Valid proto message is accepted

- **WHEN** a valid proto-encoded `AuditEvent` is consumed from Kafka
- **THEN** the handler SHALL deserialize it, validate fields, and submit to the BatchWriter

#### Scenario: Invalid proto message is skipped

- **WHEN** a message with corrupted proto bytes is consumed
- **THEN** the handler SHALL log an error at Warn level and Ack the message to skip it

#### Scenario: Missing required fields are rejected

- **WHEN** a proto message with empty `event_id` is consumed
- **THEN** the handler SHALL log a validation error and Ack the message to skip it

### Requirement: Consumer lifecycle integrates with Kratos app

The consumer SHALL start when the Kratos application starts and stop gracefully on shutdown:
- Subscribe during server initialization
- Unsubscribe and flush remaining buffer on `Close()`

#### Scenario: Graceful shutdown flushes buffer

- **WHEN** the audit service receives a shutdown signal
- **THEN** the consumer SHALL unsubscribe, flush any remaining buffered events to ClickHouse, and then close the ClickHouse connection

#### Scenario: Consumer start with nil broker

- **WHEN** the Broker is nil (Kafka not configured)
- **THEN** the consumer SHALL log a warning and operate as a no-op (no subscription, no error)
