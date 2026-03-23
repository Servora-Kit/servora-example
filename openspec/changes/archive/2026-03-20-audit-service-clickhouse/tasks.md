## 1. Proto 扩展与生成

- [x] 1.1 扩展 `conf.proto` 的 `App.Audit` message，新增 `consumer_batch_size` (field 5)、`consumer_flush_interval` (field 6)、`retention_days` (field 7)
- [x] 1.2 执行 `make api` 重新生成 Go/TS 产物，验证编译通过

## 2. 服务脚手架

- [x] 2.1 创建 `app/audit/service/` 目录结构：`cmd/server/`、`internal/{biz,data,server,service}/`、`configs/`
- [x] 2.2 创建 `app/audit/service/go.mod`，声明 module + replace directives（参考 IAM）
- [x] 2.3 创建 `app/audit/service/Makefile`（`include ../../../app.mk`）
- [x] 2.4 将 `./app/audit/service` 加入 `go.work`
- [x] 2.5 创建 `app/audit/service/configs/config.yaml`，配置 Kafka broker、ClickHouse 连接、审计参数
- [x] 2.6 创建 `cmd/server/main.go` — Kratos bootstrap（参考 IAM `cmd/server/main.go`）
- [x] 2.7 创建 `cmd/server/wire.go` — Wire provider set 骨架

## 3. 查询 API Proto

- [x] 3.1 创建 `app/audit/service/api/protos/buf.yaml`，声明 proto module
- [x] 3.2 将 `app/audit/service/api/protos` 加入根 `buf.yaml` modules 列表
- [x] 3.3 创建 `app/audit/service/api/protos/servora/audit/service/v1/audit_service.proto`：定义 `AuditQueryService`（`ListAuditEvents` + `CountAuditEvents`）、request/response messages、HTTP transcoding annotations
- [x] 3.4 执行 `make api` 生成 Go 代码，验证编译通过

## 4. ClickHouse 数据层

- [x] 4.1 实现 `internal/data/clickhouse.go` — ClickHouse 连接初始化（`NewClickHouseOptional` 模式）、DDL 自动创建（`CREATE TABLE IF NOT EXISTS audit_events`）
- [x] 4.2 实现 `internal/data/batch_writer.go` — BatchWriter：内存 buffer + size/time 双阈值 flush + `PrepareBatch`/`Append`/`Send` + 失败 Nack
- [x] 4.3 实现 `internal/data/audit_repo.go` — AuditRepo：`ListEvents`（过滤 + cursor 分页）、`CountEvents`（聚合计数）

## 5. Kafka Consumer

- [x] 5.1 实现 `internal/biz/consumer.go` — Consumer：`pkg/broker.Subscribe` 消费审计 topic → proto 反序列化 → 校验 → 提交 BatchWriter；生命周期 Start/Stop
- [x] 5.2 处理 nil Broker 场景（Kafka 未配置时 no-op）

## 6. 查询 Service 层

- [x] 6.1 实现 `internal/service/audit.go` — Kratos service，调用 AuditRepo 实现 `ListAuditEvents` + `CountAuditEvents`
- [x] 6.2 实现 `internal/server/grpc.go` + `internal/server/http.go` — 注册 AuditQueryService

## 7. Wire 注入

- [x] 7.1 创建 `internal/data/data.go` — data ProviderSet（ClickHouse conn、BatchWriter、AuditRepo）
- [x] 7.2 创建 `internal/biz/biz.go` — biz ProviderSet（Consumer）
- [x] 7.3 创建 `internal/service/service.go` — service ProviderSet（AuditService）
- [x] 7.4 创建 `internal/server/server.go` — server ProviderSet（gRPC、HTTP）
- [x] 7.5 完善 `cmd/server/wire.go`，组合所有 ProviderSet
- [x] 7.6 执行 `make wire`（audit service 目录下），验证 `wire_gen.go` 生成

## 8. 基础设施集成

- [x] 8.1 `docker-compose.dev.yaml` 新增 audit service 容器定义（build、depends_on、ports、volumes、network）
- [x] 8.2 验证 `make compose.dev` 能正常启动 audit service（kafka + clickhouse 依赖就绪后启动）

## 9. 端到端验证

- [x] 9.1 启动完整环境（kafka + clickhouse + audit service），通过已有的 `pkg/audit` BrokerEmitter 发送测试事件（sayhello 作为发布者）
- [x] 9.2 验证事件从 Kafka 被 audit service 消费并写入 ClickHouse（`flushed 1 events to ClickHouse` 日志确认）
- [x] 9.3 通过查询 API（HTTP）验证 `ListAuditEvents` 和 `CountAuditEvents` 返回正确结果（含 bug fix：nil timestamp 导致 `occurred_at < 1970-01-01` 过滤全量数据）
- [x] 9.4 验证分页、过滤条件、默认值等查询行为（nil StartTime/EndTime 默认无过滤，nextPageToken 正常返回）

## 10. 文档与 OpenSpec 同步

- [x] 10.1 更新 `docs/plans/2026-03-20-keycloak-openfga-audit-design.md`，标记 Phase 2b 为 ✅ 已完成
- [x] 10.2 同步 OpenSpec specs 到 `openspec/specs/`（新增 4 个 + 更新 2 个）
