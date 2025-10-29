import pytest
import os
from test_utils import (
    get_all_bodies_for_all_sent_content, 
    retry_until_ok, 
    run_shell_command,
    get_loki_client
)

endpoint = os.getenv("TIMESERIES_MOCK_ENDPOINT", "localhost:8088")
url = f'http://{endpoint}/events.json'
pod_name = 'dummy-pod'
expected_event = f'Started container {pod_name}'

# Check if we should use Loki or file-based testing
USE_LOKI = os.getenv("USE_LOKI", "false").lower() == "true"

def setup_function():
    run_shell_command(f"kubectl run {pod_name} --labels \"test-label=test-value\" --overrides=\"{{ \\\"apiVersion\\\": \\\"v1\\\", \\\"metadata\\\": {{\\\"annotations\\\": {{ \\\"test-annotation\\\":\\\"test-value\\\" }} }} }}\" --image bash:alpine3.19 -n default -- -ec \"while :; do sleep 5 ; done\"")

def teardown_function():
    run_shell_command(f'kubectl delete pod {pod_name} -n default')

def test_events_generated():
    """
    Main test function that delegates to file-based or Loki-based assertion.
    This enables gradual migration from file storage to Loki.
    """
    if USE_LOKI:
        print("Using Loki-based event collection test")
        test_events_generated_loki()
    else:
        print("Using file-based event collection test")
        test_events_generated_file()

def test_events_generated_file():
    """
    Legacy file-based event collection test.
    Queries events from nginx-served JSON file.
    """
    retry_until_ok(url, assert_test_event_found, print_failure)

def test_events_generated_loki():
    """
    Loki-based event collection test.
    Queries events directly from Loki using LogQL.
    
    Note: Events are stored as logs with sw.k8s.log.type="event" attribute.
    This attribute is stored as structured metadata, not as an index label.
    """
    client = get_loki_client()
    
    # Build LogQL query to find events from our test pod
    # Events are identified by namespace and contain pod-related information
    # The sw.k8s.log.type attribute is structured metadata, so we filter on it
    query = f'{{k8s_namespace_name="default"}} | sw_k8s_log_type="event" |= "{pod_name}"'
    
    print(f"Querying Loki for events with: {query}")
    
    # Wait for the event to appear in Loki
    try:
        client.wait_for_log_match(
            query=query,
            expected_pattern=expected_event,
            timeout=600,  # 10 minutes timeout (events might take longer to appear)
            poll_interval=10
        )
        print(f"✅ Successfully found event in Loki: {expected_event}")
    except ValueError as e:
        # Print debug information on failure
        print(f"❌ Failed to find event in Loki: {e}")
        print(f"Query used: {query}")
        
        # Try querying for any events from the namespace to help debug
        debug_query = f'{{k8s_namespace_name="default"}} | sw_k8s_log_type="event"'
        debug_events = client.query_recent_logs(debug_query, duration='5m', limit=10)
        
        print(f"\nDebug: Found {len(debug_events)} events in namespace 'default' in last 5 minutes:")
        for ts, event in debug_events[:5]:
            print(f"  {event[:200]}")
        
        # Also try without the log type filter
        debug_query_all = f'{{k8s_namespace_name="default"}} |= "{pod_name}"'
        debug_all = client.query_recent_logs(debug_query_all, duration='5m', limit=10)
        
        print(f"\nDebug: Found {len(debug_all)} logs mentioning '{pod_name}' in last 5 minutes:")
        for ts, log in debug_all[:5]:
            print(f"  {log[:200]}")
        
        raise

def assert_test_event_found(content):
    raw_bodies = get_all_bodies_for_all_sent_content(content)
    test_event_found = any(expected_event in body for body in raw_bodies)
    return test_event_found

def print_failure(content):
    raw_bodies = get_all_bodies_for_all_sent_content(content)
    print(f'Failed to find "{expected_event}"')
    print('Sent events:')
    print(raw_bodies)

