import json
import time
import requests
import traceback
import subprocess
import re

def get_all_log_resources(log_bulk):
    result = [resource
              for resource in log_bulk["resourceLogs"]
              ]
    return result

def get_all_bodies(log_bulk):
    result = [records["body"]["stringValue"]
              for resource in log_bulk["resourceLogs"]
              for scope in resource["scopeLogs"]
              for records in scope["logRecords"]
              ]
    return result

def get_all_bodies_for_all_sent_content(content):
    lines = content.splitlines()
    log_bulks = [json.loads(line) for line in lines]
    return [get_all_bodies(log_bulk) for log_bulk in log_bulks]

def get_all_resources_for_all_sent_content(content):
    lines = content.splitlines()
    log_bulks = [json.loads(line) for line in lines]
    return [get_all_log_resources(log_bulk) for log_bulk in log_bulks]


def retry_until_ok(url, func, print_failure):
    timeout = 120  # set the timeout in seconds
    start_time = time.time()
    last_exception = None
    while time.time() - start_time < timeout:
        is_ok = False
        try: 
            response = None
            try: 
                response = requests.get(url)
                response.raise_for_status()
            except requests.exceptions.RequestException as e:
                print(f"An error occurred while making the request: {e}")

            if response is not None and response.status_code == 200:
                print("Successfully downloaded!")
                is_ok = func(response.content)
            else:
                if response is not None:
                    print('Failed to download otel messages. Response code:',
                        response.status_code)

                print('Failed to download otel messages')
        except Exception as e:
            last_exception = e
            print(e, traceback.format_exc())
        if is_ok:
            print(f'Succesfully passed assert')
            break
        else:
            print('Retrying...')
            time.sleep(2)

    if time.time() - start_time >= timeout:
        if last_exception is not None:
            print('Last exception: {}'.format(last_exception))
        
        if response is not None:
            print_failure(response.content)

        raise ValueError("Timed out waiting")
    
def get_hash_key_by_attributes(obj):
    sorted_attributes = sorted(obj["attributes"], key=lambda a: a["key"])
    return "".join([f"{a['key']}={a['value']['stringValue']}" for a in sorted_attributes])


def resource_sorting_key(resource):
    return get_hash_key_by_attributes(resource["resource"])

def datapoint_sorting_key(datapoint):
    if "attributes" in datapoint:
        return get_hash_key_by_attributes(datapoint)
    elif "asDouble" in datapoint:
        return datapoint["asDouble"]
    elif "asInt" in datapoint:
        return datapoint["asInt"]
    elif "asString" in datapoint:
        return datapoint["asString"]
    else:
        return datapoint

def remove_time_in_datapoint(datapoint):
    if "timeUnixNano" in datapoint:
        datapoint["timeUnixNano"] = "0"
    if "startTimeUnixNano" in datapoint:
        datapoint["startTimeUnixNano"] = "0"


def sort_attributes(element):
    if "attributes" in element:
        element["attributes"] = sorted(
            element["attributes"], key=lambda a: a["key"])

def sort_datapoints(metric):
    metric["dataPoints"] = sorted(
        metric["dataPoints"], key=datapoint_sorting_key)


def process_metric_type(metric):
    if "dataPoints" in metric:
        for dp in metric["dataPoints"]:
            remove_time_in_datapoint(dp)
            sort_attributes(dp)
        sort_datapoints(metric)

def merge_datapoints(existing_datapoints, new_datapoints):
    existing_datapoints_dict = {datapoint_sorting_key(dp): dp for dp in existing_datapoints}

    for new_datapoint in new_datapoints:
        new_datapoint_hash_key = datapoint_sorting_key(new_datapoint)

        if new_datapoint_hash_key in existing_datapoints_dict:
            existing_datapoints_dict[new_datapoint_hash_key].update(new_datapoint)
        else:
            existing_datapoints.append(new_datapoint)

def merge_metrics(existing_metric, new_metric):
    metric_types = ["sum", "gauge", "histogram"]

    for metric_type in metric_types:
        if metric_type in existing_metric and metric_type in new_metric:
            existing_datapoints = existing_metric[metric_type]["dataPoints"]
            new_datapoints = new_metric[metric_type]["dataPoints"]
            merge_datapoints(existing_datapoints, new_datapoints)

def merge_scope_metrics(existing_scope, new_scope):
    existing_metrics = {metric["name"]: metric for metric in existing_scope["metrics"]}
    
    for new_metric in new_scope["metrics"]:
        if new_metric["name"] in existing_metrics:
            merge_metrics(existing_metrics[new_metric["name"]], new_metric)
        else:
            existing_scope["metrics"].append(new_metric)

def merge_resources(existing_resource, new_resource):
    existing_scopes = existing_resource["scopeMetrics"]
    new_scopes = new_resource["scopeMetrics"]

    for new_scope in new_scopes:
        for existing_scope in existing_scopes:
            merge_scope_metrics(existing_scope, new_scope)
            break
        else:
            existing_scopes.append(new_scope)
            
def custom_json_merge(result, new_json):
    new_resources = {resource_sorting_key(resource): resource for resource in new_json["resourceMetrics"]}
    for existing_resource in result["resourceMetrics"]:
        existing_key = resource_sorting_key(existing_resource)
        if existing_key in new_resources:
            merge_resources(existing_resource, new_resources.pop(existing_key))

    result["resourceMetrics"].extend(new_resources.values())

def get_merged_json(content):
    result = {"resourceMetrics": []}
    for line in content.splitlines():
        custom_json_merge(result, json.loads(line))

    # Sort the result and set timeStamps to 0 to make it easier to compare
    for resource in result["resourceMetrics"]:
        sort_attributes(resource["resource"])
        for scope in resource["scopeMetrics"]:
            scope["metrics"] = sorted(
                scope["metrics"], key=lambda m: m["name"])
            for metric in scope["metrics"]:
                if "sum" in metric:
                    process_metric_type(metric["sum"])
                if "gauge" in metric:
                    process_metric_type(metric["gauge"])
                if "histogram" in metric:
                    process_metric_type(metric["histogram"])

                # Get rid of value of metric called "scrape_duration_seconds" as it is not stable
                if metric["name"] == "scrape_duration_seconds":
                    metric["gauge"]["dataPoints"][0]["asDouble"] = 0
    result["resourceMetrics"] = sorted(
        result["resourceMetrics"], key=resource_sorting_key)

    return result

# Function to run a shell command and print its output and errors
def run_shell_command(command):
    print(f"{command}")
    result = subprocess.run(command, shell=True, capture_output=True, text=True)
    print(result.stdout)
    print(result.stderr)
