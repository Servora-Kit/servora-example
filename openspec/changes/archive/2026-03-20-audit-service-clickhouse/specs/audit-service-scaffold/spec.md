## ADDED Requirements

### Requirement: Audit service directory structure follows Servora conventions

`app/audit/service/` SHALL follow the standard Servora microservice layout:
- `cmd/server/main.go` — Kratos application bootstrap
- `cmd/server/wire.go` + `wire_gen.go` — Wire DI
- `internal/biz/` — business logic (consumer orchestration, query use cases)
- `internal/data/` — data access layer (ClickHouse read/write)
- `internal/server/` — gRPC + HTTP server registration
- `internal/service/` — Kratos service implementation
- `configs/` — service configuration files
- `Makefile` — `include ../../../app.mk`

#### Scenario: Service compiles independently

- **WHEN** `go build ./app/audit/service/...` is run
- **THEN** compilation SHALL succeed without errors

#### Scenario: Wire injection generates successfully

- **WHEN** `make wire` is run in `app/audit/service/`
- **THEN** `wire_gen.go` SHALL be generated with all providers wired correctly

### Requirement: Audit service Go module integrates with workspace

`app/audit/service/go.mod` SHALL declare module `github.com/Servora-Kit/servora/app/audit/service` with `replace` directives for `../../..` (root) and `../../../api/gen` (generated code), matching the IAM service pattern. The module SHALL be listed in `go.work`.

#### Scenario: go.work includes audit service

- **WHEN** `go.work` is read
- **THEN** it SHALL contain `./app/audit/service` in the `use` block

#### Scenario: Module resolves all dependencies

- **WHEN** `go mod tidy` is run in `app/audit/service/`
- **THEN** it SHALL complete without errors

### Requirement: Audit service proto module integrates with Buf workspace

`app/audit/service/api/protos/` SHALL be declared as a module in the root `buf.yaml`. The proto module SHALL have its own `buf.yaml` declaring the module path.

#### Scenario: buf.yaml includes audit service protos

- **WHEN** the root `buf.yaml` is read
- **THEN** it SHALL contain `app/audit/service/api/protos` in the modules list

#### Scenario: buf lint passes for audit service protos

- **WHEN** `make lint.proto` is run
- **THEN** audit service protos SHALL pass linting without errors

### Requirement: Audit service Wire providers follow DI conventions

Wire providers SHALL inject:
- ClickHouse connection from `conf.Data.ClickHouse`
- `pkg/broker.Broker` from `pkg/broker/kafka.NewBrokerOptional`
- BatchWriter from ClickHouse conn + `conf.App.Audit` config
- Consumer from Broker + BatchWriter + audit config
- AuditRepo from ClickHouse conn (for queries)
- AuditService from AuditRepo

#### Scenario: Wire resolves all audit service dependencies

- **WHEN** Wire provider set is defined with ClickHouse, Broker, BatchWriter, Consumer, AuditRepo, AuditService
- **THEN** `wire_gen.go` SHALL resolve all dependencies without circular imports
