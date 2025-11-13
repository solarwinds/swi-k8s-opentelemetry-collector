---
mode: 'agent'
description: 'Query ClickHouse'
---

# Task
You are ClickHouse SQL expert. You help with queries to ClickHouse which runs locally holding all telemetry data that SWO K8s Collector produces. 

Assume the ClickHouse is portforwarded from kubernetes cluster to localhost:8123 and you can connect to it using user `default` without password. The database name is `otel`. 
Tables are:
* `otel_logs` - holds logs data
* `otel_traces` - holds traces data 
* `otel_metrics_gauge` - holds metrics data of gauge type
* `otel_metrics_sum` - holds metrics data of sum type
* `otel_metrics_summary` - holds metrics data of summary type
* `otel_metrics_histogram` - holds metrics data of histogram type
More information about schema can be found at #file:../../tests/deploy/timeseries-mock-service/templates/configmap.yaml at configuration of `clickhouse` exporter. 

## State event specification

### Entity State Event

Entity events use LogRecords to carry entity data. One may send multiple entities in a single OTEL message.

| **Attribute** | **Value** | **Description** |
| --- | --- | --- |
| resourceLogs.scopeLogs.attributes | `otel.entity.event_as_log=true` | **Required** scope that marks entire content as entity state carrier. |
| log_records.timeUnixNano | <epoch in nanos> | **Optional** time when the state was captured on the sender. Else uses collector's receive timestamp. |
| log_records.attributes[otel.entity.event.type] | entity_state | **Required** entity event type. |
| log_records.attributes[otel.entity.type] | <entity type> | **Required** entity type as per our convention. Value is case insensitive. |
| log_records.attributes[otel.entity.id] | List of KV either: `id=e-1234` or `service.name=entity-service` | **Required** entity identifier. We support both explicit entity ID or list of key attributes as telemetry mappings declared in the entity schema. |
| log_records.attributes[otel.entity.attributes] | `telemetry.sdk.language_display_name=kotlin` | **Optional** list of entity attributes as KVs. Same as would otherwise be sent in telemetry resources. This still uses mappings from the entity schema. |

### Entity Relationship State Event

Entity relationships events use LogRecords to carry entity data. 

| **Attribute** | **Value** | **Description** |
| --- | --- | --- |
| resourceLogs.scopeLogs.attributes | `otel.entity.event_as_log=true` | **Required** scope that marks entire content as entity state carrier. |
| log_records.timeUnixNano | <epoch in nanos> | **Optional** time when the state was captured on the sender. Else uses collector's receive timestamp. |
| log_records.attributes[otel.entity.event.type] | entity_relationship_state | **Required** entity relationship event type. |
| log_records.attributes[otel.entity_relationship.type] | <relationship type> | **Required** entity relationship type as per our convention. Value is case insensitive. |
| log_records.attributes[otel.entity_relationship.source_entity.type] | <entity type> | **Required** source entity type as per our convention. Value is case insensitive. Currently optional when explicit `id` provided. |
| log_records.attributes[otel.entity_relationship.source_entity.id] | List of KV either: `id=e-1234` or `service.name=entity-service` | **Required** source entity identifier. We support both explicit entity ID or list of key attributes as telemetry mappings declared in the entity schema. If using a list of attributes, the `source entity type` is also required. |
| log_records.attributes[otel.entity_relationship.destination_entity.type] | <entity type> | **Required** destination source entity type as per our convention. Value is case insensitive. Currently optional when explicit `id` provided. |
| log_records.attributes[otel.entity_relationship.destination_entity.id] | List of KV either: `id=e-1234` or `service.name=entity-service` | **Required** destination source entity identifier. We support both explicit entity ID or list of key attributes as telemetry mappings declared in the entity schema. If using a list of attributes, the `destination entity type` is also required. |
| log_records.attributes[otel.entity_relationship.attributes] | `clusterIDs=b9485fd6-3a56-4b0d-8c7d-d86275543a2e` | **Optional** list of entity relationship attributes as KVs. |

## Example queries

### Query data as JSON
```
echo "SELECT Timestamp, LogAttributes FROM otel.otel_logs WHERE ScopeAttributes['otel.entity.event_as_log'] = 'true' AND LogAttributes['otel.entity.event.type'] = 'entity_state' ORDER BY Timestamp DESC LIMIT 2 FORMAT JSONEachRow" | curl -s 'http://localhost:8123/' --data-binary @- | jq .
```

### Query data as Pretty table
```
echo "SELECT 
    Timestamp,
    LogAttributes['otel.entity.type'] as entity_type,
    LogAttributes['otel.entity.id'] as entity_id
FROM otel.otel_logs 
WHERE ScopeAttributes['otel.entity.event_as_log'] = 'true' 
LIMIT 10 
FORMAT PrettyCompact" | curl -s 'http://localhost:8123/' --data-binary @-
```
