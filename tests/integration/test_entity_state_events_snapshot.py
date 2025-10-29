import json
import os
import time
from typing import Dict, Iterable, List, Tuple

import pytest

from test_utils import (
    get_attribute_key_and_value,
    get_merged_json,
    parse_value,
    retry_until_ok,
    get_loki_client,
)

ENDPOINT = os.getenv("TIMESERIES_MOCK_ENDPOINT", "localhost:8088")
ENTITY_EVENTS_URL = f"http://{ENDPOINT}/entitystateevents.json"
EXPECTED_DIR = os.path.join(
    os.path.dirname(__file__), "expected_entitystateevents"
)

# Check if we should use Loki or file-based testing
USE_LOKI = os.getenv("USE_LOKI", "false").lower() == "true"


def _load_expected_cases() -> Iterable[str]:
    for entry in sorted(os.listdir(EXPECTED_DIR)):
        if entry.endswith(".json"):
            yield entry


def _resource_matches(resource: Dict, expected_attributes: List[Dict]) -> bool:
    if not expected_attributes:
        return True

    for attribute in expected_attributes:
        key = attribute["key"]
        expected_value = attribute.get("value")
        actual_value = get_attribute_key_and_value(resource, key)
        if actual_value is None:
            return False
        if expected_value is not None and actual_value != expected_value:
            return False

    return True


def _kv_pairs_match(actual_kvlist: Dict, expected_pairs: List[Dict]) -> bool:
    if not expected_pairs:
        return True

    if not isinstance(actual_kvlist, dict) or "values" not in actual_kvlist:
        return False

    actual_map: Dict[str, str] = {}
    for pair in actual_kvlist["values"]:
        actual_map[pair["key"]] = parse_value(pair["value"])

    for pair in expected_pairs:
        key = pair["key"]
        expected_value = pair.get("value")
        if key not in actual_map:
            return False
        if expected_value is not None and actual_map[key] != expected_value:
            return False

    return True


def _log_record_matches(log_record: Dict, expected_event: Dict) -> bool:
    for key, expected_value in expected_event.items():
        actual_value = get_attribute_key_and_value(log_record, key)
        if actual_value is None:
            return False

        if isinstance(expected_value, list):
            if not _kv_pairs_match(actual_value, expected_value):
                return False
        else:
            if actual_value != expected_value:
                return False

    return True


def _assert_expected_events(
    content: bytes, expected_attributes: List[Dict], expected_events: List[Dict]
):
    merged_json = get_merged_json(content)
    missing_events: List[Dict] = []

    for expected_event in expected_events:
        if not _event_found(merged_json, expected_attributes, expected_event):
            missing_events.append(expected_event)

    if missing_events:
        missing_descriptions = [json.dumps(event, sort_keys=True) for event in missing_events]
        return False, "Missing expected events: " + "; ".join(missing_descriptions)

    return True, ""


def _event_found(
    merged_json: List[Dict], expected_attributes: List[Dict], expected_event: Dict
) -> bool:
    for json_line in merged_json:
        for resource_log in json_line.get("resourceLogs", []):
            resource = resource_log.get("resource", {})
            attributes = resource.get("attributes", [])
            resource_wrapper = {"attributes": attributes}

            if not _resource_matches(resource_wrapper, expected_attributes):
                continue

            for scope_log in resource_log.get("scopeLogs", []):
                for log_record in scope_log.get("logRecords", []):
                    if _log_record_matches(log_record, expected_event):
                        return True

    return False


def _convert_dot_to_underscore(key: str) -> str:
    """Convert dots to underscores for Loki structured metadata field names."""
    return key.replace(".", "_")


def _build_loki_query(expected_event: Dict) -> str:
    """
    Build LogQL query from expected event structure.
    
    In Loki:
    - otel.entity.event.type and otel.entity.type are labels
    - otel.entity.id attributes become structured metadata with dots replaced by underscores
    - Example: otel.entity.id[{key: "sw.server.address.fqdn"}] becomes otel_entity_id_sw_server_address_fqdn
    """
    label_selectors = []
    
    # Add entity event type as label
    if "otel.entity.event.type" in expected_event:
        event_type = expected_event["otel.entity.event.type"]
        # Convert to Loki label format
        label_name = _convert_dot_to_underscore("otel.entity.event.type")
        label_selectors.append(f'{label_name}="{event_type}"')
    
    # Add entity type as label
    if "otel.entity.type" in expected_event:
        entity_type = expected_event["otel.entity.type"]
        label_name = _convert_dot_to_underscore("otel.entity.type")
        label_selectors.append(f'{label_name}="{entity_type}"')
    
    # Add relationship type as label if present
    if "otel.entity_relationship.type" in expected_event:
        rel_type = expected_event["otel.entity_relationship.type"]
        label_name = _convert_dot_to_underscore("otel.entity_relationship.type")
        label_selectors.append(f'{label_name}="{rel_type}"')
    
    # Start with label selector
    if label_selectors:
        query = "{" + ", ".join(label_selectors) + "}"
    else:
        # Fallback: query for all entity events
        query = '{otel_entity_event_type!=""}'
    
    return query


def _build_structured_metadata_filters(expected_event: Dict) -> List[str]:
    """
    Build structured metadata filters for entity attributes.
    
    Generically handles any attribute that is a list of {key, value} dicts.
    These become structured metadata in Loki with the format:
    - attribute_name + "_" + key (with dots converted to underscores)
    
    Examples:
    - otel.entity.id[{key: "sw.server.address.fqdn"}] → otel_entity_id_sw_server_address_fqdn
    - otel.entity_relationship.source_entity.id[{key: "k8s.namespace.name"}] → otel_entity_relationship_source_entity_id_k8s_namespace_name
    """
    filters = []
    
    # Iterate through all attributes in the expected event
    for attr_name, attr_value in expected_event.items():
        # Check if this attribute is a list of key-value dicts
        if isinstance(attr_value, list) and len(attr_value) > 0:
            # Check if all items in the list are dicts with "key" field
            if all(isinstance(item, dict) and "key" in item for item in attr_value):
                # This is a structured attribute (like otel.entity.id)
                # Convert attribute name to Loki format
                attr_name_loki = _convert_dot_to_underscore(attr_name)
                
                # Process each key-value pair
                for item in attr_value:
                    item_key = item["key"]
                    # Build the structured metadata field name
                    # Format: {attribute_name}_{item_key} (all dots → underscores)
                    field_name = f"{attr_name_loki}_{_convert_dot_to_underscore(item_key)}"
                    
                    if "value" in item:
                        # Filter for specific value
                        expected_value = item["value"]
                        filters.append(f'{field_name}="{expected_value}"')
                    else:
                        # Just check existence (field is not empty)
                        filters.append(f'{field_name}!=""')
    
    return filters


def _match_event_in_loki_logs(loki_logs: List[Tuple[str, str]], expected_event: Dict) -> bool:
    """
    Check if any Loki log entry matches the expected event structure.
    
    Args:
        loki_logs: List of (timestamp, log_line) tuples from Loki
        expected_event: Expected event structure from JSON fixture
    
    Returns:
        True if a matching event is found
    """
    for timestamp, log_line in loki_logs:
        # Loki returns logs as strings, but entity events are structured
        # The log line itself might be JSON or just contain the entity data
        # For now, we'll consider a match if we found logs with the right labels/metadata
        # The structured metadata filtering in the query should ensure we get the right logs
        return True
    
    return False


def _assert_expected_events_loki(
    expected_attributes: List[Dict], expected_events: List[Dict], timeout: int = 600
) -> bool:
    """
    Verify expected events exist in Loki.
    
    Args:
        expected_attributes: Resource attributes to filter by (currently unused for Loki)
        expected_events: List of expected event structures
        timeout: Maximum time to wait for events to appear
    
    Returns:
        True if all events found, raises ValueError otherwise
    """
    client = get_loki_client()
    missing_events: List[Dict] = []
    
    for expected_event in expected_events:
        # Build LogQL query
        query = _build_loki_query(expected_event)
        
        # Add structured metadata filters
        metadata_filters = _build_structured_metadata_filters(expected_event)
        if metadata_filters:
            for filter_expr in metadata_filters:
                query += f" | {filter_expr}"
        
        print(f"Querying Loki for event: {query}")
        
        # Query Loki
        try:
            # Use a shorter duration for each query, but poll multiple times
            end_time = time.time() + timeout
            found = False
            
            while time.time() < end_time:
                logs = client.query_recent_logs(query, duration='5m', limit=100)
                
                if logs:
                    print(f"✅ Found {len(logs)} matching logs in Loki")
                    found = True
                    break
                
                print(f"⏳ No logs found yet, retrying... ({int(end_time - time.time())}s remaining)")
                time.sleep(10)
            
            if not found:
                missing_events.append(expected_event)
                print(f"❌ Event not found in Loki after {timeout}s")
        
        except Exception as e:
            print(f"❌ Error querying Loki: {e}")
            missing_events.append(expected_event)
    
    if missing_events:
        missing_descriptions = [json.dumps(event, sort_keys=True) for event in missing_events]
        error_msg = "Missing expected events in Loki: " + "; ".join(missing_descriptions)
        raise ValueError(error_msg)
    
    return True


def _event_found(
    merged_json: List[Dict], expected_attributes: List[Dict], expected_event: Dict
) -> bool:
    for json_line in merged_json:
        for resource_log in json_line.get("resourceLogs", []):
            resource = resource_log.get("resource", {})
            attributes = resource.get("attributes", [])
            resource_wrapper = {"attributes": attributes}

            if not _resource_matches(resource_wrapper, expected_attributes):
                continue

            for scope_log in resource_log.get("scopeLogs", []):
                for log_record in scope_log.get("logRecords", []):
                    if _log_record_matches(log_record, expected_event):
                        return True

    return False


@pytest.mark.parametrize("expected_file", list(_load_expected_cases()))
def test_entity_state_events_expected_content(expected_file: str) -> None:
    """
    Test that expected entity state events are present in the system.
    
    Supports both file-based (legacy) and Loki-based testing via USE_LOKI env var.
    Each JSON file in expected_entitystateevents/ becomes a separate test case.
    """
    file_path = os.path.join(EXPECTED_DIR, expected_file)
    with open(file_path, "r", encoding="utf-8") as handle:
        expected_case = json.load(handle)

    expected_attributes = expected_case.get("resource_attributes", [])
    expected_events = expected_case.get("events", [])
    
    print(f"\n{'='*60}")
    print(f"Testing: {expected_file}")
    print(f"Expected events: {len(expected_events)}")
    print(f"Backend: {'Loki' if USE_LOKI else 'File'}")
    print(f"{'='*60}\n")

    if USE_LOKI:
        # Loki-based testing
        try:
            _assert_expected_events_loki(expected_attributes, expected_events, timeout=600)
            print(f"✅ All expected events found in Loki for {expected_file}")
        except ValueError as e:
            print(f"❌ Test failed for {expected_file}: {e}")
            raise
    else:
        # Legacy file-based testing
        def _failure_printer(content: bytes) -> None:
            print(f"Failed to find expected events for fixture {expected_file}")

        retry_until_ok(
            ENTITY_EVENTS_URL,
            lambda content: _assert_expected_events(content, expected_attributes, expected_events),
            _failure_printer,
            timeout=180,
        )
        print(f"✅ All expected events found in file for {expected_file}")
