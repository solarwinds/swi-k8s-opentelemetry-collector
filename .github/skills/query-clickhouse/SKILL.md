---
name: query-clickhouse
description: 'Query ClickHouse database holding SWO K8s Collector telemetry data (entity state events, metrics, logs, traces). Use when verifying entity types, relationship events, metrics, or debugging collector output.'
tags: [clickhouse, entity-state-events, troubleshooting, collector, telemetry]
---

# Query ClickHouse - SWO K8s Collector Telemetry

You are a ClickHouse SQL expert. You help query ClickHouse which runs locally holding all telemetry data that the SWO K8s Collector produces.

## When to Use This Skill

- Verifying entity state events are produced correctly by the collector
- Checking relationship events between Kubernetes entities
- Debugging collector output (metrics, logs, traces)
- Counting unique entities or relationship types
- Investigating metric names and volumes
- Validating new entity types or relationships after code changes

## Prerequisites

ClickHouse must be port-forwarded from a Kubernetes cluster to localhost:8123. Typically from the e2e test cluster:

```bash
kubectl --context e2e-cluster port-forward -n test-namespace svc/clickhouse 8123:8123
```

Verify connectivity:

```bash
curl -s http://localhost:8123/ping
```

## Connection Details

| Parameter | Value |
|-----------|-------|
| Host | localhost |
| Port | 8123 |
| User | default |
| Password | _(none)_ |
| Database | otel |
| Protocol | HTTP |

## Database Schema

### Tables

| Table | Description |
|-------|-------------|
| `otel.otel_logs` | Logs data, including entity state events and entity relationship state events |
| `otel.otel_traces` | Traces data |
| `otel.otel_metrics_gauge` | Metrics of gauge type |
| `otel.otel_metrics_sum` | Metrics of sum type |
| `otel.otel_metrics_summary` | Metrics of summary type |
| `otel.otel_metrics_histogram` | Metrics of histogram type |

Full schema details (column definitions, engines, TTLs) are in the ClickHouse exporter configuration:
`tests/deploy/timeseries-mock-service/templates/configmap.yaml`

### Key Columns in `otel_logs`

- `Timestamp` — event timestamp
- `LogAttributes` — map of log-level attributes (entity type, id, event type, etc.)
- `ScopeAttributes` — map of scope-level attributes (e.g., `otel.entity.event_as_log`)
- `ResourceAttributes` — map of resource-level attributes

### Key Columns in `otel_metrics_*`

- `MetricName` — name of the metric
- `Timestamp` — metric timestamp
- `ResourceAttributes` — resource attributes
- `Attributes` — metric data point attributes

## Entity State Event Specification

### Entity State Event

Entity events use LogRecords to carry entity data. Multiple entities can be sent in a single OTEL message.

| Attribute | Value | Description |
|-----------|-------|-------------|
| `scopeLogs.attributes` | `otel.entity.event_as_log=true` | **Required** scope marking content as entity state carrier |
| `log_records.timeUnixNano` | `<epoch in nanos>` | **Optional** capture timestamp; defaults to collector receive time |
| `log_records.attributes[otel.entity.event.type]` | `entity_state` | **Required** entity event type |
| `log_records.attributes[otel.entity.type]` | `<entity type>` | **Required** entity type (case insensitive) |
| `log_records.attributes[otel.entity.id]` | KV list: `id=e-1234` or `service.name=entity-service` | **Required** entity identifier — explicit ID or key attributes from entity schema |
| `log_records.attributes[otel.entity.attributes]` | KV list: `telemetry.sdk.language_display_name=kotlin` | **Optional** entity attributes, same as telemetry resource attributes |

### Entity Relationship State Event

| Attribute | Value | Description |
|-----------|-------|-------------|
| `scopeLogs.attributes` | `otel.entity.event_as_log=true` | **Required** scope marking content as entity state carrier |
| `log_records.timeUnixNano` | `<epoch in nanos>` | **Optional** capture timestamp |
| `log_records.attributes[otel.entity.event.type]` | `entity_relationship_state` | **Required** relationship event type |
| `log_records.attributes[otel.entity_relationship.type]` | `<relationship type>` | **Required** relationship type (case insensitive) |
| `log_records.attributes[otel.entity_relationship.source_entity.type]` | `<entity type>` | **Required** source entity type (case insensitive) |
| `log_records.attributes[otel.entity_relationship.source_entity.id]` | KV list | **Required** source entity identifier |
| `log_records.attributes[otel.entity_relationship.destination_entity.type]` | `<entity type>` | **Required** destination entity type (case insensitive) |
| `log_records.attributes[otel.entity_relationship.destination_entity.id]` | KV list | **Required** destination entity identifier |
| `log_records.attributes[otel.entity_relationship.attributes]` | KV list | **Optional** relationship attributes |

## Query Method

Use POST body queries to the ClickHouse HTTP interface:

```bash
echo "<SQL>" | curl -s 'http://localhost:8123/' --data-binary @-
```

Use `FORMAT PrettyCompact` for readable tables, `FORMAT JSONEachRow` piped to `jq` for JSON output.

## Common Query Patterns

### Entity State Events

#### Count all entity types and their volumes

```bash
echo "SELECT LogAttributes['otel.entity.type'] as entity_type, count(*) as cnt FROM otel.otel_logs WHERE LogAttributes['otel.entity.event.type'] = 'entity_state' GROUP BY entity_type ORDER BY cnt DESC FORMAT PrettyCompact" | curl -s 'http://localhost:8123/' --data-binary @-
```

#### Query specific entity type (e.g., KubernetesContainerImage)

```bash
echo "SELECT LogAttributes['otel.entity.id'] as entity_id, LogAttributes['otel.entity.attributes'] as attrs FROM otel.otel_logs WHERE LogAttributes['otel.entity.type'] = 'KubernetesContainerImage' LIMIT 10 FORMAT PrettyCompact" | curl -s 'http://localhost:8123/' --data-binary @-
```

#### Count unique entities of a type

```bash
echo "SELECT count(DISTINCT LogAttributes['otel.entity.id']) as unique_count FROM otel.otel_logs WHERE LogAttributes['otel.entity.type'] = 'KubernetesContainerImage'" | curl -s 'http://localhost:8123/' --data-binary @-
```

#### Query entity data as JSON

```bash
echo "SELECT Timestamp, LogAttributes FROM otel.otel_logs WHERE ScopeAttributes['otel.entity.event_as_log'] = 'true' AND LogAttributes['otel.entity.event.type'] = 'entity_state' ORDER BY Timestamp DESC LIMIT 2 FORMAT JSONEachRow" | curl -s 'http://localhost:8123/' --data-binary @- | jq .
```

#### Query entity data as table

```bash
echo "SELECT Timestamp, LogAttributes['otel.entity.type'] as entity_type, LogAttributes['otel.entity.id'] as entity_id FROM otel.otel_logs WHERE ScopeAttributes['otel.entity.event_as_log'] = 'true' LIMIT 10 FORMAT PrettyCompact" | curl -s 'http://localhost:8123/' --data-binary @-
```

### Relationship Events

#### Count all relationship types and their volumes

```bash
echo "SELECT LogAttributes['otel.entity_relationship.type'] as rel_type, count(*) as cnt FROM otel.otel_logs WHERE LogAttributes['otel.entity.event.type'] = 'entity_relationship_state' GROUP BY rel_type ORDER BY cnt DESC FORMAT PrettyCompact" | curl -s 'http://localhost:8123/' --data-binary @-
```

#### Query relationships with source and destination types

```bash
echo "SELECT LogAttributes['otel.entity_relationship.type'] as rel_type, LogAttributes['otel.entity_relationship.source.type'] as src_type, LogAttributes['otel.entity_relationship.destination.type'] as dst_type, count(*) as cnt FROM otel.otel_logs WHERE LogAttributes['otel.entity.event.type'] = 'entity_relationship_state' GROUP BY rel_type, src_type, dst_type ORDER BY cnt DESC FORMAT PrettyCompact" | curl -s 'http://localhost:8123/' --data-binary @-
```

### Metrics

#### Overview of gauge metrics

```bash
echo "SELECT MetricName, count(*) as cnt FROM otel.otel_metrics_gauge GROUP BY MetricName ORDER BY cnt DESC LIMIT 20 FORMAT PrettyCompact" | curl -s 'http://localhost:8123/' --data-binary @-
```

#### Overview of sum metrics

```bash
echo "SELECT MetricName, count(*) as cnt FROM otel.otel_metrics_sum GROUP BY MetricName ORDER BY cnt DESC LIMIT 20 FORMAT PrettyCompact" | curl -s 'http://localhost:8123/' --data-binary @-
```

#### Overview of histogram metrics

```bash
echo "SELECT MetricName, count(*) as cnt FROM otel.otel_metrics_histogram GROUP BY MetricName ORDER BY cnt DESC LIMIT 20 FORMAT PrettyCompact" | curl -s 'http://localhost:8123/' --data-binary @-
```

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| Empty results for entity queries | Collector not producing entity state events | Check collector pod logs for errors; verify entity generation config |
| `curl: (7) Failed to connect` | Port-forward not active | Re-run `kubectl port-forward` command |
| No metrics in `otel_metrics_*` | Collector not scraping or exporting metrics | Check collector config and target pod annotations |
| `DB::Exception: Table otel.otel_logs doesn't exist` | ClickHouse schema not initialized | Check the timeseries-mock-service deployment and configmap |
