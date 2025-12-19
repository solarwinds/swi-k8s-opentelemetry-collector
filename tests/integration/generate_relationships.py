#!/usr/bin/env python3
"""
Generate expected entity state events and relationships JSON files from ClickHouse data.

This script connects to ClickHouse (via port-forward) and generates JSON files
for integration test assertions. It automatically discovers all entity types and
relationship types in the database.

Prerequisites:
    pip install clickhouse-connect

Usage:
    # Make sure Skaffold is running (handles port-forwarding automatically):
    skaffold dev

    # Run this script to generate all discovered entities and relationships:
    python3 generate_relationships.py

    # Or with custom host/port:
    python3 generate_relationships.py --host localhost --port 8123
"""

import argparse
import json
import os
import re
import sys
from typing import Dict, List, Any

try:
    import clickhouse_connect
except ImportError:
    print("Error: clickhouse-connect not installed", file=sys.stderr)
    print("Install it with: pip install clickhouse-connect", file=sys.stderr)
    sys.exit(1)


# Constants
TEST_PREFIX = 'test-'
CLUSTER_UID_KEY = 'sw.k8s.cluster.uid'
POD_NAME_KEY = 'k8s.pod.name'
REPLICASET_NAME_KEY = 'k8s.replicaset.name'
JOB_NAME_KEY = 'k8s.job.name'

# Kubernetes workload name keys in priority order
K8S_WORKLOAD_KEYS = [
    'k8s.deployment.name',
    'k8s.statefulset.name',
    'k8s.daemonset.name',
    'k8s.service.name',
    'k8s.pod.name',
    'k8s.replicaset.name',
    'k8s.job.name'
]

# Hash pattern constants
POD_HASH_LENGTH = 5
REPLICASET_HASH_MIN_LENGTH = 8
REPLICASET_HASH_MAX_LENGTH = 10
JOB_HASH_MIN_LENGTH = 8
JOB_HASH_MAX_LENGTH = 10


def parse_entity_id(entity_id_str: str) -> Dict[str, str]:
    try:
        return json.loads(entity_id_str)
    except (json.JSONDecodeError, TypeError):
        return {}


def to_snake_case(name: str) -> str:
    s1 = re.sub('(.)([A-Z][a-z]+)', r'\1_\2', name)
    return re.sub('([a-z0-9])([A-Z])', r'\1_\2', s1).lower()


def get_workload_name(entity_id: Dict[str, str]) -> str:
    for key in K8S_WORKLOAD_KEYS:
        if key in entity_id:
            return entity_id[key]
    return ""


def _check_pod_hash_pattern(pod_name: str) -> bool:
    """Returns True if pod name matches: <base>-<8-10chars>-<5chars> pattern.
    
    This identifies pods managed by Deployments/ReplicaSets or CronJobs,
    which get runtime-generated hashes that change on each deployment.
    """
    parts = pod_name.split('-')
    if len(parts) < 3:
        return False
    
    last_part = parts[-1]
    second_last_part = parts[-2]
    
    if len(last_part) != POD_HASH_LENGTH or not last_part.isalnum():
        return False
    
    hash_length = len(second_last_part)
    return (REPLICASET_HASH_MIN_LENGTH <= hash_length <= REPLICASET_HASH_MAX_LENGTH 
            and second_last_part.isalnum())


def _check_replicaset_hash_pattern(replicaset_name: str) -> bool:
    """Returns True if ReplicaSet name matches: <base>-<8-10chars> pattern.
    
    This identifies ReplicaSets created by Deployments, which get
    runtime-generated hashes that change when the pod template changes.
    """
    parts = replicaset_name.split('-')
    if len(parts) < 2:
        return False
    
    last_part = parts[-1]
    hash_length = len(last_part)
    return (REPLICASET_HASH_MIN_LENGTH <= hash_length <= REPLICASET_HASH_MAX_LENGTH 
            and last_part.isalnum())

def _check_job_hash_pattern(job_name: str) -> bool:
    """Returns True if Job name matches: <base>-<hash> pattern. Usually,
    the hash is a timestamp, but not always.
    
    This identifies Jobs created by CronJobs, which get
    runtime-generated names that change when the CronJob is triggered.
    """
    parts = job_name.split('-')
    if len(parts) < 2:
        return False

    last_part = parts[-1]
    hash_length = len(last_part)
    return (JOB_HASH_MIN_LENGTH <= hash_length <= JOB_HASH_MAX_LENGTH
            and last_part.isdecimal())


def has_runtime_generated_hash(entity_id: Dict[str, str], entity_type: str) -> bool:
    """Detects entities with runtime-generated hashes that change on each deploy."""
    if entity_type in ['KubernetesPod', 'KubernetesContainer']:
        pod_name = entity_id.get(POD_NAME_KEY)
        if pod_name:
            return _check_pod_hash_pattern(pod_name)
    
    elif entity_type == 'KubernetesReplicaSet':
        replicaset_name = entity_id.get(REPLICASET_NAME_KEY)
        if replicaset_name:
            return _check_replicaset_hash_pattern(replicaset_name)
        
    elif entity_type == 'KubernetesJob':
        job_name = entity_id.get(JOB_NAME_KEY)
        if job_name:
            return _check_job_hash_pattern(job_name)
    
    return False


def discover_entity_types(client) -> List[str]:
    query = """
    SELECT DISTINCT LogAttributes['otel.entity.type'] as entity_type
    FROM otel.otel_logs
    WHERE ScopeAttributes['otel.entity.event_as_log'] = 'true'
    AND LogAttributes['otel.entity.event.type'] = 'entity_state'
    AND LogAttributes['otel.entity.type'] != ''
    ORDER BY entity_type
    """
    result = client.query(query)
    return [row[0] for row in result.result_rows if row[0]]


def discover_relationship_types(client) -> List[str]:
    query = """
    SELECT DISTINCT LogAttributes['otel.entity_relationship.type'] as relationship_type
    FROM otel.otel_logs
    WHERE ScopeAttributes['otel.entity.event_as_log'] = 'true'
    AND LogAttributes['otel.entity.event.type'] = 'entity_relationship_state'
    AND LogAttributes['otel.entity_relationship.type'] != ''
    ORDER BY relationship_type
    """
    result = client.query(query)
    return [row[0] for row in result.result_rows if row[0]]


def get_entity_filter_key(entity_type: str) -> str | None:
    """Returns k8s.<workloadtype>.name for Kubernetes entities, None otherwise."""
    if not entity_type.startswith('Kubernetes'):
        return None
    
    workload_type = entity_type[len('Kubernetes'):]
    if not workload_type:
        return None
    
    return f'k8s.{workload_type.lower()}.name'


def should_include_entity(entity_id: Dict[str, str], entity_type: str) -> bool:
    """Filters entities to include only test fixtures without runtime-generated hashes."""
    filter_key = get_entity_filter_key(entity_type)
    
    if filter_key is None:
        return True
    
    entity_name = entity_id.get(filter_key)
    
    if entity_name is None or not entity_name.startswith(TEST_PREFIX):
        return False
    
    if has_runtime_generated_hash(entity_id, entity_type):
        return False
    
    return True


def filter_cluster_uid(entity_id: Dict[str, str]) -> Dict[str, str]:
    return {k: v for k, v in entity_id.items() if k != CLUSTER_UID_KEY}


def convert_entity_id_to_list(entity_id: Dict[str, str]) -> List[Dict[str, str]]:
    return [
        {"key": key, "value": value}
        for key, value in sorted(entity_id.items())
    ]


def create_output_structure(events: List[Dict]) -> Dict:
    return {
        "resource_attributes": [],
        "scope_attributes": [
            {
                "key": "otel.entity.event_as_log",
                "value": "true"
            }
        ],
        "events": events
    }


def write_json_file(data: Dict, file_path: str) -> None:
    os.makedirs(os.path.dirname(file_path), exist_ok=True)
    with open(file_path, 'w') as f:
        json.dump(data, f, indent=2, sort_keys=True)
        f.write('\n')


def generate_entity_file(client, entity_type: str, output_dir: str) -> int:
    filename = f"entity_{to_snake_case(entity_type)}.json"
    output_path = os.path.join(output_dir, filename)
    
    print(f"[entity_{to_snake_case(entity_type)}] Fetching data...", file=sys.stderr)
    
    query = f"""
    SELECT 
        LogAttributes['otel.entity.event.type'] as event_type,
        LogAttributes['otel.entity.type'] as entity_type,
        LogAttributes['otel.entity.id'] as entity_id
    FROM otel.otel_logs
    WHERE ScopeAttributes['otel.entity.event_as_log'] = 'true'
    AND LogAttributes['otel.entity.event.type'] = 'entity_state'
    AND LogAttributes['otel.entity.type'] = '{entity_type}'
    ORDER BY LogAttributes['otel.entity.id']
    """
    
    result = client.query(query)
    
    events_data = []
    for row in result.result_rows:
        events_data.append({
            'event_type': row[0],
            'entity_type': row[1],
            'entity_id': row[2]
        })
    
    print(f"[entity_{to_snake_case(entity_type)}] Fetched {len(events_data)} events", file=sys.stderr)
    
    seen_ids = set()
    unique_events = []
    
    for event in events_data:
        entity_id = parse_entity_id(event['entity_id'])
        
        if not should_include_entity(entity_id, event['entity_type']):
            continue
        
        entity_id_filtered = filter_cluster_uid(entity_id)
        entity_key = tuple(sorted(entity_id_filtered.items()))
        
        if entity_key not in seen_ids:
            seen_ids.add(entity_key)
            
            unique_events.append({
                "otel.entity.event.type": event['event_type'],
                "otel.entity.type": event['entity_type'],
                "otel.entity.id": convert_entity_id_to_list(entity_id_filtered),
                "otel.entity.attributes": []
            })
    
    unique_events.sort(key=lambda e: json.dumps(e["otel.entity.id"], sort_keys=True))
    
    print(f"[entity_{to_snake_case(entity_type)}] Processed {len(unique_events)} unique entities", file=sys.stderr)
    
    output_data = create_output_structure(unique_events)
    
    print(f"[entity_{to_snake_case(entity_type)}] Writing to {output_path}...", file=sys.stderr)
    write_json_file(output_data, output_path)
    
    print(f"[entity_{to_snake_case(entity_type)}] ✓ Done!", file=sys.stderr)
    return len(unique_events)


def generate_relationship_file(client, relationship_type: str, output_dir: str) -> int:
    filename = f"relationship_{to_snake_case(relationship_type)}.json"
    output_path = os.path.join(output_dir, filename)
    
    print(f"[relationship_{to_snake_case(relationship_type)}] Fetching data...", file=sys.stderr)
    
    query = f"""
    SELECT 
        LogAttributes['otel.entity.event.type'] as event_type,
        LogAttributes['otel.entity_relationship.type'] as relationship_type,
        LogAttributes['otel.entity_relationship.source_entity.type'] as source_type,
        LogAttributes['otel.entity_relationship.source_entity.id'] as source_id,
        LogAttributes['otel.entity_relationship.destination_entity.type'] as dest_type,
        LogAttributes['otel.entity_relationship.destination_entity.id'] as dest_id
    FROM otel.otel_logs
    WHERE ScopeAttributes['otel.entity.event_as_log'] = 'true'
    AND LogAttributes['otel.entity.event.type'] = 'entity_relationship_state'
    AND LogAttributes['otel.entity_relationship.type'] = '{relationship_type}'
    ORDER BY LogAttributes['otel.entity_relationship.source_entity.id'], LogAttributes['otel.entity_relationship.destination_entity.id']
    """
    
    result = client.query(query)
    
    relationships_data = []
    for row in result.result_rows:
        relationships_data.append({
            'event_type': row[0],
            'relationship_type': row[1],
            'source_type': row[2],
            'source_id': row[3],
            'dest_type': row[4],
            'dest_id': row[5]
        })
    
    print(f"[relationship_{to_snake_case(relationship_type)}] Fetched {len(relationships_data)} relationships", file=sys.stderr)
    
    seen_relationships = set()
    unique_events = []
    
    for rel in relationships_data:
        source_id = parse_entity_id(rel['source_id'])
        dest_id = parse_entity_id(rel['dest_id'])
        
        source_name = get_workload_name(source_id)
        dest_name = get_workload_name(dest_id)
        
        if not (source_name and source_name.startswith(TEST_PREFIX)):
            continue
        if not (dest_name and dest_name.startswith(TEST_PREFIX)):
            continue
        
        # KubernetesServiceRoutesTo: exclude relationships to entities with runtime hashes
        if relationship_type == 'KubernetesServiceRoutesTo':
            if rel['dest_type'] in ['KubernetesPod', 'KubernetesReplicaSet']:
                if has_runtime_generated_hash(dest_id, rel['dest_type']):
                    continue
            if rel['source_type'] in ['KubernetesPod', 'KubernetesReplicaSet']:
                if has_runtime_generated_hash(source_id, rel['source_type']):
                    continue
        
        source_id_filtered = filter_cluster_uid(source_id)
        dest_id_filtered = filter_cluster_uid(dest_id)
        
        source_key = tuple(sorted(source_id_filtered.items()))
        dest_key = tuple(sorted(dest_id_filtered.items()))
        rel_key = (rel['source_type'], source_key, rel['dest_type'], dest_key)
        
        if rel_key not in seen_relationships:
            seen_relationships.add(rel_key)
            
            unique_events.append({
                "otel.entity.event.type": rel['event_type'],
                "otel.entity_relationship.type": rel['relationship_type'],
                "otel.entity_relationship.source_entity.type": rel['source_type'],
                "otel.entity_relationship.source_entity.id": convert_entity_id_to_list(source_id_filtered),
                "otel.entity_relationship.destination_entity.type": rel['dest_type'],
                "otel.entity_relationship.destination_entity.id": convert_entity_id_to_list(dest_id_filtered)
            })
    
    unique_events.sort(key=lambda e: (
        e["otel.entity_relationship.source_entity.type"],
        json.dumps(e["otel.entity_relationship.source_entity.id"], sort_keys=True),
        e["otel.entity_relationship.destination_entity.type"],
        json.dumps(e["otel.entity_relationship.destination_entity.id"], sort_keys=True)
    ))
    
    print(f"[relationship_{to_snake_case(relationship_type)}] Processed {len(unique_events)} unique relationships", file=sys.stderr)
    
    output_data = create_output_structure(unique_events)
    
    print(f"[relationship_{to_snake_case(relationship_type)}] Writing to {output_path}...", file=sys.stderr)
    write_json_file(output_data, output_path)
    
    print(f"[relationship_{to_snake_case(relationship_type)}] ✓ Done!", file=sys.stderr)
    return len(unique_events)


def connect_to_clickhouse(host: str, port: int):
    print(f"Connecting to ClickHouse at {host}:{port}...", file=sys.stderr)
    
    try:
        client = clickhouse_connect.get_client(
            host=host,
            port=port,
            username='default',
            password=''
        )
        return client
    except Exception as e:
        print(f"Error connecting to ClickHouse: {e}", file=sys.stderr)
        print("\nMake sure Skaffold is running with port-forwarding:", file=sys.stderr)
        print("  skaffold dev", file=sys.stderr)
        sys.exit(1)


def main():
    parser = argparse.ArgumentParser(
        description='Generate expected entity state events and relationships JSON files from ClickHouse',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Generate all outputs (default)
  %(prog)s

  # With custom host/port
  %(prog)s --host 127.0.0.1 --port 8123

The script automatically discovers all entity types and relationship types in the
ClickHouse database and generates JSON files with pattern-based naming:
  - entity_<type_lowercase>.json (e.g., entity_publicnetworklocation.json)
  - relationship_<type_lowercase>.json (e.g., relationship_kubernetescommunicateswith.json)

Filtering rules:
  - Entities: sw.k8s.cluster.uid excluded from all entity IDs
  - Relationships: Only includes those where both source and destination workload/service names start with "test-"

Before running, make sure Skaffold is running (handles port-forwarding automatically):
  skaffold dev
        """
    )
    
    parser.add_argument(
        '--host',
        default='localhost',
        help='ClickHouse host (default: localhost)'
    )
    
    parser.add_argument(
        '--port',
        type=int,
        default=8123,
        help='ClickHouse HTTP port (default: 8123)'
    )
    
    args = parser.parse_args()
    
    # Connect to ClickHouse
    client = connect_to_clickhouse(args.host, args.port)
    
    # Get output directory
    output_dir = os.path.join(os.path.dirname(__file__), "expected_entitystateevents")
    
    # Discover and generate entity files
    print("\nDiscovering entity types...", file=sys.stderr)
    entity_types = discover_entity_types(client)
    print(f"Found {len(entity_types)} entity types: {', '.join(entity_types)}", file=sys.stderr)
    
    print("\n" + "=" * 80, file=sys.stderr)
    print("Generating entity files...", file=sys.stderr)
    print("=" * 80, file=sys.stderr)
    
    total_entities = 0
    for entity_type in entity_types:
        try:
            count = generate_entity_file(client, entity_type, output_dir)
            total_entities += count
        except Exception as e:
            print(f"\n[entity_{to_snake_case(entity_type)}] ✗ Error: {e}", file=sys.stderr)
            import traceback
            traceback.print_exc()
            sys.exit(1)
    
    # Discover and generate relationship files
    print("\n" + "=" * 80, file=sys.stderr)
    print("Discovering relationship types...", file=sys.stderr)
    relationship_types = discover_relationship_types(client)
    print(f"Found {len(relationship_types)} relationship types: {', '.join(relationship_types)}", file=sys.stderr)
    
    print("\n" + "=" * 80, file=sys.stderr)
    print("Generating relationship files...", file=sys.stderr)
    print("=" * 80, file=sys.stderr)
    
    total_relationships = 0
    for relationship_type in relationship_types:
        try:
            count = generate_relationship_file(client, relationship_type, output_dir)
            total_relationships += count
        except Exception as e:
            print(f"\n[relationship_{to_snake_case(relationship_type)}] ✗ Error: {e}", file=sys.stderr)
            import traceback
            traceback.print_exc()
            sys.exit(1)
    
    print("\n" + "=" * 80, file=sys.stderr)
    print("✓ All files generated successfully!", file=sys.stderr)
    print(f"  - Generated {len(entity_types)} entity files ({total_entities} total entities)", file=sys.stderr)
    print(f"  - Generated {len(relationship_types)} relationship files ({total_relationships} total relationships)", file=sys.stderr)
    print(f"  - Output directory: {output_dir}", file=sys.stderr)
    print("=" * 80, file=sys.stderr)


if __name__ == '__main__':
    main()
