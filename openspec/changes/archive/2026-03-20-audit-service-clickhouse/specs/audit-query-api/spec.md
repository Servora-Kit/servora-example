## ADDED Requirements

### Requirement: ListAuditEvents API supports filtered pagination

The audit service SHALL expose a `ListAuditEvents` RPC (gRPC + HTTP GET transcoding) that accepts:
- `google.protobuf.Timestamp start_time` — filter events after this time (inclusive)
- `google.protobuf.Timestamp end_time` — filter events before this time (exclusive)
- `repeated string event_types` — filter by event type names
- `string actor_id` — filter by actor ID
- `string service` — filter by service name
- `int32 page_size` — max results per page (default 50, max 200)
- `string page_token` — cursor for pagination

The response SHALL include:
- `repeated AuditEventItem events` — list of audit events
- `string next_page_token` — cursor for next page (empty if last page)

`AuditEventItem` SHALL mirror the ClickHouse row structure with `detail` as a JSON string field.

#### Scenario: List with time range filter

- **WHEN** `ListAuditEvents` is called with `start_time = 2026-03-01T00:00:00Z` and `end_time = 2026-03-02T00:00:00Z`
- **THEN** the response SHALL only contain events with `occurred_at` within the specified range

#### Scenario: List with event type filter

- **WHEN** `ListAuditEvents` is called with `event_types = ["AUTHZ_DECISION"]`
- **THEN** the response SHALL only contain events of type `AUTHZ_DECISION`

#### Scenario: Pagination with page token

- **WHEN** `ListAuditEvents` returns `next_page_token = "xxx"`
- **AND** a subsequent call is made with `page_token = "xxx"`
- **THEN** the response SHALL contain the next page of results without duplicates

#### Scenario: Default page size

- **WHEN** `ListAuditEvents` is called without `page_size`
- **THEN** the response SHALL return at most 50 events

### Requirement: CountAuditEvents API returns filtered counts

The audit service SHALL expose a `CountAuditEvents` RPC (gRPC + HTTP GET transcoding) that accepts the same filter parameters as `ListAuditEvents` (except pagination) and returns:
- `int64 total_count` — total number of matching events

#### Scenario: Count all events

- **WHEN** `CountAuditEvents` is called with no filters
- **THEN** the response SHALL contain the total count of all audit events

#### Scenario: Count with combined filters

- **WHEN** `CountAuditEvents` is called with `service = "iam"` and `event_types = ["TUPLE_CHANGED"]`
- **THEN** the response SHALL contain the count of matching events only

### Requirement: Query API proto is defined as service-private

The query API proto SHALL be located at `app/audit/service/api/protos/servora/audit/service/v1/audit_service.proto` with package `servora.audit.service.v1`. This is separate from the shared `servora.audit.v1` package.

The proto SHALL import `google/api/annotations.proto` for HTTP transcoding and use standard Kratos HTTP mapping conventions.

#### Scenario: Proto compiles and generates Go code

- **WHEN** `make api` is run
- **THEN** Go code for `servora.audit.service.v1.AuditQueryService` SHALL be generated without errors

#### Scenario: HTTP transcoding is configured

- **WHEN** the proto is inspected
- **THEN** `ListAuditEvents` SHALL have `option (google.api.http) = { get: "/v1/audit/events" }` and `CountAuditEvents` SHALL have `option (google.api.http) = { get: "/v1/audit/events:count" }`
