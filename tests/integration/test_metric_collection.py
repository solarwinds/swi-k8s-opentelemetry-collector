from prometheus_client import Gauge, Metric
import pytest
import os
import json
from test_utils import retry_until_ok, retry_until_ok_clickhouse, get_merged_json, get_merged_json_from_clickhouse, datapoint_value, parse_value
from prometheus_client.parser import text_string_to_metric_families
from clickhouse_client import ClickHouseClient
import difflib

# ClickHouse configuration
clickhouse_endpoint = os.getenv("CLICKHOUSE_ENDPOINT", "localhost:8123")
clickhouse_client = ClickHouseClient(clickhouse_endpoint)

# Legacy endpoint configuration (kept for compatibility)
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

    retry_until_ok_clickhouse(
        lambda: clickhouse_client.get_metrics_otlp(),
        lambda metrics_list: assert_metric_names_found(metrics_list, expected_metric_names),
        lambda metrics_list: print_failure_metric_names(metrics_list, expected_metric_names)
    )
    
test_cases = [
]


@pytest.mark.parametrize("file_name", os.listdir(os.path.join(os.path.dirname(__file__), 'expected_telemetry')))
def test_expected_otel_message_content_is_generated(file_name):
    # Skip files that are not JSON
    if not file_name.endswith('.json'):
        pytest.skip("Skipping non-JSON file")

    # Construct the full file path
    file_path = os.path.join(os.path.dirname(__file__), 'expected_telemetry', file_name)

    # Read the JSON file and parse the test case
    with open(file_path, 'r') as file:
        test_case = json.load(file)

    # Continue with the rest of the test using the parsed test_case
    resource_attributes = test_case["resource_attributes"]
    metrics = test_case["metrics"]

    metric_names = [item['name'] for item in metrics]
    print("Checking metrics {} with resource attributes {}".format(metric_names, resource_attributes))

    retry_until_ok_clickhouse(
        lambda: clickhouse_client.get_metrics_otlp(),
        lambda metrics_list: assert_test_contain_expected_datapoints(metrics_list, metrics, resource_attributes),
        print_failure_otel_content,
        timeout=120
    )

def test_no_metric_datapoints_for_internal_containers():
    retry_until_ok_clickhouse(
        lambda: clickhouse_client.get_metrics_otlp(),
        assert_test_no_metric_datapoints_for_internal_containers,
        print_failure_internal_containers
    )

def assert_test_original_metrics(otelContent):     
    merged_json = get_merged_json(otelContent)

    #transpose otel to metrics again so we can compare
    # it will be record "name"-[metrics]
    metrics = {}
    for json_line in merged_json:
        for resource in json_line['resourceMetrics']:
            resAttributes={}
            for attr in resource['resource']['attributes']: 
                resAttributes[attr['key']] = parse_value(attr['value'])
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
                            attributes[attr['key']] = parse_value(attr['value'])
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

def assert_metric_names_found(metrics_list, expected_metric_names):
    merged_json = get_merged_json_from_clickhouse(metrics_list)

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

def print_failure_metric_names(metrics_list, expected_metric_names):
    print(f'Failed to find some of expected metric names')
    print(expected_metric_names)

def assert_test_contain_expected_datapoints(metrics_list, metrics, resource_attributes):
    merged_json = get_merged_json_from_clickhouse(metrics_list)

    for metric_in_test_case in metrics:
        test_case_passed = False

        # Loop through each resource
        for json_line in merged_json:
            for resource in json_line["resourceMetrics"]:
                # If the test case has passed, no need to check further resources
                if test_case_passed:
                    break

                # Create a dictionary of resource attributes for easier access
                resource_attr_dict = {
                    attr["key"]: parse_value(attr["value"])
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
                    # Loop through each scope
                    for scope in resource["scopeMetrics"]:
                        # Loop through each metric
                        for metric in scope["metrics"]:
                            if metric["name"] == metric_in_test_case["name"]:
                                print(f'Found metric {metric_in_test_case["name"]}')
                                test_case_passed = True

                                # Default to empty list if 'attributes' key is not present
                                metric_attributes = metric_in_test_case.get("attributes", [])
                                if metric_attributes:  # If attributes list is not empty
                                    
                                    test_case_passed = False

                                    if 'gauge' in metric:
                                        dataPoints = metric['gauge']['dataPoints']
                                    elif 'sum' in metric:
                                        dataPoints = metric['sum']['dataPoints']
                                    else:
                                        raise Exception('unknown data type for metric')

                                    # Loop through each datapoint
                                    for datapoint in dataPoints:
                                        datapoint_attr_dict = {
                                            attr["key"]: parse_value(attr["value"])
                                            for attr in datapoint["attributes"]
                                        }
                                        # Check datapoints for the specified attribute keys
                                        if all(key in datapoint_attr_dict and datapoint_attr_dict[key] for key in metric_in_test_case["attributes"]):
                                            print("Found datapoint with all attributes")
                                            test_case_passed = True
                                            break  # Found the required datapoint, break the loop

                                # If metric is found and no attributes are specified, no need to check datapoints
                                if test_case_passed:
                                    break  # Metric found, break the metric loop

                        if test_case_passed:
                            break  # Metric found, break the scope loop

                if test_case_passed:
                    break  # Metric found, break the resource loop

            if test_case_passed:
                break  # Metric found, break the resource loop            

        if not test_case_passed:
            return (False, f'Failed to find metric {metric_in_test_case["name"]}')

    return (True, '')

def print_failure_otel_content(metrics_list):
    print(f'Failed to find some metrics in some resource groups')

def get_unique_metric_names(merged_json):
    result = list(set([metric["name"]
                       for json_line in merged_json
                       for resource in json_line["resourceMetrics"]
                       for metric in resource["scopeMetrics"][0]["metrics"]
                       ]))
    return result


def assert_test_no_metric_datapoints_for_internal_containers(metrics_list):
    merged_json = get_merged_json_from_clickhouse(metrics_list)

    container_names = get_unique_container_names(merged_json)
    if "POD" in container_names:        
        return (False, 'The response contains datapoints for internal "POD" containers')
    else:
        return (True, '')


def print_failure_internal_containers(metrics_list):
    print(f'Failed to find some of internal pod containers')

def get_unique_container_names(merged_json):
    container_names = set()
    for json_line in merged_json:
        for resource in json_line["resourceMetrics"]:
            if 'attributes' in resource['resource']:
                for resource_attribute in resource['resource']['attributes']:
                    if resource_attribute["key"] == "k8s.container.name":
                        container_names.add(parse_value(resource_attribute["value"]))
    return list(container_names)
