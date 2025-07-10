import json
import os

from test_utils import get_all_bodies_for_all_sent_content, get_all_resources_for_all_sent_content, get_attribute_key_and_value, get_attributes_of_kvmap, has_attribute_with_key_and_value, retry_until_ok, run_shell_command


endpoint = os.getenv("TIMESERIES_MOCK_ENDPOINT", "localhost:8088")
url = f'http://{endpoint}/entitystateevents.json'
pod_name = 'dummy-entitystateevents-pod'
container_name = 'dummy-container'
namespace_name = 'default'
container_entity = 'KubernetesContainer'

pod_manifest = {
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": {
        "name": pod_name
    },
    "spec": {
        "containers": [
            {
                "name": container_name,
                "image": "alpine3.19"
            }
        ]
    }
}

def setup_function():
    run_shell_command(f"kubectl run multi-container-pod --overrides '{json.dumps(pod_manifest)}' --image bash:alpine3.19 -n {namespace_name} -- -ec \"while :; do sleep 5 ; done\"")


def teardown_function():
    run_shell_command(f'kubectl delete pod {pod_name} -n {namespace_name}')


def test_entity_state_events_generated():
    retry_until_ok(url, assert_test_entitystateevents_found, print_failure)


def print_failure(content):
    raw_bodies = get_all_bodies_for_all_sent_content(content)
    print(f'Failed to find expected container within {pod_name}')
    print('All logs in raw_bodies_dump.txt')
    with open('raw_bodies_dump.txt', 'w') as file:
        json.dump(raw_bodies, file, indent=4)


def assert_test_entitystateevents_found(content):
    logs = get_all_resources_for_all_sent_content(content)
    for resource_logs in logs:
        for resource in resource_logs:
            scope_logs = resource['scopeLogs']

            for scope_log in scope_logs:
                has_correct_scope_attributes_set(scope_log)
                log_records = scope_log['logRecords']

                for log_record in log_records:
                    if not has_correct_log_record_attributes(log_record):
                        continue
                    if has_all_attributes_for_entity(log_record):
                        return True
    return False


def has_all_attributes_for_entity(log_record):
    entity = get_attribute_key_and_value(log_record, 'otel.entity.type')
    if (entity == container_entity):
        return has_id_attributes_for_container(log_record)
    return False


def has_correct_scope_attributes_set(scope_log):
    scope = scope_log['scope']
    if not has_attribute_with_key_and_value(scope, 'otel.entity.event_as_log', True):
        raise Exception('Attribute "otel.entity.event_as_log" is not set')

def has_correct_log_record_attributes(log_record):
    if has_attribute_with_key_and_value(log_record, 'otel.entity.event.type', 'entity_relationship_state'):
        print('Entity relationship state event found, skipping')
        return False
    if not has_attribute_with_key_and_value(log_record, 'otel.entity.event.type', 'entity_state'):
        raise Exception('Attribute "otel.entity.event.type" has unexpected value')

    return True


def has_id_attributes_for_container(log_record):
    id_attrs = get_attributes_of_kvmap(log_record, 'otel.entity.id')
    if id_attrs['k8s.pod.name'] != pod_name:
        return False
    if id_attrs['k8s.namespace.name'] != namespace_name:
        print('Container has incorrect namespace set')
        return False
    if id_attrs['k8s.container.name'] != container_name:
        print('Container has unexpected name')
        return False
    
    attrs = get_attributes_of_kvmap(log_record, 'otel.entity.attributes')
    # wait until container status is loaded
    if attrs['sw.k8s.container.status'] == '':
        return False
    
    return True
