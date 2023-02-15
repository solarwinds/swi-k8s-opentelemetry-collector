import pytest
import os
import subprocess
from test_utils import get_all_bodies_for_all_sent_content, retry_until_ok

endpoint = os.getenv("TIMESERIES_MOCK_ENDPOINT", "localhost:8088")
url = f'http://{endpoint}/logs.json'
pod_name = 'dummy-logging-pod'
tested_log = '!!testlog!!'

def setup_function():
    subprocess.run(f'kubectl run {pod_name} --image ubuntu -- /bin/bash -ec "while :; do echo \'{tested_log}\'; sleep 5 ; done"', shell=True)

def teardown_function():
    subprocess.run(f'kubectl delete pod {pod_name}', shell=True)

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

