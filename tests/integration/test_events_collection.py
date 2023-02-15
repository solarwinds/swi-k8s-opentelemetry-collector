import pytest
import os
import subprocess
from test_utils import get_all_bodies_for_all_sent_content, retry_until_ok

endpoint = os.getenv("TIMESERIES_MOCK_ENDPOINT", "localhost:8088")
url = f'http://{endpoint}/events.json'
pod_name = 'dummy-pod'
expected_event = f'Started container {pod_name}'

def setup_function():
    subprocess.run(f'kubectl run {pod_name} --image bash:alpine3.16 -- -ec "while :; do sleep 5 ; done"', shell=True)

def teardown_function():
    subprocess.run(f'kubectl delete pod {pod_name}', shell=True)

def test_events_generated():
    retry_until_ok(url, assert_test_event_found, print_failure)

def assert_test_event_found(content):
    raw_bodies = get_all_bodies_for_all_sent_content(content)
    test_event_found = any(expected_event in body for body in raw_bodies)
    return test_event_found

def print_failure(content):
    raw_bodies = get_all_bodies_for_all_sent_content(content)
    print(f'Failed to find "{expected_event}"')
    print('Sent events:')
    print(raw_bodies)


