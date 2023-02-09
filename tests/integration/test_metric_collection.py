import pytest
import time
import requests
import os
import json
from jsonmerge import merge


def test_metric_names_generated():
    endpoint = os.getenv("TIMESERIES_MOCK_ENDPOINT", "localhost:8088")
    timeout = 120  # set the timeout in seconds
    url = f'http://{endpoint}/metrics.json'
    expected_metric_names = ['scrape_samples_scraped', 'k8s.deployment.condition.available', 'k8s.kube_deployment_status_replicas_available', 'k8s.cluster.nodes', 'k8s.kube_pod_start_time', 'k8s.kube_deployment_status_condition', 'k8s.kube_pod_created', 'k8s.node.status.condition.memorypressure', 'k8s.container_network_transmit_packets_dropped_total', 'k8s.cluster.memory.capacity', 'k8s.cluster.spec.memory.requests', 'k8s.pod.network.receive_packets_dropped', 'k8s.node.status.condition.ready', 'k8s.kube_pod_container_state_started', 'k8s.kube_pod_container_status_ready', 'k8s.cluster.memory.utilization', 'k8s.kube_pod_container_status_terminated', 'k8s.kube_namespace_created', 'k8s.kube_namespace_status_phase', 'k8s.kube_deployment_status_replicas_unavailable', 'k8s.node.fs.iops', 'k8s.node.network.bytes_transmitted', 'k8s.kube_replicaset_created', 'k8s.kube_daemonset_status_desired_number_scheduled', 'k8s.kube_pod_status_phase', 'scrape_duration_seconds', 'k8s.kube.replicaset.owner.deployment', 'k8s.node.network.transmit_packets_dropped', 'k8s.pod.containers.running', 'scrape_samples_post_metric_relabeling', 'k8s.kube_daemonset_status_updated_number_scheduled', 'k8s.kube_node_spec_unschedulable', 'k8s.container.status', 'k8s.kube_pod_container_resource_limits', 'k8s.node.network.packets_transmitted', 'k8s.container_network_receive_packets_dropped_total', 'k8s.kube_pod_container_status_waiting', 'k8s.node.memory.capacity', 'k8s.kube_replicaset_owner', 'k8s.container_fs_reads_bytes_total', 'k8s.container_fs_reads_total', 'k8s.pod.fs.reads.rate', 'k8s.kube_node_info', 'k8s.kube_node_status_ready', 'k8s.node.status.condition.networkunavailable', 'k8s.container_cpu_usage_seconds_total', 'k8s.kube_daemonset_status_current_number_scheduled', 'k8s.node.memory.working_set', 'k8s.kube_daemonset_status_number_unavailable', 'k8s.kube.pod.owner.replicaset', 'k8s.pod.memory.working_set', 'k8s.pod.fs.writes.rate', 'k8s.cluster.nodes.ready', 'k8s.container_memory_working_set_bytes', 'k8s.kube_pod_container_status_restarts_total', 'k8s.deployment.condition.progressing',
                             'k8s.kube_deployment_spec_replicas', 'k8s.kube_node_created', 'k8s.pod.spec.memory.limit', 'k8s.cluster.cpu.utilization', 'k8s.kube_pod_info', 'k8s.pod.network.transmit_packets_dropped', 'k8s.kube_node_status_allocatable', 'k8s.node.memory.allocatable', 'k8s.kube_daemonset_created', 'k8s.cluster.cpu.capacity', 'k8s.node.cpu.capacity', 'k8s.container_spec_cpu_period', 'k8s.container_network_transmit_bytes_total', 'k8s.kube_deployment_status_replicas_ready', 'k8s.node.status.condition.diskpressure', 'k8s.node.network.bytes_received', 'k8s.kube_deployment_created', 'k8s.kube_deployment_status_replicas', 'k8s.kube_node_status_capacity', 'k8s.kube_daemonset_status_number_available', 'k8s.pod.containers', 'k8s.kube_pod_owner', 'k8s.node.cpu.usage.seconds.rate', 'up', 'k8s.kube_deployment_spec_paused', 'k8s.container_network_receive_bytes_total', 'k8s.kube_deployment_labels', 'k8s.kube_daemonset_status_number_ready', 'k8s.container.spec.memory.requests', 'k8s.container_network_transmit_packets_total', 'k8s.node.cpu.allocatable', 'k8s.cluster.memory.allocatable', 'k8s.kube_daemonset_status_number_misscheduled', 'k8s.kube_pod_container_info', 'k8s.container_fs_writes_bytes_total', 'k8s.pod.fs.iops', 'k8s.node.pods', 'k8s.kube_pod_container_status_running', 'k8s.node.network.receive_packets_dropped', 'k8s.node.status.condition.pidpressure', 'k8s.pod.network.bytes_transmitted', 'k8s.pod.spec.memory.requests', 'k8s.cluster.pods.running', 'k8s.cluster.nodes.ready.avg', 'k8s.pod.network.packets_transmitted', 'k8s.kube_daemonset_labels', 'k8s.pod.fs.usage.bytes', 'k8s.container_network_receive_packets_total', 'k8s.kube_node_status_condition', 'k8s.cluster.cpu.allocatable', 'k8s.container_fs_usage_bytes', 'k8s.cluster.pods', 'k8s.container_spec_memory_limit_bytes', 'k8s.kube_pod_status_ready', 'k8s.pod.network.bytes_received', 'scrape_series_added', 'k8s.pod.network.packets_received', 'k8s.node.network.packets_received', 'k8s.pod.cpu.usage.seconds.rate', 'k8s.container_fs_writes_total', 'k8s.kube_pod_container_resource_requests', 'k8s.kube_deployment_status_replicas_updated', 'k8s.node.fs.usage']
    start_time = time.time()

    with open('expected_output.json', "r", newline='\n') as file_with_expected:
        expected_json_raw = json.load(file_with_expected)

    metric_matches = False
    length_matches = False
    while time.time() - start_time < timeout:
        response = None
        try: 
            response = requests.get(url)
            response.raise_for_status()
        except requests.exceptions.RequestException as e:
            print(f"An error occurred while making the request: {e}")

        if response is not None and response.status_code == 200:
            processed_successfully = True
            try:
                print("Successfully downloaded!")
                lines = response.content.splitlines()
                metrics = [json.loads(line) for line in lines]
                merged_json = merge_jsons(metrics)

                actual_json = json.dumps(merged_json, sort_keys=True, indent=2)
                expected_json = json.dumps(expected_json_raw, sort_keys=True, indent=2)

                write_actual = os.getenv("WRITE_ACTUAL_JSON", "False")
                if write_actual == "True":
                    with open("actual.json", "w", newline='\n') as f:
                        f.write(actual_json)

                length_matches = False
                if len(actual_json.splitlines()) == len(expected_json.splitlines()):
                    print(
                        f'LineCount of outputs matches, expected: {len(expected_json.splitlines())}, actual: {len(actual_json.splitlines())}')
                    length_matches = True
                else:
                    print(
                        f'LineCount of outputs does not match, expected: {len(expected_json.splitlines())}, actual: {len(actual_json.splitlines())}')

                metric_names = get_unique_metric_names(merged_json)
                metric_matches = False
                if all(name in metric_names for name in expected_metric_names):
                    print("All specific metric names are found in the response.")
                    metric_matches = True
                else:
                    print('Some specific metric names are not found in the response')
                    print(f'Expected: {expected_metric_names}')
                    print(f'Actual: {metric_names}')
            except Exception as e:
                print('An exception occurred: {}'.format(e))
                processed_successfully = False;


            if processed_successfully and metric_matches and length_matches:
                break
            else:
                print('Retrying...')
                time.sleep(2)
        else:
            if response is not None:
                print('Failed to download metrics. Response code:',
                    response.status_code)
            print('Retrying...')
            time.sleep(2)

    if time.time() - start_time >= timeout:
        if not metric_matches and actual_json is not None:
            print(f'Actual json:')
            print(actual_json)

        raise ValueError("Timed out waiting for specific metric names")


def merge_jsons(jsons):
    result = {}
    for json_ in jsons:
        result = merge(result, json_)
    return result


def get_unique_metric_names(merged_json):
    result = list(set([metric["name"]
                       for resource in merged_json["resourceMetrics"]
                       for metric in resource["scopeMetrics"][0]["metrics"]
                       ]))
    return result
