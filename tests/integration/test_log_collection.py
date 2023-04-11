import pytest
import os
from test_utils import get_all_bodies_for_all_sent_content, retry_until_ok, run_shell_command

endpoint = os.getenv("TIMESERIES_MOCK_ENDPOINT", "localhost:8088")
url = f'http://{endpoint}/logs.json'
pod_name = 'dummy-logging-pod'
tested_log = '!!testlog!!'

def setup_function():
    run_shell_command(f'kubectl run {pod_name} --image bash:alpine3.16 -- -ec "while :; do echo \'{tested_log}\'; sleep 5 ; done"')

def teardown_function():
    run_shell_command(f'kubectl delete pod {pod_name}')

def test_logs_generated():
    retry_until_ok(url, assert_test_log_found, print_failure)

def assert_test_log_found(content):
    raw_bodies = get_all_bodies_for_all_sent_content(content)
    test_log_found = any(f'{tested_log}\n' in body for body in raw_bodies)
    return test_log_found

def print_failure(content):
    raw_bodies = get_all_bodies_for_all_sent_content(content)
    print(f'Failed to find {tested_log}')
    print('Sent logs:')
    print(raw_bodies)

