## Context

Phase 1 建设了审计骨架（`pkg/audit`、`pkg/broker`、audit proto），Phase 2a 完成了 emit 接入（`pkg/authz` 自动 emit `authz.decision`、`pkg/openfga` 自动 emit `tuple.changed`）。事件已经可以通过 `BrokerEmitter` 发布到 Kafka topic `servora.audit.events`，但缺少消费端。

Phase 2b 补齐最后一环：新建 `app/audit/service` 微服务，从 Kafka 消费审计事件，持久化到 ClickHouse，并暴露查询 API。

当前基础设施状态：
- Kafka（KRaft，docker-compose 已配置，含 EXTERNAL listener 29092）
- ClickHouse（docker-compose 已配置，`servora_audit` 库，端口 18123/19000）
- `pkg/broker` 已实现 `Broker.Subscribe` 消费能力（franz-go）
- `servora.audit.v1.AuditEvent` proto 已定义，含 4 种 typed detail

参考：
- `app/iam/service` — 服务分层、Wire DI、Makefile、go.mod 结构
- `/Users/horonlee/projects/go/Kemate` — Kafka consumer 与 ClickHouse 写入模式

## Goals / Non-Goals

**Goals:**
- 新建完整 Kratos 微服务 `app/audit/service`
- 实现 Kafka → proto 反序列化 → 校验 → batch buffer → ClickHouse 写入管线
- 提供 `ListAuditEvents` + `CountAuditEvents` 查询 API（gRPC + HTTP 转码）
- 所有运行参数可配置（batch_size、flush_interval、retention_days 等）
- 服务启动时自动创建 ClickHouse 表（idempotent DDL）

**Non-Goals:**
- 不构建 `pkg/clickhouse` 框架包（仅 audit service 内部使用）
- 不做 ClickHouse 列打平 / materialized column
- 不接入 `authn.result`（Phase 3）或 `resource.mutation` 自动 emit（Phase 4）
- 不做审计仪表盘 / 聚合统计

## Decisions

### D1: 服务分层 — 参考 `app/iam/service`

```
app/audit/service/
├── cmd/server/          # main.go + wire.go + wire_gen.go
├── internal/
│   ├── biz/             # 业务逻辑（consumer 编排、查询）
│   ├── data/            # ClickHouse 读写
│   ├── server/          # gRPC + HTTP server 注册
│   └── service/         # Kratos service 实现
├── api/protos/          # 私有查询 API proto
├── configs/             # 服务配置文件
├── Makefile             # include ../../../app.mk
└── go.mod
```

**为什么参考 IAM**：保持 Servora 服务一致的分层惯例，降低认知负担。

### D2: ClickHouse 客户端 — 官方 native driver

使用 `github.com/ClickHouse/clickhouse-go/v2` 的 native 协议（非 HTTP），配合 `PrepareBatch` + `Append` + `Send` 进行高效批量写入。

**替代方案**：HTTP JSONEachRow（Kemate 方式）——但 native 协议的压缩效率和批量写入性能更优，且官方 driver 直接支持 `PrepareBatch` API。

### D3: Kafka consumer — 复用 `pkg/broker.Subscribe`

直接使用 `pkg/broker.Subscribe` 消费审计 topic，不在 audit service 中直接依赖 franz-go。

```
pkg/broker.Subscribe("servora.audit.events", handler, WithQueue("audit-consumer"))
    ↓ handler
proto.Unmarshal(event.Message().Body) → validate → batch buffer
```

**为什么不直接用 franz-go**：保持 broker 抽象层的价值，audit service 作为框架消费端的参考实现。

### D4: Batch 写入策略

```
┌──────────┐     ┌───────────────────────┐     ┌─────────────┐
│  Kafka   │────▶│  BatchWriter          │────▶│ ClickHouse  │
│  Event   │     │  ┌─────────────────┐  │     │             │
│          │     │  │ in-memory buffer│  │     │ audit_events│
│          │     │  │ (configurable)  │  │     │             │
│          │     │  └─────────────────┘  │     └─────────────┘
└──────────┘     │  flush triggers:      │
                 │  - size >= batch_size │
                 │  - age >= interval    │
                 └───────────────────────┘
```

- 内存 buffer 收集反序列化后的 `AuditEvent`
- 达到 `consumer_batch_size`（默认 100）或 `consumer_flush_interval`（默认 1s）先到先 flush
- flush 使用 `PrepareBatch` + 逐条 `Append` + `Send`
- flush 失败时，将 batch 中的事件 Nack 让 Kafka re-deliver
- flush 成功后，Ack 所有已消费的事件

**风险**：flush 失败导致重复消费 → ClickHouse 端依靠 `event_id` 去重（`ORDER BY` 含 `event_id`），幂等可接受。

### D5: ClickHouse schema

```sql
CREATE TABLE IF NOT EXISTS audit_events (
    event_id     String,
    event_type   LowCardinality(String),
    event_version String,
    occurred_at  DateTime64(3, 'UTC'),

    service      LowCardinality(String),
    operation    String,

    actor_id     String,
    actor_type   LowCardinality(String),
    actor_display_name String,

    target_type  LowCardinality(String),
    target_id    String,
    target_name  String,

    success      Bool,
    error_code   String,
    error_message String,

    trace_id     String,
    request_id   String,

    detail       String  -- JSON string, proto oneof detail 序列化

) ENGINE = MergeTree()
PARTITION BY toDate(occurred_at)
ORDER BY (service, event_type, occurred_at, event_id)
TTL occurred_at + INTERVAL 90 DAY
SETTINGS index_granularity = 8192
```

**关键决策：**
- `detail` 列使用纯 JSON string 存储 proto `oneof detail` 序列化内容。后续如有高频查询需求，可事后加 `MATERIALIZED` 列提取特定字段
- `PARTITION BY toDate(occurred_at)` 按天分区
- TTL 通过 `retention_days` 配置控制，DDL 中硬编码默认值 90 天
- `LowCardinality` 用于低基数枚举列以压缩存储

**DDL 管理**：Go 代码内嵌 `CREATE TABLE IF NOT EXISTS`，服务启动时执行。TTL 的 INTERVAL 值从配置读取并格式化到 DDL 中。

### D6: 查询 API proto — 服务私有

```
app/audit/service/api/protos/servora/audit/service/v1/
├── audit_service.proto    # ListAuditEvents + CountAuditEvents
└── buf.yaml               # proto module 声明
```

与 IAM 模式一致，私有 API proto 放在服务目录下。注意区分：
- `api/protos/servora/audit/v1/` — 共享 audit event/annotation proto
- `app/audit/service/api/protos/servora/audit/service/v1/` — audit 微服务的查询 API

查询 API 支持：
- 时间范围筛选（`start_time` / `end_time`）
- 事件类型筛选（`event_types`）
- Actor 筛选（`actor_id`）
- 服务筛选（`service`）
- 分页（`page_size` + `page_token`，基于 `occurred_at + event_id` cursor）

### D7: `conf.proto` 扩展

`App.Audit` 新增 consumer 侧字段：

```protobuf
message Audit {
  bool enabled = 1;
  string emitter_type = 2;
  string topic = 3;
  string service_name = 4;
  // Phase 2b consumer 侧
  int32 consumer_batch_size = 5;       // 默认 100
  google.protobuf.Duration consumer_flush_interval = 6;  // 默认 1s
  int32 retention_days = 7;            // ClickHouse TTL 天数，默认 90
}
```

这些字段由 audit service 消费，emit 端无需关注。

### D8: Wire 依赖注入

```
┌──────────────┐
│   main.go    │
└──────┬───────┘
       │ Wire
       ▼
┌──────────────────────────────────────────────────────┐
│  ClickHouse conn ← conf.Data.ClickHouse              │
│  Broker ← pkg/broker/kafka.NewBrokerOptional          │
│  BatchWriter ← ClickHouse conn + conf.App.Audit       │
│  AuditRepo ← ClickHouse conn (查询)                   │
│  Consumer ← Broker + BatchWriter + conf.App.Audit     │
│  AuditService ← AuditRepo                            │
│  gRPC/HTTP Server ← AuditService                     │
└──────────────────────────────────────────────────────┘
```

### D9: docker-compose.dev 集成

audit service 加入 `docker-compose.dev.yaml`，依赖 kafka + clickhouse，暴露 gRPC + HTTP 端口。

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Kafka 消费失败导致事件丢失 | franz-go 自动 commit offset 仅在 Ack 后，flush 失败 Nack 触发 re-delivery |
| ClickHouse 重复写入（at-least-once） | `event_id` 唯一，重复写入不影响查询语义（MergeTree 保留所有行，查询时可 FINAL 去重） |
| batch buffer 内存压力 | batch_size 默认 100，单条事件约 1-2KB，100 条约 100-200KB，可接受 |
| ClickHouse DDL 变更（schema migration） | 当前仅 Phase 2b 建表，未来 schema 变更用 ALTER TABLE 或版本化 DDL；暂不引入 migration 框架 |
| `detail` JSON 查询性能 | 初期不做索引，如需按 detail 内字段过滤，后续添加 materialized column |

## Migration Plan

无 breaking change。纯新增服务，不影响已有 pkg 或服务。

步骤：
1. 扩展 `conf.proto` → `make api`
2. 新建 `app/audit/service` 目录 → 加入 `go.work` + `buf.yaml`
3. 实现 consumer + writer + query → `make wire`（audit service 内）
4. 定义查询 API proto → `make api`
5. docker-compose.dev 集成 → `make compose.dev` 验证
6. 端到端测试：业务操作 → Kafka → audit service → ClickHouse → 查询 API

## Open Questions

无（所有设计决策在 Phase 2b explore 阶段已确认）。
