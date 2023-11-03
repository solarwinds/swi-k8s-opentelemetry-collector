from prometheus_client import Gauge, Metric
import pytest
import os
import json
from test_utils import retry_until_ok, get_merged_json, datapoint_value
from prometheus_client.parser import text_string_to_metric_families
import difflib

endpoint = os.getenv("TIMESERIES_MOCK_ENDPOINT", "localhost:8088")
ci = os.getenv("CI", "")
url = f'http://{endpoint}/metrics.json'

endpointPrometheus = os.getenv("PROMETHEUS_MOCK_ENDPOINT", "localhost:8080")
urlMetrics = [f'http://{endpointPrometheus}/metrics',
              f'http://{endpointPrometheus}/federate?match%5B%5D=container_cpu_usage_seconds_total&match%5B%5D=container_spec_cpu_quota&match%5B%5D=container_spec_cpu_period&match%5B%5D=container_memory_working_set_bytes&match%5B%5D=container_spec_memory_limit_bytes&match%5B%5D=container_cpu_cfs_throttled_periods_total&match%5B%5D=container_cpu_cfs_periods_total&match%5B%5D=container_fs_reads_total&match%5B%5D=container_fs_writes_total&match%5B%5D=container_fs_reads_bytes_total&match%5B%5D=container_fs_writes_bytes_total&match%5B%5D=container_fs_usage_bytes&match%5B%5D=container_network_receive_bytes_total&match%5B%5D=container_network_transmit_bytes_total&match%5B%5D=container_network_receive_packets_total&match%5B%5D=container_network_transmit_packets_total&match%5B%5D=container_network_receive_packets_dropped_total&match%5B%5D=container_network_transmit_packets_dropped_total&match%5B%5D=apiserver_request_total&match%5B%5D=kubelet_volume_stats_available_percent&match%5B%5D=%7B__name__%3D%22kubernetes_build_info%22%2C+job%3D~%22.%2Aapiserver.%2A%22%7D']


def test_expected_metric_names_are_generated():
    expected_metric_names = []

    with open(os.path.join(os.path.dirname(__file__), 'expected_metric_names.txt'), "r", newline='\n') as file_with_expected_metric_names:
        expected_metric_names = file_with_expected_metric_names.read().splitlines()

    retry_until_ok(url, 
                   lambda content: assert_metric_names_found(content, expected_metric_names),
                   lambda content: print_failure_metric_names(content, expected_metric_names))
    
def test_expected_network_metric_names_are_generated():
    if ci != "true":
        pytest.skip("Skipping this test on local environment")
    else:
        expected_metric_names = ["k8s.tcp.bytes"]

        retry_until_ok(url, 
                    lambda content: assert_metric_names_found(content, expected_metric_names),
                    lambda content: print_failure_metric_names(content, expected_metric_names))

test_cases = [
    {
        "metrics": [
            { "name": "k8s.container.cpu.usage.seconds.rate" },
            { "name": "k8s.container.status" },
            { "name": "k8s.container_cpu_usage_seconds_total" },
            { "name": "k8s.container_memory_working_set_bytes" },
            { "name": "k8s.container_spec_cpu_period" },
            { "name": "k8s.container_spec_memory_limit_bytes" },
            { "name": "k8s.kube_pod_container_info" },
            { "name": "k8s.kube_pod_container_state_started" },
            { "name": "k8s.kube_pod_container_status_ready" },
            { "name": "k8s.kube_pod_container_status_restarts_total" },
            { "name": "k8s.kube_pod_container_status_running" },
            { "name": "k8s.kube_pod_container_status_terminated" },
            { "name": "k8s.kube_pod_container_status_waiting" },
            { "name": "k8s.kube_pod_created" },
            { "name": "k8s.kube_pod_info" },
            { "name": "k8s.kube_pod_owner" },
            { "name": "k8s.kube_pod_start_time" },
            { "name": "k8s.kube_pod_status_phase" },
            { "name": "k8s.kube_pod_status_ready" },
            { "name": "k8s.pod.containers" },
            { "name": "k8s.pod.containers.running" },
        ],
        "resource_attributes": [
            "sw.k8s.cluster.uid",
            { "key":"k8s.cluster.name", "value": "cluster name" },
            { "key":"k8s.namespace.name", "value": "test-namespace" }, 
            { "key":"k8s.pod.name", "value": "test-pod" }, 
            { "key":"k8s.pod.labels.app", "value": "test-pod"},
            { "key":"k8s.pod.annotations.test-annotation", "value": "test-value"},
        ],
    },
    {
        "metrics": [
            { "name": "k8s.container.status" },
            { "name": "k8s.kube_pod_container_info" },
            { "name": "k8s.kube_pod_container_state_started" },
            { "name": "k8s.kube_pod_container_status_ready" },
            { "name": "k8s.kube_pod_container_status_restarts_total" },
            { "name": "k8s.kube_pod_container_status_running" },
            { "name": "k8s.kube_pod_container_status_terminated" },
            { "name": "k8s.kube_pod_container_status_waiting" },
        ],
        "resource_attributes": [
            "sw.k8s.cluster.uid",
            { "key":"k8s.cluster.name", "value": "cluster name" },
            { "key":"k8s.namespace.name", "value": "test-namespace" }, 
            { "key":"k8s.pod.name", "value": "test-pod" },
            { "key":"k8s.container.name", "value": "test-container" },
        ],
    },
    {
        "metrics": [
            { "name": "k8s.kube.pod.owner.daemonset" },
            { "name": "k8s.kube_daemonset_created" },
            { "name": "k8s.kube_daemonset_labels" },
            { "name": "k8s.kube_daemonset_status_current_number_scheduled" },
            { "name": "k8s.kube_daemonset_status_desired_number_scheduled" },
            { "name": "k8s.kube_daemonset_status_number_available" },
            { "name": "k8s.kube_daemonset_status_number_misscheduled" },
            { "name": "k8s.kube_daemonset_status_number_ready" },
            { "name": "k8s.kube_daemonset_status_number_unavailable" },
            { "name": "k8s.kube_daemonset_status_updated_number_scheduled" },
        ],
        "resource_attributes": [
            "sw.k8s.cluster.uid",
            { "key":"k8s.cluster.name", "value": "cluster name" },
            { "key":"k8s.namespace.name", "value": "test-namespace" }, 
            { "key":"k8s.daemonset.name", "value": "test-daemonset" }, 
            { "key":"k8s.daemonset.labels.app", "value": "test-daemonset"},
            { "key":"k8s.daemonset.annotations.test-annotation", "value": "test-value"},
        ],
    },
    {
        "metrics": [
            { "name": "k8s.deployment.condition.available" },
            { "name": "k8s.deployment.condition.progressing" },
            { "name": "k8s.kube.pod.owner.replicaset" },
            { "name": "k8s.kube.replicaset.owner.deployment" },
            { "name": "k8s.kube_deployment_created" },
            { "name": "k8s.kube_deployment_labels" },
            { "name": "k8s.kube_deployment_spec_paused" },
            { "name": "k8s.kube_deployment_spec_replicas" },
            { "name": "k8s.kube_deployment_status_condition" },
            { "name": "k8s.kube_deployment_status_replicas" },
            { "name": "k8s.kube_deployment_status_replicas_available" },
            { "name": "k8s.kube_deployment_status_replicas_ready" },
            { "name": "k8s.kube_deployment_status_replicas_unavailable" },
            { "name": "k8s.kube_deployment_status_replicas_updated" },
        ],
        "resource_attributes": [
            "sw.k8s.cluster.uid",
            { "key":"k8s.cluster.name", "value": "cluster name" },
            { "key":"k8s.namespace.name", "value": "test-namespace" }, 
            { "key":"k8s.deployment.name", "value": "test-deployment" }, 
            { "key":"k8s.deployment.labels.app", "value": "test-deployment"},
            { "key":"k8s.deployment.annotations.test-annotation", "value": "test-value"},
        ],
    },
    {
        "metrics": [
            { "name": "k8s.kube.pod.owner.statefulset" },
        ],
        "resource_attributes": [
            "sw.k8s.cluster.uid",
            { "key":"k8s.cluster.name", "value": "cluster name" },
            { "key":"k8s.namespace.name", "value": "test-namespace" }, 
            { "key":"k8s.statefulset.name", "value": "test-statefulset" }, 
            { "key":"k8s.statefulset.labels.app", "value": "test-statefulset"},
            { "key":"k8s.statefulset.annotations.test-annotation", "value": "test-value"},
        ],
    },
    {
        "metrics": [
            { "name": "k8s.kube.pod.owner.replicaset" },
        ],
        "resource_attributes": [
            "sw.k8s.cluster.uid",
            { "key":"k8s.cluster.name", "value": "cluster name" },
            { "key":"k8s.namespace.name", "value": "test-namespace" }, 
            { "key":"k8s.replicaset.name", "value": "test-replicaset" }, 
            { "key":"k8s.replicaset.labels.app", "value": "test-replicaset"},
            { "key":"k8s.replicaset.annotations.test-annotation", "value": "test-value"},
        ],
    },
    {
        "metrics": [
            { "name": "k8s.kube.job.owner.cronjob" },
            { "name": "k8s.kube.pod.owner.job" },
        ],
        "resource_attributes": [
            "sw.k8s.cluster.uid",
            { "key":"k8s.cluster.name", "value": "cluster name" },
            { "key":"k8s.namespace.name", "value": "test-namespace" }, 
            { "key":"k8s.cronjob.name", "value": "test-cronjob" }, 
            { "key":"k8s.cronjob.labels.app", "value": "test-cronjob"},
            { "key":"k8s.cronjob.annotations.test-annotation", "value": "test-value"},
        ],
    },
    {
        "metrics": [
            { "name": "k8s.kube_persistentvolume_claim_ref" },
            { "name": "k8s.kube_persistentvolumeclaim_info" },
        ],
        "resource_attributes": [
            "sw.k8s.cluster.uid",
            { "key":"k8s.cluster.name", "value": "cluster name" },
            { "key":"k8s.namespace.name", "value": "test-namespace" }, 
            { "key":"k8s.persistentvolume.name", "value": "test-pv" }, 
            { "key":"k8s.persistentvolume.labels.type", "value": "local"},
            { "key":"k8s.persistentvolume.annotations.example.com/annotation", "value": "example-annotation"},
        ],
    },
    {
        "metrics": [
            { "name": "k8s.kube_persistentvolume_claim_ref" },
            { "name": "k8s.kube_persistentvolumeclaim_access_mode" },
            { "name": "k8s.kube_persistentvolumeclaim_created" },
            { "name": "k8s.kube_persistentvolumeclaim_info" },
            { "name": "k8s.kube_persistentvolumeclaim_resource_requests_storage_bytes" },
            { "name": "k8s.kube_persistentvolumeclaim_status_phase" },
            { "name": "k8s.persistentvolumeclaim.status.phase" },
        ],
        "resource_attributes": [
            "sw.k8s.cluster.uid",
            { "key":"k8s.cluster.name", "value": "cluster name" },
            { "key":"k8s.namespace.name", "value": "test-namespace" }, 
            { "key":"k8s.persistentvolumeclaim.name", "value": "test-pvc" }, 
            { "key":"k8s.persistentvolumeclaim.labels.example.com/label", "value": "example-label"},
            { "key":"k8s.persistentvolumeclaim.annotations.example.com/annotation", "value": "example-annotation"},
        ],
    },
    {
        "metrics": [
            { "name": "k8s.kube_endpoint_address_available" },
            { "name": "k8s.kube_endpoint_address_not_ready" },
            { "name": "k8s.kube_endpoint_created" },
            { "name": "k8s.kube_endpoint_info" },
            { "name": "k8s.kube_service_created" },
            { "name": "k8s.kube_service_info" },
            { "name": "k8s.kube_service_spec_type" },
        ],
        "resource_attributes": [
            "sw.k8s.cluster.uid",
            { "key":"k8s.cluster.name", "value": "cluster name" },
            { "key":"k8s.namespace.name", "value": "test-namespace" }, 
            { "key":"k8s.service.name", "value": "test-service" }, 
            { "key":"k8s.service.labels.example.com/label", "value": "example-label"},
            { "key":"k8s.service.annotations.example.com/annotation", "value": "example-annotation"},
        ],
    },
]


@pytest.mark.parametrize("test_case", test_cases)
def test_expected_otel_message_content_is_generated(test_case):
    resource_attributes = test_case["resource_attributes"]
    metrics = test_case["metrics"]

    retry_until_ok(url, 
                   lambda content: assert_test_contain_expected_datapoints(content, metrics, resource_attributes),
                   print_failure_otel_content,
                   )

def test_no_metric_datapoints_for_internal_containers():
    retry_until_ok(url, assert_test_no_metric_datapoints_for_internal_containers,
                   print_failure_internal_containers)

def assert_test_original_metrics(otelContent):     
    merged_json = get_merged_json(otelContent)

    #transpose otel to metrics again so we can compare
    # it will be record "name"-[metrics]
    metrics = {}
    for resource in merged_json['resourceMetrics']:
        resAttributes={}
        for attr in resource['resource']['attributes']: 
            resAttributes[attr['key']] = attr['value']['stringValue']
        for scope in resource['scopeMetrics']:
            for metric in scope['metrics']:        
                metricName = metric['name'].replace('k8s.', '')
                if( '.' in metricName ):
                    continue
                m = Metric(metricName, '', 'gauge')
                list = metrics.setdefault(m.name, [])
                list.append(m)
                dataPoints = {}
                if( 'gauge' in metric ):
                    dataPoints = metric['gauge']['dataPoints']
                elif ('sum' in metric) :
                    dataPoints = metric['sum']['dataPoints']                                        
                else :
                    raise Exception('unknown data')
            
                for dataPoint in dataPoints:
                    attributes = resAttributes.copy()
                    for attr in dataPoint['attributes']:                    
                        attributes[attr['key']] = attr['value']['stringValue']
                    m.add_sample(m.name, attributes, datapoint_value(dataPoint), dataPoint['timeUnixNano'])            
    
    for url in urlMetrics :
        retry_until_ok(url, lambda metricsContent: assert_prometheus_metrics(metricsContent, metrics), '')
        
    return (True, '')

def assert_prometheus_metrics(metricsContent, metrics):     
    ok = True
    error = ''
    for family in text_string_to_metric_families(metricsContent.decode('utf-8')):
        if( family.name in metrics):
            list = metrics[family.name]
            for sample in family.samples:
                # try to find metric which has same labels
                found = False
                missing_items={}
                for m in list: 
                    for s in m.samples:
                        missing_items = {key: sample.labels[key] for key in set(sample.labels) - set(s.labels) if sample.labels[key] != s.labels.get(key)}

                        # we are dropping metrics with this attribute
                        if( missing_items.get('container') == 'POD'):
                            found = True
                            break

                        # we are removing these prometheus attributes from all datapoints
                        missing_items.pop('prometheus_replica')
                        missing_items.pop('prometheus')
                        missing_items.pop('endpoint', '')

                        #ignore instance,job as they are dropped by mock receiver
                        missing_items.pop('instance')
                        missing_items.pop('job')

                        if len(missing_items.items()) == 0:
                            found = True
                            break
                if not found :                    
                    ok = False
                    error = f'Metric {sample.name} is missing following attributes:'
                    for key in missing_items:
                        error += f'\n\t{key}:{missing_items[key]}'                        
    
    return (ok, error)

def assert_metric_names_found(content, expected_metric_names):
    merged_json = get_merged_json(content)

    metric_names = get_unique_metric_names(merged_json)
    if len(metric_names) == 0:
        return False

    write_actual = os.getenv("WRITE_ACTUAL", "False")
    if write_actual == "True":
        with open(os.path.join(os.path.dirname(__file__), 'expected_metric_names.txt'), "w", newline='\n') as f:
            f.write("\n".join(sorted(metric_names)))

    metric_matches = False
    error = ''
    if all(name in metric_names for name in expected_metric_names):
        print("All specific metric names are found in the response.")
        metric_matches = True
    else:
        missing_metric_names = [
            name for name in expected_metric_names if name not in metric_names]

        error = f'Some specific metric names are not found in the response. \
Missing metrics: {missing_metric_names}'

    return (metric_matches, error)

def print_failure_metric_names(content, expected_metric_names):
    print(f'Failed to find some of expected metric names')
    print(expected_metric_names)

def assert_test_contain_expected_datapoints(content, metrics, resource_attributes):
    merged_json = get_merged_json(content)

    for metric_in_test_case in metrics:
        test_case_passed = False

        # Loop through each resource
        for resource in merged_json["resourceMetrics"]:
            # If the test case has passed, no need to check further resources
            if test_case_passed:
                break

            # Create a dictionary of resource attributes for easier access
            resource_attr_dict = {
                attr["key"]: attr["value"]["stringValue"]
                for attr in resource["resource"]["attributes"]
            }

            # Check if resource has all attributes with non-empty values or specific values
            attributes_match = True
            for attribute in resource_attributes:
                if isinstance(attribute, dict):
                    # Check both key and value for dict-type attributes
                    key = attribute['key']
                    value = attribute['value']
                    if not (key in resource_attr_dict and resource_attr_dict[key] == value):
                        attributes_match = False
                        break
                else:
                    # Check only key for string-type attributes
                    if not (attribute in resource_attr_dict and resource_attr_dict[attribute]):
                        attributes_match = False
                        break

            if attributes_match:
                print("Found resource with all attributes")

                # Loop through each scope
                for scope in resource["scopeMetrics"]:
                    # Loop through each metric
                    for metric in scope["metrics"]:
                        if metric["name"] == metric_in_test_case["name"]:
                            print("Found metric with name")
                            test_case_passed = True  # Mark as passed since metric is found

                            # Default to empty list if 'attributes' key is not present
                            metric_attributes = metric_in_test_case.get("attributes", [])
                            if metric_attributes:  # If attributes list is not empty
                                if 'gauge' in metric:
                                    dataPoints = metric['gauge']['dataPoints']
                                elif 'sum' in metric:
                                    dataPoints = metric['sum']['dataPoints']
                                else:
                                    raise Exception('unknown data type for metric')

                                # Loop through each datapoint
                                for datapoint in dataPoints:
                                    datapoint_attr_dict = {
                                        attr["key"]: attr["value"]["stringValue"]
                                        for attr in datapoint["attributes"]
                                    }
                                    # Check datapoints for the specified attribute keys
                                    if all(key in datapoint_attr_dict and datapoint_attr_dict[key] for key in metric_in_test_case["attributes"]):
                                        print("Found datapoint with all attributes")
                                        break  # Found the required datapoint, break the loop

                            # If metric is found and no attributes are specified, no need to check datapoints
                            if test_case_passed:
                                break  # Metric found, break the metric loop

                    if test_case_passed:
                        break  # Metric found, break the scope loop

            if test_case_passed:
                break  # Metric found, break the resource loop

        if not test_case_passed:
            return (False, f'Failed to find metric {metric_in_test_case["name"]} in resource group')

    return (True, '')

def print_failure_otel_content(content):
    print(f'Failed to find some metrics in some resource groups')

def get_unique_metric_names(merged_json):
    result = list(set([metric["name"]
                       for resource in merged_json["resourceMetrics"]
                       for metric in resource["scopeMetrics"][0]["metrics"]
                       ]))
    return result


def assert_test_no_metric_datapoints_for_internal_containers(content):
    merged_json = get_merged_json(content)

    container_names = get_unique_container_names(merged_json)
    if "POD" in container_names:        
        return (False, 'The response contains datapoints for internal "POD" containers')
    else:
        return (True, '')


def print_failure_internal_containers(content):
    print(f'Failed to find some of internal pod containers')

def get_unique_container_names(merged_json):
    container_names = set()
    for resource in merged_json["resourceMetrics"]:
        if 'attributes' in resource['resource']:
            for resource_attribute in resource['resource']['attributes']:
                if resource_attribute["key"] == "k8s.container.name":
                    container_names.add(resource_attribute["value"]["stringValue"])
    return list(container_names)
