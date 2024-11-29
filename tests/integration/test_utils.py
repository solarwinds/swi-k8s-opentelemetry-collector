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


def retry_until_ok(url, func, print_failure, timeout = 600):
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
            if( last_error != ''): 
                print(last_error)
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
    
def datapoint_value(datapoint):    
    if "asDouble" in datapoint:
        return datapoint["asDouble"]
    elif "asInt" in datapoint:
        return datapoint["asInt"]
    elif "asString" in datapoint:
        return datapoint["asString"]
    else:
        raise Exception('Unknown data point value')

def get_merged_json(content):
    result = []
    for line in content.splitlines():
        result.append(json.loads(line))

    return result

# Function to run a shell command and print its output and errors
def run_shell_command(command):
    print(f"{command}")
    result = subprocess.run(command, shell=True, capture_output=True, text=True)
    print(result.stdout)
    print(result.stderr)

def has_attribute_with_key_and_value(resource, target_key, expected_value):
    attributes = resource.get("attributes", [])
    for attribute in attributes:
        key = attribute.get("key", "")
        value = attribute.get("value", {}).get("stringValue", "")
        if key == target_key and value == expected_value:
            print(f"Resource has attribute with key '{target_key}' and value '{expected_value}'.")
            return True

    print(f"Resource does not have attribute with key '{target_key}' and value '{expected_value}'.")
    return False