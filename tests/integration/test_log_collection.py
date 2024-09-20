import pytest
import os
import json
from test_utils import get_all_bodies_for_all_sent_content, retry_until_ok, run_shell_command

endpoint = os.getenv("TIMESERIES_MOCK_ENDPOINT", "localhost:8088")
url = f'http://{endpoint}/logs.json'
pod_name = 'dummy-logging-pod'
tested_log = 'testlog-swo-k8s-collector-integration-test'

def setup_function():
    run_shell_command(f'kubectl run {pod_name} --image bash:alpine3.19 -n default -- -ec "while :; do echo \'{tested_log}\'; sleep 5 ; done"')

def teardown_function():
    run_shell_command(f'kubectl delete pod {pod_name} -n default')

def test_logs_generated():
    retry_until_ok(url, assert_test_log_found, print_failure)

def assert_test_log_found(content):
    raw_bodies = get_all_bodies_for_all_sent_content(content)
    test_log_found = any(f'{tested_log}' in body for body in raw_bodies)
    return test_log_found

def print_failure(content):
    raw_bodies = get_all_bodies_for_all_sent_content(content)
    print(f'Failed to find {tested_log}')
    print('All logs in raw_bodies_dump.txt')
    #print(raw_bodies)
    # Dump the raw_bodies to a file
    with open('raw_bodies_dump.txt', 'w') as file:
        # Convert the list to a JSON string for better formatting
        json.dump(raw_bodies, file, indent=4)


