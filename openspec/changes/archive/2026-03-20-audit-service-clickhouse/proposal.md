## Why

Phase 2a 完成了审计事件的生产端（`pkg/authz` 和 `pkg/openfga` 自动 emit `authz.decision` / `tuple.changed`），但事件消费端尚未建设。当前审计事件仅可通过 `LogEmitter` 输出到日志或通过 `BrokerEmitter` 发送到 Kafka topic，但没有服务从 Kafka 消费、持久化、查询这些事件。Phase 2b 补齐这条链路的最后一环——从 Kafka 到 ClickHouse 再到查询 API。

参考主设计文档 **Phase 2b: Audit Service + ClickHouse**（`docs/plans/2026-03-20-keycloak-openfga-audit-design.md` Section 10）。

## What Changes

- **新建 `app/audit/service` 微服务**：完整的 Kratos 微服务（cmd/internal/api/configs），参考 `app/iam/service` 分层结构
- **Kafka consumer**：复用 `pkg/broker.Subscribe` 消费审计 topic，proto 反序列化 + 校验 + 投递到 batch buffer
- **ClickHouse writer**：batch flush 策略（`PrepareBatch` + `Append` + `Send`），可配置 batch_size / flush_interval
- **ClickHouse schema**：`audit_events` 表，`detail` 列使用纯 JSON string 存储，按天分区，可配置 TTL
- **DDL 管理**：Go 代码内嵌 `CREATE TABLE IF NOT EXISTS`，服务启动时自动执行
- **查询 API proto**：`app/audit/service/api/protos/servora/audit/service/v1/` 定义 `ListAuditEvents` + `CountAuditEvents`（gRPC + HTTP 转码）
- **`conf.proto` 扩展**：`App.Audit` 新增 `consumer_batch_size`、`consumer_flush_interval`、`retention_days` 字段
- **docker-compose.dev 集成**：audit service 加入开发环境

## Non-goals

- 不涉及 Keycloak 集成（Phase 3）
- 不涉及 `protoc-gen-servora-audit` 代码生成器（Phase 4）
- 不涉及 `resource.mutation` 类型的自动 emit（Phase 4）
- 不涉及 `authn.result` 类型事件的接入（Phase 3）
- 不新建 `pkg/clickhouse` 框架包——ClickHouse 客户端仅在 audit service 内部使用，待有第二个服务需要时再框架化
- 不做 ClickHouse 列打平或 materialized column 优化（后续按需添加）
- 不做审计事件的聚合统计或仪表盘

## Capabilities

### New Capabilities

- `audit-service-scaffold`: audit 微服务脚手架（目录结构、cmd、go.mod、Makefile、Wire 注入、configs）
- `audit-clickhouse-storage`: ClickHouse schema 设计、DDL 自动创建、batch writer 实现
- `audit-kafka-consumer`: Kafka consumer 实现（复用 pkg/broker.Subscribe、proto 反序列化、校验、batch 投递）
- `audit-query-api`: 审计查询 API proto 定义与 gRPC/HTTP 实现（ListAuditEvents + CountAuditEvents）

### Modified Capabilities

- `config-proto-extension`: `App.Audit` 新增 consumer 侧配置字段（consumer_batch_size / consumer_flush_interval / retention_days）
- `infra-kafka-clickhouse`: docker-compose.dev 新增 audit service 容器

## Impact

- **新增目录**: `app/audit/service/`（含 cmd/internal/api/configs/Makefile/go.mod）
- **新增 Go module**: `app/audit/service` 加入 `go.work`
- **新增 proto module**: `app/audit/service/api/protos/` 加入 `buf.yaml` workspace
- **修改 proto**: `api/protos/servora/conf/v1/conf.proto`（App.Audit 扩展）
- **修改 compose**: `docker-compose.dev.yaml` 新增 audit service
- **新依赖**: `github.com/ClickHouse/clickhouse-go/v2`（audit service go.mod 内）
- **生成产物**: `make api` 重新生成（conf.proto 变更）；audit service 内部 `make wire`
