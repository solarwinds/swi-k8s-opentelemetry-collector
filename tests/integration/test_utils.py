import json
import time
import requests
import traceback
import subprocess
import re
import os
from typing import List, Dict, Optional, Tuple

class LokiClient:
    """
    Client for querying Grafana Loki API.
    Supports LogQL queries and provides utilities for integration testing.
    """
    
    def __init__(self, base_url: str = "http://localhost:3100"):
        """
        Initialize Loki client.
        
        Args:
            base_url: Base URL for Loki API (default: http://localhost:3100)
        """
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
        self.session.headers.update({'Content-Type': 'application/json'})
    
    def query_range(self, query: str, start: int, end: int, limit: int = 1000, direction: str = "backward") -> Dict:
        """
        Query Loki for logs within a time range.
        
        Args:
            query: LogQL query string (e.g., '{namespace="default"}')
            start: Start timestamp in nanoseconds
            end: End timestamp in nanoseconds
            limit: Maximum number of log lines to return
            direction: Query direction ("forward" or "backward")
        
        Returns:
            Parsed JSON response from Loki API
        
        Raises:
            requests.exceptions.RequestException: On API errors
        """
        url = f"{self.base_url}/loki/api/v1/query_range"
        
        params = {
            'query': query,
            'start': start,
            'end': end,
            'limit': limit,
            'direction': direction
        }
        
        try:
            response = self.session.get(url, params=params, timeout=30)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.Timeout:
            raise ValueError(f"Loki query timed out after 30s: {query}")
        except requests.exceptions.RequestException as e:
            raise ValueError(f"Loki query failed: {e}")
    
    def query_recent_logs(self, query: str, duration: str = '10m', limit: int = 1000) -> List[Tuple[int, str]]:
        """
        Query logs from recent time window.
        
        Args:
            query: LogQL query string
            duration: Time window duration (e.g., '10m', '1h', '30s')
            limit: Maximum number of log lines to return
        
        Returns:
            List of (timestamp_ns, log_line) tuples
        """
        end_ns = int(time.time() * 1e9)
        start_ns = end_ns - self._parse_duration_to_ns(duration)
        
        response = self.query_range(query, start_ns, end_ns, limit)
        
        # Extract log lines from response
        logs = []
        if 'data' in response and 'result' in response['data']:
            for stream in response['data']['result']:
                for values in stream.get('values', []):
                    timestamp_ns = int(values[0])
                    log_line = values[1]
                    logs.append((timestamp_ns, log_line))
        
        return logs
    
    def wait_for_log_match(self, query: str, expected_pattern: str, timeout: int = 600, poll_interval: int = 10) -> bool:
        """
        Poll Loki until a log matching the pattern appears or timeout occurs.
        
        Args:
            query: LogQL query string
            expected_pattern: Regex pattern to match against log lines
            timeout: Maximum wait time in seconds
            poll_interval: Time between polls in seconds
        
        Returns:
            True if matching log found
        
        Raises:
            ValueError: If timeout occurs before match is found
        """
        start_time = time.time()
        pattern = re.compile(expected_pattern)
        
        print(f"Waiting for log matching pattern: {expected_pattern}")
        print(f"Query: {query}")
        
        while time.time() - start_time < timeout:
            elapsed = int(time.time() - start_time)
            
            try:
                logs = self.query_recent_logs(query, duration='1m')
                
                for timestamp_ns, log_line in logs:
                    if pattern.search(log_line):
                        print(f"✅ Found matching log after {elapsed}s")
                        print(f"Log line: {log_line[:200]}")
                        return True
                
                print(f"No match yet... ({elapsed}s / {timeout}s) - Found {len(logs)} logs")
                
            except Exception as e:
                print(f"Query error (will retry): {e}")
            
            time.sleep(poll_interval)
        
        raise ValueError(f"Timed out after {timeout}s waiting for log matching: {expected_pattern}")
    
    def count_logs_matching(self, query: str, duration: str = '10m') -> int:
        """
        Count number of log lines matching a query.
        
        Args:
            query: LogQL query string
            duration: Time window duration
        
        Returns:
            Number of matching log lines
        """
        logs = self.query_recent_logs(query, duration)
        return len(logs)
    
    def is_ready(self) -> bool:
        """
        Check if Loki is ready to accept queries.
        
        Returns:
            True if Loki is ready, False otherwise
        """
        try:
            response = self.session.get(f"{self.base_url}/ready", timeout=5)
            return response.status_code == 200
        except:
            return False
    
    def _parse_duration_to_ns(self, duration: str) -> int:
        """
        Parse duration string to nanoseconds.
        
        Args:
            duration: Duration string (e.g., '10m', '1h', '30s')
        
        Returns:
            Duration in nanoseconds
        """
        duration = duration.strip()
        
        if duration.endswith('ns'):
            return int(duration[:-2])
        elif duration.endswith('us'):
            return int(duration[:-2]) * 1_000
        elif duration.endswith('ms'):
            return int(duration[:-2]) * 1_000_000
        elif duration.endswith('s'):
            return int(duration[:-1]) * 1_000_000_000
        elif duration.endswith('m'):
            return int(duration[:-1]) * 60 * 1_000_000_000
        elif duration.endswith('h'):
            return int(duration[:-1]) * 3600 * 1_000_000_000
        elif duration.endswith('d'):
            return int(duration[:-1]) * 86400 * 1_000_000_000
        else:
            raise ValueError(f"Invalid duration format: {duration}")


def get_loki_client() -> LokiClient:
    """
    Get configured LokiClient instance.
    
    Reads endpoint from LOKI_ENDPOINT environment variable.
    Default: http://localhost:3100
    
    Returns:
        Configured LokiClient instance
    """
    endpoint = os.environ.get('LOKI_ENDPOINT', 'http://localhost:3100')
    return LokiClient(base_url=endpoint)


def search_logs_by_attributes(
    namespace: Optional[str] = None,
    pod: Optional[str] = None,
    container: Optional[str] = None,
    log_type: Optional[str] = None,
    text_filter: Optional[str] = None,
    duration: str = '10m',
    limit: int = 100
) -> List[Tuple[int, str]]:
    """
    Search logs using common attributes with flexible filtering.
    
    Args:
        namespace: Kubernetes namespace to filter by
        pod: Pod name to filter by
        container: Container name to filter by
        log_type: Log type to filter by (from sw.k8s.log.type attribute)
        text_filter: Text to search for in log lines
        duration: Time window to query
        limit: Maximum number of results
    
    Returns:
        List of (timestamp_ns, log_line) tuples
    """
    # Build LogQL query with label matchers
    label_selectors = []
    
    if namespace:
        label_selectors.append(f'k8s_namespace_name="{namespace}"')
    if pod:
        label_selectors.append(f'k8s_pod_name="{pod}"')
    if container:
        label_selectors.append(f'k8s_container_name="{container}"')
    
    # Construct base query
    if label_selectors:
        query = '{' + ', '.join(label_selectors) + '}'
    else:
        query = '{}'
    
    # Add text filter if provided
    if text_filter:
        query += f' |= "{text_filter}"'
    
    # Note: log_type is structured metadata, not a label, so it's filtered differently
    # It would need to be added via structured metadata filter: | sw_k8s_log_type="value"
    if log_type:
        query += f' | sw_k8s_log_type="{log_type}"'
    
    client = get_loki_client()
    return client.query_recent_logs(query, duration, limit)



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
        value = parse_value(attribute.get("value", {}))
        if key == target_key and value == expected_value:
            print(f"Resource has attribute with key '{target_key}' and value '{expected_value}'.")
            return True

    print(f"Resource does not have attribute with key '{target_key}' and value '{expected_value}'.")
    return False


def get_attribute_key_and_value(resource, target_key):
    attributes = resource.get("attributes", [])
    for attribute in attributes:
        key = attribute.get("key", "")
        if (key == target_key):
            return parse_value(attribute.get("value", {}))
    return None


def get_attributes_of_kvmap(resource, target_key):
    kvmap = get_attribute_key_and_value(resource, target_key)['values']
    result = dict()
    for pair in kvmap:
        key = pair['key']
        value = parse_value(pair['value'])
        result[key] = value
    return result


def parse_value(value):
    val_str = value.get('stringValue', None)
    if (val_str != None):
        return val_str
    val_bool = value.get('boolValue', None)
    if (val_bool != None):
        return val_bool
    val_kvmap = value.get('kvlistValue', None)
    return val_kvmap

