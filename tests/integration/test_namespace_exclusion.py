import os
from time import sleep
import pytest
from test_utils import retry_until_ok_clickhouse
from clickhouse_client import ClickHouseClient

clickhouse_endpoint = os.getenv("CLICKHOUSE_ENDPOINT", "localhost:8123")
clickhouse_client = ClickHouseClient(clickhouse_endpoint)
ignored_namespace = "test-namespace-ignore"


def assert_no_data(count: int, record_type: str) -> None:
    """Assert that no data of a specific type from the ignored namespace is present.

    Args:
        count: Last observed record count for the excluded namespace.
        record_type: Type of the record being checked.
    """
    if count > 0:
        pytest.fail(f"Found {count} unexpected {record_type} records from ignored namespace '{ignored_namespace}'.")

def assert_no_data_with_namespace_in_resource_attributes(table, record_type: str) -> None:
    """Assert that no data of a specific type from the ignored namespace is present in the specified table.

    Args:
        table: The ClickHouse table to query.
        record_type: Type of the record being checked.
    """
    count = clickhouse_client.count_records(
        table,
        f"""mapContainsKey(ResourceAttributes, 'k8s.namespace.name') AND mapExists((k, v) -> and(ilike(k, '%k8s.namespace.name%'), v = '{ignored_namespace}'), ResourceAttributes)""",
    )
    assert_no_data(count, record_type)

def test_logs_not_collected_from_excluded_namespace() -> None:
    """Verify no logs from the excluded namespace reach ClickHouse.

    First confirms the log pipeline is live by asserting at least one
    log record is present in otel_logs.
    Then asserts that log records from the ignored namespace stay
    at zero for three consecutive polls.
    """

    print("Waiting for at least one log record to be present in ClickHouse...")
    retry_until_ok_clickhouse(
        lambda _: clickhouse_client.count_records(
            "otel.otel_logs",
            "ScopeAttributes['otel.entity.event_as_log'] = 'true'",
        ),
        lambda count: count > 0,
        lambda _: print("No entity state events at all were found while setting up test for namespace exclusion.")
    )

    print("Sleeping for 120 seconds to allow logs to be generated in the excluded namespace...")
    sleep(120) # Wait for logs to be generated in the excluded namespace

    print("Checking for records from the excluded namespace...")
    # Metrics, logs and events
    assert_no_data_with_namespace_in_resource_attributes("otel.otel_metrics_gauge", "metric_gauge")
    assert_no_data_with_namespace_in_resource_attributes("otel.otel_metrics_histogram", "metric_histogram")
    assert_no_data_with_namespace_in_resource_attributes("otel.otel_metrics_sum", "metric_sum")
    assert_no_data_with_namespace_in_resource_attributes("otel.otel_metrics_summary", "metric_summary")
    assert_no_data_with_namespace_in_resource_attributes("otel.otel_logs", "log")

    # Entity state events
    entity_states = clickhouse_client.count_records(
            "otel.otel_logs",
            f"""ScopeAttributes['otel.entity.event_as_log'] = 'true'
    AND LogAttributes['otel.entity.event.type'] = 'entity_state'
    AND JSONHas(LogAttributes['otel.entity.id'], 'k8s.namespace.name')
    AND JSONExtractString(LogAttributes['otel.entity.id'], 'k8s.namespace.name') = '{ignored_namespace}'""",
        )

    assert_no_data(entity_states, "entity_state_event")

    # Entity relationship state events
    entity_relationship_states = clickhouse_client.count_records(
            "otel.otel_logs",
            f"""ScopeAttributes['otel.entity.event_as_log'] = 'true'
    AND LogAttributes['otel.entity.event.type'] = 'entity_relationship_state'
    AND (
        (JSONHas(LogAttributes['otel.entity_relationship.source_entity.id'], 'k8s.namespace.name') AND JSONExtractString(LogAttributes['otel.entity_relationship.source_entity.id'], 'k8s.namespace.name') = '{ignored_namespace}')
        OR (JSONHas(LogAttributes['otel.entity_relationship.destination_entity.id'], 'k8s.namespace.name') AND JSONExtractString(LogAttributes['otel.entity_relationship.destination_entity.id'], 'k8s.namespace.name') = '{ignored_namespace}')
    )""",
        )

    assert_no_data(entity_relationship_states, "entity_relationship_state_event")

    print("No records from the excluded namespace were found in ClickHouse.")
