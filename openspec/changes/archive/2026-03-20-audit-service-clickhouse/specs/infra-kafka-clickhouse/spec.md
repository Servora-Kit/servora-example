## MODIFIED Requirements

### Requirement: ClickHouse service in docker-compose

`docker-compose.yaml` SHALL include a ClickHouse service using `clickhouse/clickhouse-server:latest`, with:
- Container name: `servora_clickhouse`
- HTTP interface on port 8123 (mapped to host 18123)
- Native interface on port 9000 (mapped to host 19000)
- Health check using clickhouse-client
- Named volume for data persistence
- Connected to `servora-network`
- Database `servora_audit` pre-created via environment variable

No table schemas SHALL be created by Docker init scripts. Table creation is managed by the audit service at startup.

#### Scenario: ClickHouse starts and becomes healthy

- **WHEN** `docker compose up clickhouse` is run
- **THEN** the ClickHouse container SHALL start, pass health check, and accept queries

#### Scenario: servora_audit database exists

- **WHEN** the ClickHouse container is healthy
- **THEN** the `servora_audit` database SHALL exist and be accessible

## ADDED Requirements

### Requirement: Audit service in docker-compose.dev

`docker-compose.dev.yaml` SHALL include an audit service container that:
- Builds from `app/audit/service/`
- Depends on `kafka` (healthy) and `clickhouse` (healthy)
- Mounts `app/audit/service/configs/` for configuration
- Exposes gRPC and HTTP ports
- Connects to `servora-network`

#### Scenario: make compose.dev starts audit service

- **WHEN** `make compose.dev` is run
- **THEN** the audit service container SHALL start after kafka and clickhouse are healthy

#### Scenario: Audit service connects to Kafka and ClickHouse

- **WHEN** the audit service container is running
- **THEN** it SHALL successfully connect to the Kafka broker and ClickHouse server within the Docker network
