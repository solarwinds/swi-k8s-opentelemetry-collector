import pytest
import os
import json
from jsonmerge import merge
from test_utils import retry_until_ok
import difflib

endpoint = os.getenv("TIMESERIES_MOCK_ENDPOINT", "localhost:8088")
url = f'http://{endpoint}/metrics.json'

with open('expected_metric_names.txt', "r", newline='\n') as file_with_expected_metric_names:
    expected_metric_names = file_with_expected_metric_names.read().splitlines()

def test_metric_names_generated():
    retry_until_ok(url, assert_test_metric_names_found,
                   print_failure_metric_names)


def test_metric_line_count_generated():
    retry_until_ok(url, assert_test_metrics_line_length_match,
                   print_failure_line_count)

def test_no_metric_datapoints_for_internal_containers():
    retry_until_ok(url, assert_test_no_metric_datapoints_for_internal_containers,
                   print_failure_internal_containers)

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
    if all(name in metric_names for name in expected_metric_names):
        print("All specific metric names are found in the response.")
        metric_matches = True
    else:
        missing_metric_names = [
            name for name in expected_metric_names if name not in metric_names]
        print('Some specific metric names are not found in the response')
        print(f'Missing metrics: {missing_metric_names}')

    return metric_matches


def print_failure_metric_names(content):
    print(f'Failed to find some of expected metric names')
    print(expected_metric_names)


def assert_test_metrics_line_length_match(content):
    with open('expected_output.json', "r", newline='\n') as file_with_expected:
        expected_json_raw = json.load(file_with_expected)

    merged_json = get_merged_json(content)

    actual_json = json.dumps(merged_json, sort_keys=True, indent=2)
    expected_json = json.dumps(expected_json_raw, sort_keys=True, indent=2)

    write_actual = os.getenv("WRITE_ACTUAL", "False")
    if write_actual == "True":
        with open("actual.json", "w", newline='\n') as f:
            f.write(actual_json)

    length_matches = False

    if actual_json == expected_json:
        print(
            f'Outputs matches, expected chars: {len(expected_json)}, actual chars: {len(actual_json)}')
        length_matches = True
    else:
        print('Outputs does not match')
        for line in difflib.unified_diff(
                expected_json.splitlines(), actual_json.splitlines(), lineterm='\n'):
            print(line)

    return length_matches


def print_failure_line_count(content):
    merged_json = get_merged_json(content)
    actual_json = json.dumps(merged_json, sort_keys=True, indent=2)
    print('Actual json:')
    print(actual_json)

def resource_sorting_key(metric):
    return "".join([f"{a['key']}={a['value']['stringValue']}" for a in metric["resource"]["attributes"]])

def remove_time_in_datapoint(datapoint):
    if "timeUnixNano" in datapoint:
        datapoint["timeUnixNano"] = "0"
    if "startTimeUnixNano" in datapoint:
        datapoint["startTimeUnixNano"] = "0"

def datapoint_sorting_key(datapoint):
    if "attributes" in datapoint:
        return "".join([f"{a['key']}={a['value']['stringValue']}" for a in datapoint["attributes"]])
    elif "asDouble" in datapoint:
        return datapoint["asDouble"]
    elif "asInt" in datapoint:
        return datapoint["asInt"]
    elif "asString" in datapoint:
        return datapoint["asString"]
    else:
        return datapoint

def merge_jsons(jsons):
    result = {}
    for json_ in jsons:
        result = merge(result, json_)

    # Sort the result and set timeStamps to 0 to make it easier to compare
    result["resourceMetrics"] = sorted(result["resourceMetrics"], key=resource_sorting_key)
    for resource in result["resourceMetrics"]:
        resource["resource"]["attributes"] = sorted(resource["resource"]["attributes"], key=lambda a: a["key"])
        for scope in resource["scopeMetrics"]:
            scope["metrics"] = sorted(scope["metrics"], key=lambda m: m["name"])
            for metric in scope["metrics"]:
                if "sum" in metric:
                    metric["sum"]["dataPoints"] = sorted(metric["sum"]["dataPoints"], key=datapoint_sorting_key)
                    for dp in metric["sum"]["dataPoints"]:
                        remove_time_in_datapoint(dp)
                if "gauge" in metric:
                    metric["gauge"]["dataPoints"] = sorted(metric["gauge"]["dataPoints"], key=datapoint_sorting_key)
                    for dp in metric["gauge"]["dataPoints"]:
                        remove_time_in_datapoint(dp)
                if "histogram" in metric:
                    metric["histogram"]["dataPoints"] = sorted(metric["histogram"]["dataPoints"], key=datapoint_sorting_key)
                    for dp in metric["histogram"]["dataPoints"]:
                        remove_time_in_datapoint(dp)

                # Get rid of value of metric called "scrape_duration_seconds" as it is not stable
                if metric["name"] == "scrape_duration_seconds":
                    metric["gauge"]["dataPoints"][0]["asDouble"] = 0

    return result


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
        print('The response contains datapoints for internal "POD" containers')
        return False
    else:
        return True

def print_failure_internal_containers(content):
    print(f'Failed to find some of expected metric names')
    print(expected_metric_names)

def get_unique_container_names(merged_json):
    result = list(set([resource_attribute["value"]["stringValue"]
                       for resource in merged_json["resourceMetrics"]
                       for resource_attribute in resource["resource"]["attributes"] if resource_attribute["key"] == "k8s.container.name"
                       ]))
    return result

def get_merged_json(content):
    lines = content.splitlines()
    metrics = [json.loads(line) for line in lines]
    return merge_jsons(metrics)