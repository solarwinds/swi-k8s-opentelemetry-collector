import json
import os
from test_utils import retry_until_ok_clickhouse
from clickhouse_client import ClickHouseClient

clickhouse_endpoint = os.getenv("CLICKHOUSE_ENDPOINT", "localhost:8123")
clickhouse_client = ClickHouseClient(clickhouse_endpoint)

SERVICE_NAME = "trace-gen-test-service"
timeout = int(os.getenv("TEST_TIMEOUT", "60"))

def test_traces_received_by_gateway():
    """Baseline liveness check: traces from the generator reach ClickHouse via the tail-sampling pipeline."""
    retry_until_ok_clickhouse(
        fetch_func=lambda _attempt: clickhouse_client.get_traces(
            where_clause=f"ServiceName = '{SERVICE_NAME}'"
        ),
        assert_func=lambda spans: len(spans) > 0,
        print_failure=_print_failure,
        timeout=timeout,
    )


def test_error_traces_are_sampled():
    """Error-policy spans (StatusCode=ERROR) must always be sampled and reach ClickHouse."""
    retry_until_ok_clickhouse(
        fetch_func=lambda _attempt: clickhouse_client.get_traces(
            where_clause=f"ServiceName = '{SERVICE_NAME}' AND SpanName = 'error-request'"
        ),
        assert_func=lambda spans: len(spans) > 0,
        print_failure=_print_failure,
        timeout=timeout,
    )


def test_healthcheck_traces_are_dropped():
    """Healthcheck-policy spans must be dropped and must not appear in ClickHouse.

    This test first waits for the pipeline to be live (error spans present), then
    asserts that healthcheck spans are absent. Checking for error spans first
    ensures the pipeline has processed at least one decision_wait window.
    """
    def assert_pipeline_live_and_healthcheck_absent(spans):
        error_spans = clickhouse_client.get_traces(
            where_clause=f"ServiceName = '{SERVICE_NAME}' AND SpanName = 'error-request'"
        )
        if not error_spans:
            return False, "Waiting for pipeline to be live (no error spans yet)"
        if spans:
            return False, f"Expected zero healthcheck spans but found {len(spans)}"
        return True, ""

    retry_until_ok_clickhouse(
        fetch_func=lambda _attempt: clickhouse_client.get_traces(
            where_clause=f"ServiceName = '{SERVICE_NAME}' AND SpanName = 'healthcheck-request'"
        ),
        assert_func=assert_pipeline_live_and_healthcheck_absent,
        print_failure=_print_failure,
        timeout=timeout,
    )


def _print_failure(spans):
    print("Trace assertion failed.")
