import pytest
import os
from test_utils import get_all_bodies_for_all_sent_content, retry_until_ok, run_shell_command

endpoint = os.getenv("TIMESERIES_MOCK_ENDPOINT", "localhost:8088")
url = f'http://{endpoint}/events.json'
pod_name = 'dummy-pod'
expected_event = f'Started container {pod_name}'

def setup_function():
    run_shell_command(f"kubectl run {pod_name} --labels \"test-label=test-value\" --overrides=\"{{ \\\"apiVersion\\\": \\\"v1\\\", \\\"metadata\\\": {{\\\"annotations\\\": {{ \\\"test-annotation\\\":\\\"test-value\\\" }} }} }}\" --image bash:alpine3.19 -n default -- -ec \"while :; do sleep 5 ; done\"")

def teardown_function():
    run_shell_command(f'kubectl delete pod {pod_name} -n default')

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
