import json
import time
import requests
import traceback
from jsonmerge import merge
import re

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
    
def resource_sorting_key(metric):
    return "".join([f"{a['key']}={a['value']['stringValue']}" for a in metric["resource"]["attributes"]])


def remove_time_in_datapoint(datapoint):
    if "timeUnixNano" in datapoint:
        datapoint["timeUnixNano"] = "0"
    if "startTimeUnixNano" in datapoint:
        datapoint["startTimeUnixNano"] = "0"


def sort_attributes(element):
    if "attributes" in element:
        element["attributes"] = sorted(
            element["attributes"], key=lambda a: a["key"])

def replace_uid_attributes(element):
    if "attributes" in element:
        for attribute in element["attributes"]:
            if re.match(r"^.*(deployment|statefulset|replicaset|daemonset|job|cronjob|node)\.uid$", attribute["key"]):
                attribute["value"]["stringValue"] = "00000000-0000-0000-0000-000000000000"

def sort_datapoints(metric):
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

    metric["dataPoints"] = sorted(
        metric["dataPoints"], key=datapoint_sorting_key)


def process_metric_type(metric):
    if "dataPoints" in metric:
        for dp in metric["dataPoints"]:
            remove_time_in_datapoint(dp)
            sort_attributes(dp)
        sort_datapoints(metric)


def get_merged_json(content):
    result = {}
    for line in content.splitlines():
        result = merge(result, json.loads(line))

    # Sort the result and set timeStamps to 0 to make it easier to compare
    for resource in result["resourceMetrics"]:
        replace_uid_attributes(resource["resource"])
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

