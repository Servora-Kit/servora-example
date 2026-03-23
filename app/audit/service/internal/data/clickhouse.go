package data

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	conf "github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
	pkgch "github.com/Servora-Kit/servora/pkg/db/clickhouse"
	"github.com/Servora-Kit/servora/pkg/logger"
)

// NewClickHouseClient opens a ClickHouse connection via pkg/db/clickhouse.
// Returns (nil, nil) when ClickHouse is not configured or unreachable — callers
// must nil-check the returned conn. The connection lifecycle (Close) is owned by
// NewData, following the same pattern as IAM's NewDBClient + NewData.
func NewClickHouseClient(cfg *conf.Data, l logger.Logger) (driver.Conn, error) {
	conn := pkgch.NewConnOptional(cfg, l)
	return conn, nil
}

// createAuditEventsTable executes the DDL to create the audit_events table idempotently.
func createAuditEventsTable(ctx context.Context, conn driver.Conn, retentionDays int32) error {
	if retentionDays <= 0 {
		retentionDays = 90
	}
	ddl := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS audit_events (
    event_id              String,
    event_type            LowCardinality(String),
    event_version         String,
    occurred_at           DateTime64(3, 'UTC'),

    service               LowCardinality(String),
    operation             String,

    actor_id              String,
    actor_type            LowCardinality(String),
    actor_display_name    String,

    target_type           LowCardinality(String),
    target_id             String,
    target_name           String,

    success               Bool,
    error_code            String,
    error_message         String,

    trace_id              String,
    request_id            String,

    detail                String
) ENGINE = MergeTree()
PARTITION BY toDate(occurred_at)
ORDER BY (service, event_type, occurred_at, event_id)
TTL occurred_at + INTERVAL %d DAY
SETTINGS index_granularity = 8192
`, retentionDays)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return conn.Exec(ctx, ddl)
}
