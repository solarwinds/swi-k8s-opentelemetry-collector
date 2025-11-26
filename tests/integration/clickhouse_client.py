"""ClickHouse client for integration tests.

This module provides utilities for querying ClickHouse and converting
the results to OTLP-compatible format for test validation.
"""

import json
import os
import sys
from datetime import datetime
from typing import Dict, List

try:
    import clickhouse_connect
except ImportError:
    print("Error: clickhouse-connect not installed", file=sys.stderr)
    print("Install it with: pip install clickhouse-connect", file=sys.stderr)
    sys.exit(1)


class ClickHouseClient:
    """Client for interacting with ClickHouse in integration tests."""
    
    def __init__(self, endpoint: str = None):
        """Initialize ClickHouse client.
        
        Args:
            endpoint: ClickHouse HTTP endpoint (host:port). 
                     Defaults to CLICKHOUSE_ENDPOINT env var or 'localhost:8123'.
        """
        endpoint = endpoint or os.getenv("CLICKHOUSE_ENDPOINT", "localhost:8123")
        
        # Parse host and port from endpoint
        if ':' in endpoint:
            host, port_str = endpoint.split(':', 1)
            port = int(port_str)
        else:
            host = endpoint
            port = 8123
        
        self.client = clickhouse_connect.get_client(
            host=host,
            port=port,
            username='default',
            password=''
        )
    
    def query(self, query: str, timeout: int = 10) -> List[Dict]:
        """Execute a query against ClickHouse and return results as list of dicts.
        
        Args:
            query: SQL query to execute
            timeout: Request timeout in seconds (currently unused with clickhouse-connect)
            
        Returns:
            List of dictionaries representing query results
            
        Raises:
            Exception: If query fails
        """
        try:
            # Execute query and get result
            result = self.client.query(query)
            
            # Convert to list of dictionaries
            if not result.result_rows:
                return []
            
            # Get column names
            column_names = result.column_names
            
            # Convert rows to dictionaries
            rows = []
            for row in result.result_rows:
                row_dict = {}
                for i, col_name in enumerate(column_names):
                    row_dict[col_name] = row[i]
                rows.append(row_dict)
            
            return rows
        except Exception as e:
            print(f"ClickHouse query failed: {e}")
            raise
    
    def get_entity_state_events(self) -> List[Dict]:
        """Fetch all entity state events from ClickHouse in OTLP format.
        
        Queries the otel_logs table for logs with the entity event marker
        (ScopeAttributes['otel.entity.event_as_log'] = 'true') and converts
        them to OTLP-compatible JSON format.
        
        Returns:
            List of OTLP-style resource logs containing entity state events
        """
        query = """
        SELECT 
            Timestamp,
            ResourceAttributes,
            ScopeAttributes,
            LogAttributes
        FROM otel.otel_logs
        WHERE ScopeAttributes['otel.entity.event_as_log'] = 'true'
        ORDER BY Timestamp DESC
        """
        
        try:
            rows = self.query(query)
        except Exception:
            return []
        
        # Convert ClickHouse format to OTLP JSON format
        resource_logs = []
        
        for row in rows:
            resource_attrs = self._convert_map_to_attributes(row.get('ResourceAttributes', {}))
            scope_attrs = self._convert_map_to_attributes(row.get('ScopeAttributes', {}))
            log_attrs = self._convert_map_to_attributes(row.get('LogAttributes', {}))
            
            # Build OTLP-like structure
            log_record = {
                'timeUnixNano': str(self._parse_clickhouse_timestamp(row['Timestamp'])),
                'attributes': log_attrs
            }
            
            resource_log = {
                'resource': {
                    'attributes': resource_attrs
                },
                'scopeLogs': [{
                    'scope': {
                        'attributes': scope_attrs
                    },
                    'logRecords': [log_record]
                }]
            }
            
            resource_logs.append({'resourceLogs': [resource_log]})
        
        return resource_logs
    
    def get_logs(self, where_clause: str = None) -> List[Dict]:
        """Fetch logs from ClickHouse in OTLP format.
        
        Args:
            where_clause: Optional WHERE clause to filter logs (without 'WHERE' keyword)
            
        Returns:
            List of OTLP-style resource logs
        """
        query = """
        SELECT 
            Timestamp,
            ResourceAttributes,
            ScopeAttributes,
            LogAttributes,
            Body,
            SeverityText,
            SeverityNumber
        FROM otel.otel_logs
        """
        
        if where_clause:
            query += f"\nWHERE {where_clause}"
        
        query += "\nORDER BY Timestamp DESC"
        
        try:
            rows = self.query(query)
        except Exception:
            return []
        
        resource_logs = []
        
        for row in rows:
            resource_attrs = self._convert_map_to_attributes(row.get('ResourceAttributes', {}))
            scope_attrs = self._convert_map_to_attributes(row.get('ScopeAttributes', {}))
            log_attrs = self._convert_map_to_attributes(row.get('LogAttributes', {}))
            
            log_record = {
                'timeUnixNano': str(self._parse_clickhouse_timestamp(row['Timestamp'])),
                'attributes': log_attrs,
                'body': {'stringValue': row.get('Body', '')},
                'severityText': row.get('SeverityText', ''),
                'severityNumber': row.get('SeverityNumber', 0)
            }
            
            resource_log = {
                'resource': {
                    'attributes': resource_attrs
                },
                'scopeLogs': [{
                    'scope': {
                        'attributes': scope_attrs
                    },
                    'logRecords': [log_record]
                }]
            }
            
            resource_logs.append({'resourceLogs': [resource_log]})
        
        return resource_logs
    
    def get_metrics(self, where_clause: str = None) -> List[Dict]:
        """Fetch metrics from ClickHouse.
        
        Args:
            where_clause: Optional WHERE clause to filter metrics (without 'WHERE' keyword)
            
        Returns:
            List of metric records from ClickHouse
        """
        # Note: ClickHouse stores metrics in multiple tables based on type
        # This is a basic implementation - expand as needed
        query = """
        SELECT 
            TimeUnix,
            ResourceAttributes,
            ScopeAttributes,
            MetricName,
            Attributes,
            Value
        FROM otel.otel_metrics_gauge
        """
        
        if where_clause:
            query += f"\nWHERE {where_clause}"
        
        query += "\nORDER BY TimeUnix DESC"
        
        try:
            return self.query(query)
        except Exception:
            return []
    
    def get_metrics_otlp(self, where_clause: str = None) -> List[Dict]:
        """Fetch all metrics from ClickHouse in OTLP JSON format.
        
        Queries all metric tables (gauge, sum, summary, histogram, exp_histogram)
        and converts them to OTLP-compatible JSON format for test validation.
        
        Args:
            where_clause: Optional WHERE clause to filter metrics (without 'WHERE' keyword)
            
        Returns:
            List of OTLP-style resource metrics
        """
        all_metrics = []
        
        # Query gauge metrics
        gauge_query = """
        SELECT 
            TimeUnix,
            ResourceAttributes,
            ScopeAttributes,
            MetricName,
            Attributes,
            Value
        FROM otel.otel_metrics_gauge
        """
        if where_clause:
            gauge_query += f"\nWHERE {where_clause}"
        gauge_query += "\nORDER BY TimeUnix DESC"
        
        # Query sum metrics
        sum_query = """
        SELECT 
            TimeUnix,
            ResourceAttributes,
            ScopeAttributes,
            MetricName,
            Attributes,
            Value,
            IsMonotonic,
            AggregationTemporality
        FROM otel.otel_metrics_sum
        """
        if where_clause:
            sum_query += f"\nWHERE {where_clause}"
        sum_query += "\nORDER BY TimeUnix DESC"
        
        # Query histogram metrics
        histogram_query = """
        SELECT 
            TimeUnix,
            ResourceAttributes,
            ScopeAttributes,
            MetricName,
            Attributes,
            Count,
            Sum,
            BucketCounts,
            ExplicitBounds,
            Min,
            Max,
            AggregationTemporality
        FROM otel.otel_metrics_histogram
        """
        if where_clause:
            histogram_query += f"\nWHERE {where_clause}"
        histogram_query += "\nORDER BY TimeUnix DESC"
        
        try:
            # Fetch gauge metrics
            gauge_rows = self.query(gauge_query)
            for row in gauge_rows:
                all_metrics.append(self._convert_metric_row_to_otlp(row, 'gauge'))
            
            # Fetch sum metrics
            sum_rows = self.query(sum_query)
            for row in sum_rows:
                all_metrics.append(self._convert_metric_row_to_otlp(row, 'sum'))
            
            # Fetch histogram metrics
            histogram_rows = self.query(histogram_query)
            for row in histogram_rows:
                all_metrics.append(self._convert_metric_row_to_otlp(row, 'histogram'))
                
        except Exception as e:
            print(f"Error fetching metrics from ClickHouse: {e}")
            return []
        
        # Group metrics by resource and scope
        return self._group_metrics_by_resource(all_metrics)
    
    def _convert_metric_row_to_otlp(self, row: Dict, metric_type: str) -> Dict:
        """Convert a single metric row from ClickHouse to OTLP format.
        
        Args:
            row: Metric row from ClickHouse
            metric_type: Type of metric ('gauge', 'sum', 'summary', 'histogram')
            
        Returns:
            Dict with metric data and metadata
        """
        resource_attrs = self._convert_map_to_attributes(row.get('ResourceAttributes', {}))
        scope_attrs = self._convert_map_to_attributes(row.get('ScopeAttributes', {}))
        metric_attrs = self._convert_map_to_attributes(row.get('Attributes', {}))
        
        # Build datapoint based on metric type
        datapoint = {
            'timeUnixNano': str(self._parse_clickhouse_timestamp(row['TimeUnix'])),
            'attributes': metric_attrs
        }
        
        if metric_type == 'histogram':
            datapoint['count'] = row.get('Count', 0)
            datapoint['sum'] = row.get('Sum', 0.0)
            datapoint['bucketCounts'] = row.get('BucketCounts', [])
            datapoint['explicitBounds'] = row.get('ExplicitBounds', [])
            if 'Min' in row:
                datapoint['min'] = row['Min']
            if 'Max' in row:
                datapoint['max'] = row['Max']
        else:
            # Add value based on type
            value = row.get('Value', 0)
            if isinstance(value, float):
                datapoint['asDouble'] = value
            elif isinstance(value, int):
                datapoint['asInt'] = value
            else:
                datapoint['asDouble'] = float(value)
        
        # Build metric structure
        metric = {
            'name': row['MetricName'],
            metric_type: {
                'dataPoints': [datapoint]
            }
        }
        
        # Add sum-specific fields
        if metric_type == 'sum' and 'IsMonotonic' in row:
            metric[metric_type]['isMonotonic'] = row.get('IsMonotonic', False)
            metric[metric_type]['aggregationTemporality'] = row.get('AggregationTemporality', 2)
            
        # Add histogram-specific fields
        if metric_type == 'histogram':
            metric[metric_type]['aggregationTemporality'] = row.get('AggregationTemporality', 2)
        
        return {
            'resource_attrs': resource_attrs,
            'scope_attrs': scope_attrs,
            'metric': metric
        }
    
    def _group_metrics_by_resource(self, metrics_list: List[Dict]) -> List[Dict]:
        """Group metrics by resource and scope into OTLP structure.
        
        Args:
            metrics_list: List of metric dicts with resource_attrs, scope_attrs, and metric
            
        Returns:
            List of OTLP-style resource metrics
        """
        # Group by resource attributes (as JSON string for dict key)
        resource_map = {}
        
        for metric_data in metrics_list:
            # Create keys for grouping
            resource_key = json.dumps(metric_data['resource_attrs'], sort_keys=True)
            scope_key = json.dumps(metric_data['scope_attrs'], sort_keys=True)
            
            if resource_key not in resource_map:
                resource_map[resource_key] = {
                    'resource_attrs': metric_data['resource_attrs'],
                    'scopes': {}
                }
            
            if scope_key not in resource_map[resource_key]['scopes']:
                resource_map[resource_key]['scopes'][scope_key] = {
                    'scope_attrs': metric_data['scope_attrs'],
                    'metrics': []
                }
            
            resource_map[resource_key]['scopes'][scope_key]['metrics'].append(metric_data['metric'])
        
        # Convert to OTLP format
        result = []
        for resource_data in resource_map.values():
            scope_metrics = []
            for scope_data in resource_data['scopes'].values():
                scope_metrics.append({
                    'scope': {
                        'attributes': scope_data['scope_attrs']
                    },
                    'metrics': scope_data['metrics']
                })
            
            result.append({
                'resourceMetrics': [{
                    'resource': {
                        'attributes': resource_data['resource_attrs']
                    },
                    'scopeMetrics': scope_metrics
                }]
            })
        
        return result
    
    def count_records(self, table: str, where_clause: str = None) -> int:
        """Count records in a ClickHouse table.
        
        Args:
            table: Table name (e.g., 'otel.otel_logs', 'otel.otel_metrics_gauge')
            where_clause: Optional WHERE clause (without 'WHERE' keyword)
            
        Returns:
            Number of records matching the criteria
        """
        query = f"SELECT count(*) as count FROM {table}"
        
        if where_clause:
            query += f" WHERE {where_clause}"
        
        try:
            result = self.query(query)
            if result:
                return result[0].get('count', 0)
        except Exception:
            pass
        
        return 0
    
    @staticmethod
    def _convert_map_to_attributes(attr_map: Dict[str, str]) -> List[Dict]:
        """Convert ClickHouse Map to OTLP attribute array format.
        
        Args:
            attr_map: Dictionary of attribute key-value pairs
            
        Returns:
            List of attribute objects in OTLP format
        """
        attributes = []
        for key, value in attr_map.items():
            # Convert non-string types to strings first
            if not isinstance(value, str):
                if isinstance(value, datetime):
                    value = value.isoformat()
                else:
                    value = str(value)
            
            # Check if the value is a JSON string that should be parsed
            if key.endswith('.id') or \
               key == 'otel.entity.id' or \
               key == 'otel.entity_relationship.source_entity.id' or \
               key == 'otel.entity_relationship.destination_entity.id' or \
               key == 'otel.entity.attributes':
                try:
                    # Try to parse as JSON
                    parsed = json.loads(value)
                    if isinstance(parsed, dict):
                        # Convert dict to kvlistValue format
                        kv_list = []
                        for k, v in parsed.items():
                            kv_list.append({
                                'key': k,
                                'value': {'stringValue': v}
                            })
                        attributes.append({
                            'key': key,
                            'value': {'kvlistValue': {'values': kv_list}}
                        })
                        continue
                    elif isinstance(parsed, list):
                        # Already in the right format
                        attributes.append({
                            'key': key,
                            'value': {'kvlistValue': {'values': parsed}}
                        })
                        continue
                except (json.JSONDecodeError, TypeError):
                    # Not JSON, treat as string
                    pass
            
            # Default: treat as string value
            attributes.append({
                'key': key,
                'value': {'stringValue': value}
            })
        return attributes
    
    @staticmethod
    def _parse_clickhouse_timestamp(timestamp) -> int:
        """Parse ClickHouse timestamp to Unix nanoseconds.
        
        ClickHouse can return timestamps as either datetime objects (via clickhouse-connect)
        or strings (via raw HTTP). This method handles both formats.
        
        Args:
            timestamp: Timestamp as datetime object or string (e.g., '2025-11-08 16:11:44.626029090')
            
        Returns:
            Unix timestamp in nanoseconds
        """
        # If it's already a datetime object, convert directly
        if isinstance(timestamp, datetime):
            return int(timestamp.timestamp() * 1e9)
        
        # Otherwise, parse the string
        timestamp_str = str(timestamp)
        
        # Split timestamp into main part and fractional seconds
        if '.' in timestamp_str:
            main_part, frac_part = timestamp_str.split('.')
            # Truncate to 6 digits for microseconds (Python's %f limitation)
            frac_part_us = frac_part[:6]
            # Keep the remaining nanoseconds
            if len(frac_part) > 6:
                extra_nanos = frac_part[6:]
            else:
                extra_nanos = '0'
            
            # Parse the main part with microseconds
            timestamp_us = f"{main_part}.{frac_part_us}"
            dt = datetime.strptime(timestamp_us, '%Y-%m-%d %H:%M:%S.%f')
            
            # Convert to nanoseconds
            unix_nanos = int(dt.timestamp() * 1e9)
            # Add the extra nanoseconds that were truncated
            unix_nanos += int(extra_nanos.ljust(3, '0'))  # Pad to 3 digits for remaining nanos
            
            return unix_nanos
        else:
            # No fractional seconds
            dt = datetime.strptime(timestamp_str, '%Y-%m-%d %H:%M:%S')
            return int(dt.timestamp() * 1e9)
