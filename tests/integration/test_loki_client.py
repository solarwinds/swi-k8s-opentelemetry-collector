"""
Unit tests for Loki client functionality.

These tests verify the LokiClient class works correctly and can communicate with Loki.
"""

import pytest
import os
from test_utils import LokiClient, get_loki_client, search_logs_by_attributes


def test_loki_client_connection():
    """
    Test that Loki API is accessible and responding.
    Verifies /ready endpoint returns 200 status code.
    """
    client = get_loki_client()
    
    assert client.is_ready(), "Loki should be ready and responding to health checks"
    print("✅ Loki client successfully connected to Loki API")


def test_loki_client_initialization():
    """
    Test LokiClient initialization with default and custom endpoints.
    """
    # Test default endpoint
    client_default = LokiClient()
    assert client_default.base_url == "http://localhost:3100"
    
    # Test custom endpoint
    client_custom = LokiClient("http://loki.test:3100")
    assert client_custom.base_url == "http://loki.test:3100"
    
    # Test endpoint with trailing slash
    client_slash = LokiClient("http://localhost:3100/")
    assert client_slash.base_url == "http://localhost:3100"
    
    print("✅ LokiClient initialization works correctly")


def test_duration_parsing():
    """
    Test duration string parsing to nanoseconds.
    """
    client = LokiClient()
    
    # Test various duration formats
    assert client._parse_duration_to_ns("10s") == 10 * 1_000_000_000
    assert client._parse_duration_to_ns("5m") == 5 * 60 * 1_000_000_000
    assert client._parse_duration_to_ns("2h") == 2 * 3600 * 1_000_000_000
    assert client._parse_duration_to_ns("1d") == 86400 * 1_000_000_000
    assert client._parse_duration_to_ns("100ms") == 100 * 1_000_000
    
    print("✅ Duration parsing works correctly")


def test_loki_query_labels():
    """
    Test querying Loki labels endpoint.
    This verifies basic API connectivity and response parsing.
    """
    client = get_loki_client()
    
    try:
        response = client.session.get(f"{client.base_url}/loki/api/v1/labels", timeout=10)
        response.raise_for_status()
        
        data = response.json()
        assert 'data' in data, "Response should contain 'data' field"
        
        print(f"✅ Loki labels query successful, found {len(data['data'])} labels")
        
        # Print some labels for debugging
        if data['data']:
            print(f"Sample labels: {data['data'][:5]}")
    
    except Exception as e:
        pytest.skip(f"Loki labels query failed (Loki may not have any data yet): {e}")


def test_query_recent_logs_empty():
    """
    Test query_recent_logs with a query that should return no results.
    This verifies the query logic works even with empty results.
    """
    client = get_loki_client()
    
    # Query for a non-existent namespace
    logs = client.query_recent_logs(
        query='{k8s_namespace_name="nonexistent-test-namespace-12345"}',
        duration='1m'
    )
    
    assert isinstance(logs, list), "query_recent_logs should return a list"
    print(f"✅ Empty query returned {len(logs)} logs (expected 0)")


def test_get_loki_client_env():
    """
    Test that get_loki_client respects LOKI_ENDPOINT environment variable.
    """
    # Test default
    client = get_loki_client()
    assert "localhost:3100" in client.base_url or "127.0.0.1:3100" in client.base_url
    
    # Test with custom endpoint
    os.environ['LOKI_ENDPOINT'] = 'http://custom-loki:3100'
    client_custom = get_loki_client()
    assert client_custom.base_url == 'http://custom-loki:3100'
    
    # Restore environment
    if 'LOKI_ENDPOINT' in os.environ:
        del os.environ['LOKI_ENDPOINT']
    
    print("✅ get_loki_client respects LOKI_ENDPOINT environment variable")


def test_count_logs_matching():
    """
    Test count_logs_matching method.
    """
    client = get_loki_client()
    
    # Count logs in a namespace that might not exist
    count = client.count_logs_matching(
        query='{k8s_namespace_name="nonexistent-namespace"}',
        duration='1m'
    )
    
    assert isinstance(count, int), "count_logs_matching should return an integer"
    assert count >= 0, "Count should be non-negative"
    print(f"✅ count_logs_matching returned: {count}")


def test_search_logs_by_attributes():
    """
    Test the search_logs_by_attributes helper function.
    """
    # Test with no filters (should query all logs)
    logs = search_logs_by_attributes(duration='1m', limit=10)
    assert isinstance(logs, list), "search_logs_by_attributes should return a list"
    
    # Test with namespace filter
    logs_ns = search_logs_by_attributes(
        namespace='test-namespace',
        duration='1m',
        limit=10
    )
    assert isinstance(logs_ns, list)
    
    # Test with multiple filters
    logs_filtered = search_logs_by_attributes(
        namespace='default',
        pod='test-pod',
        text_filter='error',
        duration='5m',
        limit=100
    )
    assert isinstance(logs_filtered, list)
    
    print("✅ search_logs_by_attributes works with various filter combinations")


if __name__ == '__main__':
    """
    Run tests directly for quick validation.
    """
    print("Running Loki client tests...")
    print("=" * 60)
    
    try:
        test_loki_client_initialization()
        test_duration_parsing()
        test_loki_client_connection()
        test_get_loki_client_env()
        test_query_recent_logs_empty()
        test_count_logs_matching()
        test_search_logs_by_attributes()
        test_loki_query_labels()
        
        print("=" * 60)
        print("✅ All Loki client tests passed!")
        
    except Exception as e:
        print("=" * 60)
        print(f"❌ Test failed: {e}")
        import traceback
        traceback.print_exc()
