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
    result = [records["body"]["stringValue"] if "stringValue" in records["body"] else records["body"]
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
    timeout = 240  # set the timeout in seconds
    start_time = time.time()
    last_exception = None
    last_error = ''
    while time.time() - start_time < timeout:
        is_ok = False
        response = None
        try:
            try: 
                response = requests.get(url)
                response.raise_for_status()
            except requests.exceptions.RequestException as e:
                print(f"An error occurred while making the request: {e}")
        except Exception as e:
            last_exception = e
            print(e, traceback.format_exc())

        if response is not None and response.status_code == 200:
            if( last_error == ''): 
                print("Successfully downloaded!")
            result = func(response.content)
            if( type(result) != tuple):
                is_ok = result
            else:
                is_ok = result[0]
                if( last_error != result[1]):
                    last_error = result[1]
                    print(last_error)            
        else:
            if response is not None:
                print('Failed to download otel messages. Response code:',
                    response.status_code)

            print('Failed to download otel messages')
        
        if is_ok:
            print(f'Succesfully passed assert')
            return True
        else:
            print('Retrying...')
            time.sleep(10)

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
    
def datapoint_value(datapoint):    
    if "asDouble" in datapoint:
        return datapoint["asDouble"]
    elif "asInt" in datapoint:
        return datapoint["asInt"]
    elif "asString" in datapoint:
        return datapoint["asString"]
    else:
        raise Exception('Unknown data point value')

def remove_time_in_datapoint(datapoint):
    if "timeUnixNano" in datapoint:
        datapoint["timeUnixNano"] = "0"
    if "startTimeUnixNano" in datapoint:
        datapoint["startTimeUnixNano"] = "0"


def sort_attributes(element):
    if "attributes" in element:
        element["attributes"] = sorted(
            element["attributes"], key=lambda a: a["key"])

def sanitize_attributes(element):
    if "attributes" in element:
        # Exclude node attributes as they comes from real cluster so it is not predictable in integrations test to match them
        exclude_prefixes = ['k8s.node.annotations', 'k8s.node.labels']

        filtered_attributes = [
            attr for attr in element["attributes"]
            if all(not attr["key"].startswith(prefix) for prefix in exclude_prefixes)
        ]

        # Sort the remaining attributes
        element["attributes"] = sorted(
            filtered_attributes, key=lambda a: a["key"]
        )

        # 'k8s.node.name' also comes from real cluster so sanitize it to `test-node` to make it predictable
        for attr in element["attributes"]:
            if attr["key"] == "k8s.node.name":
                attr["value"] = {"stringValue": "test-node"}

def sort_datapoints(metric):
    metric["dataPoints"] = sorted(
        metric["dataPoints"], key=datapoint_sorting_key)


def process_metric_type(metric):
    if "dataPoints" in metric:
        for dp in metric["dataPoints"]:
            remove_time_in_datapoint(dp)
            sanitize_attributes(dp)
            sort_attributes(dp)
        sort_datapoints(metric)

def merge_datapoints(existing_datapoints, new_datapoints):
    merged_datapoints = [(datapoint_sorting_key(dp), dp) for dp in existing_datapoints]

    for new_datapoint in new_datapoints:
        new_datapoint_hash_key = datapoint_sorting_key(new_datapoint)
        found = False  # flag to track if a matching datapoint was found
        
        for key, existing_datapoint in merged_datapoints:
            if key == new_datapoint_hash_key:
                existing_datapoint.update(new_datapoint)  # update existing_datapoint in-place
                found = True  # set flag to True since a matching datapoint was found
                break  # exit the loop since a match was found and handled
        
        if not found:  # if no matching datapoint was found, append the new datapoint
            merged_datapoints.append((new_datapoint_hash_key, new_datapoint))

    existing_datapoints.clear()
    existing_datapoints.extend(dp for _, dp in merged_datapoints)
    
def merge_metrics(existing_metric, new_metric):
    metric_types = ["sum", "gauge", "histogram"]

    for metric_type in metric_types:
        if metric_type in existing_metric and metric_type in new_metric:
            existing_datapoints = existing_metric[metric_type]["dataPoints"]
            new_datapoints = new_metric[metric_type]["dataPoints"]
            merge_datapoints(existing_datapoints, new_datapoints)

def merge_scope_metrics(existing_scope, new_scope):
    for new_metric in new_scope["metrics"]:
        new_metric_name = new_metric["name"]
        for existing_metric in existing_scope["metrics"]:
            if existing_metric["name"] == new_metric_name:
                merge_metrics(existing_metric, new_metric)
                break
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
    new_resources = [(resource_sorting_key(resource), resource) for resource in new_json["resourceMetrics"]]

    for existing_resource in result["resourceMetrics"]:
        existing_key = resource_sorting_key(existing_resource)
        matching_new_resources = [item for item in new_resources if item[0] == existing_key]
        for _, new_resource in matching_new_resources:
            merge_resources(existing_resource, new_resource)
            new_resources.remove((existing_key, new_resource))

    result["resourceMetrics"].extend(resource for _, resource in new_resources)

def get_merged_json(content):
    result = {"resourceMetrics": []}
    for line in content.splitlines()[-10:]: # only process the last 10 json lines
        custom_json_merge(result, json.loads(line))

    # Sort the result and set timeStamps to 0 to make it easier to compare
    for resource in result["resourceMetrics"]:
        sanitize_attributes(resource["resource"])
        sort_attributes(resource["resource"])
        for scope in resource["scopeMetrics"]:
            scope["scope"] = {}
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
