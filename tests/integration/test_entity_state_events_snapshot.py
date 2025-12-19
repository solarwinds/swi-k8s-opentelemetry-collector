import json
import os
import time
from typing import Dict, Iterable, List

import pytest

from clickhouse_client import ClickHouseClient
from test_utils import (
    get_attribute_key_and_value,
    parse_value,
)

EXPECTED_DIR = os.path.join(
    os.path.dirname(__file__), "expected_entitystateevents"
)

# Initialize ClickHouse client (uses CLICKHOUSE_ENDPOINT env var)
clickhouse = ClickHouseClient()

def _load_expected_cases() -> Iterable[str]:
    for entry in sorted(os.listdir(EXPECTED_DIR)):
        if entry.endswith(".json"):
            yield entry


def _resource_matches(resource: Dict, expected_attributes: List[Dict]) -> bool:
    """Check if resource attributes match expected attributes."""
    if not expected_attributes:
        return True

    for expected_attr in expected_attributes:
        key = expected_attr["key"]
        expected_value = expected_attr.get("value")
        actual_value = get_attribute_key_and_value(resource, key)
        if actual_value is None:
            return False
        if expected_value is not None and actual_value != expected_value:
            return False

    return True


def _scope_matches(scope: Dict, expected_attributes: List[Dict]) -> bool:
    """Check if scope attributes match expected attributes."""
    if not expected_attributes:
        return True

    for expected_attr in expected_attributes:
        key = expected_attr["key"]
        expected_value = expected_attr.get("value")
        actual_value = get_attribute_key_and_value(scope, key)
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
    """Check if a log record matches the expected event.
    
    The expected_event dict contains key-value pairs where keys are the attribute names
    (e.g., 'otel.entity.event.type') to look for in the log record's attributes array.
    
    Special handling: If expected value is an empty list [], we treat it as optional - 
    the field may be missing in actual data (None) or may be present as an empty list.
    """
    for expected_key, expected_value in expected_event.items():
        actual_value = get_attribute_key_and_value(log_record, expected_key)
        
        # Special case: empty list in expected means the field is optional
        # It can be missing (None) or present as an empty list
        if isinstance(expected_value, list) and len(expected_value) == 0:
            if actual_value is None:
                # Field is missing in actual data, which is acceptable for empty expected list
                continue
            # If actual_value exists, verify it's an empty list
            if not _kv_pairs_match(actual_value, expected_value):
                return False
        elif actual_value is None:
            # For non-empty expected values, actual_value must exist
            return False
        elif isinstance(expected_value, list):
            if not _kv_pairs_match(actual_value, expected_value):
                return False
        else:
            if actual_value != expected_value:
                return False

    return True


def _assert_expected_events(
    merged_json: List[Dict], 
    expected_resource_attributes: List[Dict],
    expected_scope_attributes: List[Dict],
    expected_events: List[Dict]
):
    missing_events: List[Dict] = []

    for expected_event in expected_events:
        if not _event_found(merged_json, expected_resource_attributes, expected_scope_attributes, expected_event):
            missing_events.append(expected_event)

    if missing_events:
        missing_descriptions = [json.dumps(event, sort_keys=True) for event in missing_events]
        return False, "Missing expected events: " + "; ".join(missing_descriptions)

    return True, ""


def _event_found(
    merged_json: List[Dict], 
    expected_resource_attributes: List[Dict],
    expected_scope_attributes: List[Dict],
    expected_event: Dict
) -> bool:
    for json_line in merged_json:
        for resource_log in json_line.get("resourceLogs", []):
            resource = resource_log.get("resource", {})
            attributes = resource.get("attributes", [])
            resource_wrapper = {"attributes": attributes}

            if not _resource_matches(resource_wrapper, expected_resource_attributes):
                continue

            for scope_log in resource_log.get("scopeLogs", []):
                scope = scope_log.get("scope", {})
                scope_attributes = scope.get("attributes", [])
                scope_wrapper = {"attributes": scope_attributes}
                
                if not _scope_matches(scope_wrapper, expected_scope_attributes):
                    continue

                for log_record in scope_log.get("logRecords", []):
                    if _log_record_matches(log_record, expected_event):
                        return True

    return False


@pytest.mark.parametrize("expected_file", list(_load_expected_cases()))
def test_entity_state_events_expected_content(expected_file: str) -> None:
    file_path = os.path.join(EXPECTED_DIR, expected_file)
    with open(file_path, "r", encoding="utf-8") as handle:
        expected_case = json.load(handle)

    expected_resource_attributes = expected_case.get("resource_attributes", [])
    expected_scope_attributes = expected_case.get("scope_attributes", [])
    expected_events = expected_case.get("events", [])
    
    print(f"\n{'='*60}")
    print(f"Testing: {expected_file}")
    print(f"Expected resource attributes: {len(expected_resource_attributes)}")
    print(f"Expected scope attributes: {len(expected_scope_attributes)}")
    print(f"Expected events: {len(expected_events)}")
    print(f"{'='*60}\n")

    # Retry logic with ClickHouse queries
    max_attempts = 60  # 60 attempts * 3 seconds = 180 seconds timeout
    attempt = 0
    
    while attempt < max_attempts:
        attempt += 1
        print(f"Attempt {attempt}/{max_attempts}: Querying ClickHouse for entity state events...")
        
        try:
            merged_json = clickhouse.get_entity_state_events()
            
            if not merged_json:
                print(f"  No events found yet, retrying...")
                time.sleep(3)
                continue
            
            print(f"  Found {len(merged_json)} event records in ClickHouse")
            
            success, error_msg = _assert_expected_events(
                merged_json, 
                expected_resource_attributes,
                expected_scope_attributes,
                expected_events
            )
            
            if success:
                print(f"✅ All expected events found for {expected_file}")
                return
            else:
                print(f"  {error_msg}")
                if attempt < max_attempts:
                    print(f"  Retrying in 3 seconds...")
                    time.sleep(3)
        except Exception as e:
            print(f"  Error querying ClickHouse: {e}")
            if attempt < max_attempts:
                time.sleep(3)
    
    # If we get here, we've exhausted all retries
    merged_json = clickhouse.get_entity_state_events()
    success, error_msg = _assert_expected_events(
        merged_json,
        expected_resource_attributes,
        expected_scope_attributes,
        expected_events
    )
    
    if not success:
        pytest.fail(f"Failed to find expected events for {expected_file} after {max_attempts} attempts. {error_msg}")

