import pytest
import os
import json
from test_utils import (
    get_all_bodies_for_clickhouse_logs,
    retry_until_ok_clickhouse,
    run_shell_command,
)
from clickhouse_client import ClickHouseClient

clickhouse_endpoint = os.getenv("CLICKHOUSE_ENDPOINT", "localhost:8123")
clickhouse_client = ClickHouseClient(clickhouse_endpoint)
pod_name = 'dummy-logging-pod'
tested_log = 'testlog-swo-k8s-collector-integration-test'

def setup_function():
    run_shell_command(f'kubectl run {pod_name} --image bash:alpine3.19 -n default -- -ec "while :; do echo \'{tested_log}\'; sleep 5 ; done"')

def teardown_function():
    run_shell_command(f'kubectl delete pod {pod_name} -n default')

def test_logs_generated():
    retry_until_ok_clickhouse(
        lambda: clickhouse_client.get_logs(),
        assert_test_log_found,
        print_failure
    )

def assert_test_log_found(logs_list):
    raw_bodies = get_all_bodies_for_clickhouse_logs(logs_list)
    test_log_found = any(
        f'{tested_log}' in entry
        for body in raw_bodies
        for entry in body
    )
    return test_log_found

def print_failure(logs_list):
    raw_bodies = get_all_bodies_for_clickhouse_logs(logs_list)
    print(f'Failed to find {tested_log}')
    print('All logs in raw_bodies_dump.txt')
    # Dump the raw_bodies to a file
    with open('raw_bodies_dump.txt', 'w') as file:
        # Convert the list to a JSON string for better formatting
        json.dump(raw_bodies, file, indent=4)



