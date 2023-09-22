from prometheus_client import Gauge, Metric
import pytest
import os
import json
from test_utils import retry_until_ok, get_merged_json, datapoint_value
from prometheus_client.parser import text_string_to_metric_families
import difflib

endpoint = os.getenv("TIMESERIES_MOCK_ENDPOINT", "localhost:8088")
url = f'http://{endpoint}/metrics.json'


endpointPrometheus = os.getenv("PROMETHEUS_MOCK_ENDPOINT", "localhost:8080")
urlMetrics = [f'http://{endpointPrometheus}/metrics',
              f'http://{endpointPrometheus}/federate?match%5B%5D=container_cpu_usage_seconds_total&match%5B%5D=container_spec_cpu_quota&match%5B%5D=container_spec_cpu_period&match%5B%5D=container_memory_working_set_bytes&match%5B%5D=container_spec_memory_limit_bytes&match%5B%5D=container_cpu_cfs_throttled_periods_total&match%5B%5D=container_cpu_cfs_periods_total&match%5B%5D=container_fs_reads_total&match%5B%5D=container_fs_writes_total&match%5B%5D=container_fs_reads_bytes_total&match%5B%5D=container_fs_writes_bytes_total&match%5B%5D=container_fs_usage_bytes&match%5B%5D=container_network_receive_bytes_total&match%5B%5D=container_network_transmit_bytes_total&match%5B%5D=container_network_receive_packets_total&match%5B%5D=container_network_transmit_packets_total&match%5B%5D=container_network_receive_packets_dropped_total&match%5B%5D=container_network_transmit_packets_dropped_total&match%5B%5D=apiserver_request_total&match%5B%5D=kubelet_volume_stats_available_percent&match%5B%5D=%7B__name__%3D%22kubernetes_build_info%22%2C+job%3D~%22.%2Aapiserver.%2A%22%7D']

with open('expected_metric_names.txt', "r", newline='\n') as file_with_expected_metric_names:
    expected_metric_names = file_with_expected_metric_names.read().splitlines()

def test_expected_metric_names_are_generated():
    retry_until_ok(url, assert_test_metric_names_found,
                   print_failure_metric_names)
    
def test_expected_otel_message_content_is_generated():
    retry_until_ok(url, assert_test_expected_otel_message_content_is_generated,
                   print_failure_otel_content)


def test_no_metric_datapoints_for_internal_containers():
    retry_until_ok(url, assert_test_no_metric_datapoints_for_internal_containers,
                   print_failure_internal_containers)

def test_original_metrics_are_not_modified():     
        retry_until_ok(url, assert_test_original_metrics, lambda content:print(f'Metrics were modified'))


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

def assert_test_metric_names_found(content):
    merged_json = get_merged_json(content)

    metric_names = get_unique_metric_names(merged_json)
    if (len(metric_names) == 0):
        return False

    write_actual = os.getenv("WRITE_ACTUAL", "False")
    if write_actual == "True":
        with open("expected_metric_names.txt", "w", newline='\n') as f:
            f.write("\n".join(sorted(metric_names)))

    metric_matches = False
    error = ''
    if all(name in metric_names for name in expected_metric_names):
        print("All specific metric names are found in the response.")
        metric_matches = True
    else:
        missing_metric_names = [
            name for name in expected_metric_names if name not in metric_names]

        error = f'Some specific metric names are not found in the response\
Missing metrics: {missing_metric_names}'        

    return (metric_matches, error)


def print_failure_metric_names(content):
    print(f'Failed to find some of expected metric names')
    print(expected_metric_names)


def assert_test_expected_otel_message_content_is_generated(content):    
    with open('expected_output.json', "r", newline='\n') as file_with_expected:
        expected_json_raw = json.load(file_with_expected)

    # do to problems with delayed annotations of pvc we evaluate only last X records    
    content = '\n'.join(content.splitlines()[-5:])    

    merged_json = get_merged_json(content)

    actual_json = json.dumps(merged_json, sort_keys=True, indent=2)
    expected_json = json.dumps(expected_json_raw, sort_keys=True, indent=2)

    write_actual = os.getenv("WRITE_ACTUAL", "False")
    if write_actual == "True":
        with open("actual.json", "w", newline='\n') as f:
            f.write(actual_json)

    length_matches = False
    error = ''
    if actual_json == expected_json:
        print(
            f'Outputs matches, expected chars: {len(expected_json)}, actual chars: {len(actual_json)}')
        length_matches = True
    else:
        error = 'Outputs does not match'
        for line in difflib.unified_diff(
                expected_json.splitlines(), actual_json.splitlines(), lineterm='\n'):
            error += '\n'
            error += line

    return (length_matches, error)


def print_failure_otel_content(content):
    merged_json = get_merged_json(content)
    actual_json = json.dumps(merged_json, sort_keys=True, indent=2)
    print('Actual json:')
    print(actual_json)

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
    print(f'Failed to find some of expected metric names')
    print(expected_metric_names)


def get_unique_container_names(merged_json):
    result = list(set([resource_attribute["value"]["stringValue"]
                       for resource in merged_json["resourceMetrics"]
                       for resource_attribute in resource["resource"]["attributes"] if resource_attribute["key"] == "k8s.container.name"
                       ]))
    return result
