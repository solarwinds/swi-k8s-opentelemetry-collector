import pytest
import os
from test_utils import get_all_bodies_for_all_sent_content, get_all_resources_for_all_sent_content, retry_until_ok, run_shell_command

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

def test_events_has_labels():
    retry_until_ok(url, assert_test_event_label_found, print_labels_failure)

def assert_test_event_found(content):
    raw_bodies = get_all_bodies_for_all_sent_content(content)
    test_event_found = any(expected_event in body for body in raw_bodies)
    return test_event_found

def print_failure(content):
    raw_bodies = get_all_bodies_for_all_sent_content(content)
    print(f'Failed to find "{expected_event}"')
    print('Sent events:')
    print(raw_bodies)

def assert_test_event_label_found(content):
    raw_bodies = get_all_resources_for_all_sent_content(content)
    resource = find_resource_with_specific_body(raw_bodies, expected_event)
    print(resource)

    if resource is not None:
        return (has_attribute_with_key_and_value(resource, "k8s.pod.labels.test-label", "test-value") and
                does_not_have_attribute_with_key(resource, "k8s.pod.annotations.test-annotation"))
    else:
        print("Resource not found.")
        return False

def print_labels_failure(content):
    raw_bodies = get_all_resources_for_all_sent_content(content)
    print(f'Failed to find "{expected_event}"')
    print('Sent events:')
    print(raw_bodies)

def find_resource_with_specific_body(raw_bodies, target_body):
    for inner_list in raw_bodies:
        for obj in inner_list:
            scope_logs = obj.get("scopeLogs", [])
            for scope_log in scope_logs:
                log_records = scope_log.get("logRecords", [])
                for log_record in log_records:
                    body = log_record.get("body", {}).get("stringValue", "")
                    if target_body in body:
                        return obj["resource"]

    return None

def has_attribute_with_key_and_value(resource, target_key, expected_value):
    attributes = resource.get("attributes", [])
    for attribute in attributes:
        key = attribute.get("key", "")
        value = attribute.get("value", {}).get("stringValue", "")
        if key == target_key and value == expected_value:
            print(f"Resource has attribute with key '{target_key}' and value '{expected_value}'.")
            return True

    print(f"Resource does not have attribute with key '{target_key}' and value '{expected_value}'.")
    return False

def does_not_have_attribute_with_key(resource, target_key):
    attributes = resource.get("attributes", [])
    for attribute in attributes:
        key = attribute.get("key", "")
        if key == target_key:
            print(f"Resource has attribute with key '{target_key}'.")
            return False

    print(f"Resource does not have attribute with key '{target_key}'.")
    return True


