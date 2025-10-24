---
description: 'YAML and Helm template coding conventions and guidelines'
applyTo: '**/*.yaml,**/*.yml,**/*.tpl'
---

# YAML and Helm Template Conventions

## General YAML Structure

- Use **2 spaces** for indentation (never tabs).
- Maintain consistent indentation throughout all YAML files.
- Use lowercase with underscores for keys (snake_case) in configuration files: `my_config_key: value`.
- Use camelCase for Kubernetes resource keys following K8s conventions: `serviceAccountName`, `nodeSelector`.
- Keep lines under 120 characters when possible.
- Always quote string values that may be ambiguous (e.g., version numbers, special characters).
- Use block scalars (|, |-) for multi-line strings to improve readability.

## OpenTelemetry Collector Configuration

### Standard Structure Order

Follow this consistent order for OTEL collector configuration sections:
1. `exporters:`
2. `extensions:`
3. `processors:`
4. `connectors:`
5. `receivers:`
6. `service:`

### Pipeline Naming Conventions

- Use descriptive pipeline names with type prefix: `metrics/`, `logs/`, `traces/`
- Examples: `metrics/discovery`, `logs/container`, `metrics/prometheus-server`
- Suffix pipelines for specific purposes: `/stateevents`, `/relationships`, `/discovery`
- For temporary or intermediate pipelines, use descriptive suffixes: `/clean`, `/preparation`, `/filtering`

### Processor Configuration

- Group related processors logically in pipeline definitions
- Use filter processors with descriptive names: `filter/receiver`, `filter/remove_internal`, `filter/keep-entity-state-events`
- Place resource processors at the end of processing chain: `resource/metrics`, `resource/all`
- Use transform processors for attribute manipulation: `transform/istio-metrics`, `transform/unify_node_attribute`
- Batch processors should be last before exporters: `batch/stateevents`

### Metric Transformations

- Use `metricstransform` for renaming and aggregating metrics
- Temporary metrics during processing should use `__swo_temp` suffix
- Follow naming convention: `k8s.<entity>.<metric>.<aggregation>`
  - Examples: `k8s.pod.cpu.usage.seconds.rate`, `k8s.node.fs.throughput`
- Use `experimental_match_labels` for label-based filtering and aggregation
- Always document complex metric transformations with comments

## Helm Template Conventions

### Template Function Usage

#### Common Helper Templates

The repository defines reusable templates in `_helpers.tpl` and `_common-config.tpl`:

**Identity and Naming:**
- `{{ include "common.fullname" . }}` - Generate resource names
- `{{ include "common.chart" . }}` - Chart name and version
- `{{ include "common.cluster-uid" . }}` - Cluster UID from values

**Labels and Annotations:**
- `{{ include "common.labels" . | indent 4 }}` - Standard labels
- `{{ include "common.annotations" . | indent 4 }}` - Standard annotations
- `{{ include "common.pod-labels" . | indent 8 }}` - Pod-specific labels

**Image Resolution:**
- `{{ include "common.image" (tuple . .Values.otel "image") }}` - Resolve image path
- Supports Azure-specific image configuration
- Handles repository, tag, and digest resolution

**OTEL Configuration Helpers:**
- `{{ include "common.k8s-instrumentation" . }}` - K8s instrumentation config
- `{{ include "common.events-error-conditions" . }}` - Event error filter conditions
- `{{ include "common.events-warning-conditions" . }}` - Event warning filter conditions
- `{{ include "common.maxStaleness" . }}` - Calculate max_staleness from scrape_interval
- `{{ include "common.tripleInterval" . }}` - Triple a time interval
- `{{ include "common.prometheus.relabelconfigs" . }}` - Prometheus relabel configs

**Pipeline Components:**
- `{{ include "common-config.filter-reciever" . }}` - Filter receiver metrics
- `{{ include "common-config.filter-remove-internal" . }}` - Remove internal containers
- `{{ include "common-config.transform-node-attributes" . }}` - Node attribute transformation
- `{{ include "common-config.metricstransform-preprocessing-cadvisor" . }}` - cAdvisor preprocessing
- `{{ include "common-config.resource-metrics" . }}` - Resource attribute mapping
- `{{ include "common-config.groupbyattrs-node" . }}` - Group by node attributes
- `{{ include "common-config.groupbyattrs-pod" . }}` - Group by pod attributes
- `{{ include "common-config.groupbyattrs-all" . }}` - Group by all relevant attributes

**Discovery Configuration:**
- `{{ include "common-discovery-config.processors" . }}` - Discovery processors
- `{{ include "common-discovery-config.connectors" . }}` - Discovery connectors
- `{{ include "common-discovery-config.pipelines" (tuple . $receiver $exporter) }}` - Complete discovery pipelines

### Value Access Patterns

- Always use `.Values` prefix for chart values
- Use `.Release.Name`, `.Release.Namespace` for release info
- Use `.Chart.Name`, `.Chart.Version`, `.Chart.AppVersion` for chart metadata
- Access nested values with dot notation: `.Values.otel.metrics.enabled`

### Environment Variables

- Use `envFrom` with `configMapRef` or `secretRef` for bulk env var injection
- Define individual env vars in `env:` section when needed
- Environment variable names should be SCREAMING_SNAKE_CASE: `CLUSTER_UID`, `MANIFEST_VERSION`
- Reference ConfigMaps/Secrets: `valueFrom.configMapKeyRef`, `valueFrom.secretKeyRef`

## Kubernetes Resource Best Practices

### Metadata

```yaml
metadata:
  name: {{ include "common.fullname" (tuple . "-component") }}
  namespace: {{ .Release.Namespace }}
  labels:
    app.kubernetes.io/name: swo-k8s-collector
{{ include "common.labels" . | indent 4 }}
  annotations:
{{ include "common.annotations" . | indent 4 }}
```

### Pod Specifications

- Always specify `terminationGracePeriodSeconds`
- Use `serviceAccountName: {{ include "common.fullname" . }}`
- Define `nodeSelector`, `tolerations`, and `affinity` conditionally from values
- Use `imagePullSecrets` when provided in values
- Include resource limits and requests

### Volume Mounts

- Define volumes and volumeMounts separately
- Use descriptive volume names: `config-vol`, `storage-vol`, `secret-vol`
- Document the purpose of each volume with comments
- Use `subPath` when mounting specific files from ConfigMaps

## Comments and Documentation

- Add comments for complex transformations or non-obvious configurations
- Document the purpose of filters and processors
- Explain regex patterns and complex conditionals
- Add section headers for major configuration blocks:

```yaml
# ===== Receivers Configuration =====
receivers:
  # Prometheus receiver for scraping metrics
  prometheus:
    config:
      # ...
```

## Validation and Schema

- All values must be defined in `values.schema.json`
- Use JSON Schema for type validation and constraints
- Provide descriptions for all schema properties
- Define required fields and default values
- Use `additionalProperties: false` to prevent typos

## Backward Compatibility

- Never remove or rename existing values without deprecation period
- Use feature flags for new functionality: `.Values.feature.enabled`
- Maintain deprecated fields with warnings in NOTES.txt
- Document breaking changes in CHANGELOG.md
- Support multiple API versions when needed

## Filters and Conditions

### OTTL (OpenTelemetry Transformation Language)

```yaml
# Attribute-based filtering
filter/example:
  metrics:
    datapoint:
      - 'attributes["container"] == "POD"'
      - 'IsMatch(metric.name, "^container_network_.*")'
      
  logs:
    log_record:
      - 'not(attributes["otel.entity.event.type"] == "entity_state")'
```

### Common Filter Patterns

- Use `IsMatch()` for regex matching
- Combine conditions with `and`, `or`, `not()`
- Filter by metric name: `metric.name == "specific_metric"`
- Filter by resource attributes: `resource.attributes["k8s.namespace.name"]`
- Filter by datapoint attributes: `datapoint.attributes["key"]`

## Common Patterns to Avoid

- Don't hardcode values that should be configurable
- Don't use complex logic directly in templates (use helpers instead)
- Avoid deeply nested conditionals (refactor to helper templates)
- Don't duplicate configuration across files (use includes)
- Avoid inconsistent indentation or formatting
- Don't create circular dependencies between templates
