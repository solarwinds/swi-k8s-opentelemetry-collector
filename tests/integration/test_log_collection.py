import pytest
import os
import json
from test_utils import (
    get_all_bodies_for_all_sent_content, 
    retry_until_ok, 
    run_shell_command,
    get_loki_client
)

endpoint = os.getenv("TIMESERIES_MOCK_ENDPOINT", "localhost:8088")
url = f'http://{endpoint}/logs.json'
pod_name = 'dummy-logging-pod'
tested_log = 'testlog-swo-k8s-collector-integration-test'

# Check if we should use Loki or file-based testing
USE_LOKI = os.getenv("USE_LOKI", "false").lower() == "true"

def setup_function():
    run_shell_command(f'kubectl run {pod_name} --image bash:alpine3.19 -n default -- -ec "while :; do echo \'{tested_log}\'; sleep 5 ; done"')

def teardown_function():
    run_shell_command(f'kubectl delete pod {pod_name} -n default')

def test_logs_generated():
    """
    Main test function that delegates to file-based or Loki-based assertion.
    This enables gradual migration from file storage to Loki.
    """
    if USE_LOKI:
        print("Using Loki-based log collection test")
        test_logs_generated_loki()
    else:
        print("Using file-based log collection test")
        test_logs_generated_file()

def test_logs_generated_file():
    """
    Legacy file-based log collection test.
    Queries logs from nginx-served JSON file.
    """
    retry_until_ok(url, assert_test_log_found, print_failure)

def test_logs_generated_loki():
    """
    Loki-based log collection test.
    Queries logs directly from Loki using LogQL.
    """
    client = get_loki_client()
    
    # Build LogQL query to find logs from our test pod
    query = f'{{k8s_namespace_name="default", k8s_pod_name="{pod_name}"}} |= "{tested_log}"'
    
    print(f"Querying Loki with: {query}")
    
    # Wait for the log to appear in Loki
    try:
        client.wait_for_log_match(
            query=query,
            expected_pattern=tested_log,
            timeout=600,  # 10 minutes timeout
            poll_interval=10
        )
        print(f"✅ Successfully found test log in Loki: {tested_log}")
    except ValueError as e:
        # Print debug information on failure
        print(f"❌ Failed to find log in Loki: {e}")
        print(f"Query used: {query}")
        
        # Try querying for any logs from the pod to help debug
        debug_query = f'{{k8s_namespace_name="default", k8s_pod_name="{pod_name}"}}'
        debug_logs = client.query_recent_logs(debug_query, duration='5m', limit=10)
        
        print(f"\nDebug: Found {len(debug_logs)} logs from pod {pod_name} in last 5 minutes:")
        for ts, log in debug_logs[:5]:
            print(f"  {log}")
        
        raise

def assert_test_log_found(content):
    raw_bodies = get_all_bodies_for_all_sent_content(content)
    test_log_found = any(
        f'{tested_log}' in entry
        for body in raw_bodies
        for entry in body
    )
    return test_log_found

def print_failure(content):
    raw_bodies = get_all_bodies_for_all_sent_content(content)
    print(f'Failed to find {tested_log}')
    print('All logs in raw_bodies_dump.txt')
    #print(raw_bodies)
    # Dump the raw_bodies to a file
    with open('raw_bodies_dump.txt', 'w') as file:
        # Convert the list to a JSON string for better formatting
        json.dump(raw_bodies, file, indent=4)


