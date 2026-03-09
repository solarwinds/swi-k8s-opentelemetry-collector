import pytest
import os
from test_utils import (
    get_all_bodies_for_clickhouse_logs,
    retry_until_ok_clickhouse,
    run_shell_command,
)
from clickhouse_client import ClickHouseClient

clickhouse_endpoint = os.getenv("CLICKHOUSE_ENDPOINT", "localhost:8123")
clickhouse_client = ClickHouseClient(clickhouse_endpoint)
pod_name = 'dummy-pod'
expected_event = f'Successfully assigned default/{pod_name} to '

def setup_function():
    run_shell_command(f"kubectl run {pod_name} --labels \"test-label=test-value\" --overrides=\"{{ \\\"apiVersion\\\": \\\"v1\\\", \\\"metadata\\\": {{\\\"annotations\\\": {{ \\\"test-annotation\\\":\\\"test-value\\\" }} }} }}\" --image bash:alpine3.19 -n default -- -ec \"while :; do sleep 5 ; done\"")

def teardown_function():
    run_shell_command(f'kubectl delete pod {pod_name} -n default')

def test_events_generated():
    retry_until_ok_clickhouse(
        lambda _attempt: clickhouse_client.get_logs(),
        assert_test_event_found,
        print_failure
    )

def assert_test_event_found(logs_list):
    raw_bodies = get_all_bodies_for_clickhouse_logs(logs_list)
    all_strings = [s for body in raw_bodies for s in body]
    test_event_found = any(s.startswith(expected_event) for s in all_strings)
    return test_event_found

def print_failure(logs_list):
    raw_bodies = get_all_bodies_for_clickhouse_logs(logs_list)
    print(f'Failed to find "{expected_event}"')
    print('Sent events:')
    print(raw_bodies)


