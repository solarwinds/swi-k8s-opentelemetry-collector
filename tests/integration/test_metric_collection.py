import pytest
import os
import json
from jsonmerge import merge
from test_utils import retry_until_ok

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


def assert_test_metric_names_found(content):
    lines = content.splitlines()
    metrics = [json.loads(line) for line in lines]
    merged_json = merge_jsons(metrics)

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
            name in metric_names for name in expected_metric_names]
        print('Some specific metric names are not found in the response')
        print(f'Missing metrics: {missing_metric_names}')

    return metric_matches


def print_failure_metric_names(content):
    print(f'Failed to find some of expected metric names')
    print(expected_metric_names)


def assert_test_metrics_line_length_match(content):
    with open('expected_output.json', "r", newline='\n') as file_with_expected:
        expected_json_raw = json.load(file_with_expected)

    lines = content.splitlines()
    metrics = [json.loads(line) for line in lines]
    merged_json = merge_jsons(metrics)

    actual_json = json.dumps(merged_json, sort_keys=True, indent=2)
    expected_json = json.dumps(expected_json_raw, sort_keys=True, indent=2)

    write_actual = os.getenv("WRITE_ACTUAL", "False")
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

    return length_matches


def print_failure_line_count(content):
    lines = content.splitlines()
    metrics = [json.loads(line) for line in lines]
    merged_json = merge_jsons(metrics)
    actual_json = json.dumps(merged_json, sort_keys=True, indent=2)
    print('Actual json:')
    print(actual_json)


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
