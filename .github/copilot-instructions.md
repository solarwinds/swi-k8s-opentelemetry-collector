# GitHub Copilot Instructions for SWO K8s Collector

## Priority Guidelines

When generating code for this repository:

1. **Context-Driven Patterns**: Derive all code patterns, structure, and conventions strictly from existing code in this repository
2. **Technology-Specific Files**: Prioritize patterns and standards defined in `.github/instructions/` directory (python.instructions.md, yaml.instructions.md)
3. **Architectural Consistency**: Maintain the multi-component OpenTelemetry collector architecture with clear separation between MetricsCollector, MetricsDiscovery, EventsCollector, NodeCollector, and Gateway
4. **Backward Compatibility**: Never remove or rename existing Helm values without deprecation period; maintain compatibility with existing deployments
5. **Schema Validation**: All Helm value changes must be reflected in `values.schema.json` with proper types and descriptions

## Project Structure

### Core Technologies
- **OpenTelemetry**: SolarWinds-built OTEL Collector with custom components
- **Helm**: Chart-based deployment with extensive templating
- **Kubernetes**: Multi-component architecture (Deployments, StatefulSets, DaemonSets)
- **Python**: Integration tests and utility scripts
- **YAML**: Extensive OTEL pipeline configurations

### Directory Layout
- **deploy/helm/** – Helm chart for deploying the collector
  - **templates/** – Kubernetes resource definitions
    - **metrics-deployment.yaml** – MetricsCollector deployment
    - **metrics-discovery-deployment.yaml** – MetricsDiscovery deployment
    - **events-collector-statefulset.yaml** – EventsCollector statefulset
    - **node-collector-daemon-set.yaml** – NodeCollector daemonset (Linux)
    - **node-collector-daemon-set-windows.yaml** – NodeCollector daemonset (Windows)
    - **gateway/**, **network/**, **beyla/**, **operator/**, **autoupdate/**, **openshift/** – Specialized components
    - **_helpers.tpl**, **_common-config.tpl**, **_common-discovery-config.tpl** – Reusable template helpers
  - **metrics-collector-config.yaml** – MetricsCollector OTEL pipeline
  - **metrics-discovery-config.yaml** – MetricsDiscovery OTEL pipeline
  - **events-collector-config.yaml** – EventsCollector OTEL pipeline
  - **gateway-collector-config.yaml** – Gateway OTEL pipeline
  - **node-collector-config.yaml** – NodeCollector OTEL pipeline
  - **values.yaml** – Chart default values
  - **values.schema.json** – JSON schema for value validation
- **tests/integration/** – Python-based integration tests
- **doc/** – Architecture and development documentation
- **utils/** – Python utility scripts

## OTEL Collector Architecture

### MetricsCollector (Deployment)
Cluster-level metrics collection with pipelines:
- `metrics` – Main export via OTLP
- `metrics/kubestatemetrics` – Kube State Metrics scraping
- `metrics/otlp`, `metrics/prometheus` – Protocol receivers
- `metrics/prometheus-node-metrics`, `metrics/prometheus-server` – Prometheus integrations

### MetricsDiscovery (Deployment)
Auto-discovery for annotated pods (AWS Fargate support):
- `metrics/discovery` – receiver_creator + k8s_observer
- `metrics` – Process and export discovered metrics

### EventsCollector (StatefulSet)
Kubernetes events and manifests:
- `logs` – k8s_events receiver for cluster events
- `logs/manifests` – k8sobjects receiver for K8s object collection

### NodeCollector (DaemonSet)
Node-level telemetry collection:
- `logs`, `logs/container`, `logs/journal` – Log collection pipelines
- `metrics`, `metrics/discovery`, `metrics/node` – Node metrics pipelines

### Gateway (Deployment)
Central aggregation layer:
- `traces`, `metrics`, `logs` – OTLP receivers for forwarding

## Coding Patterns (Observed in Codebase)

### Helm Template Conventions
- **Helper Usage**: Always use `{{ include "common.fullname" . }}` for resource names, never hardcode
- **Image Resolution**: Use `{{ include "common.image" (tuple . .Values.otel "image") }}` pattern
- **Labels**: Apply `{{ include "common.labels" . | indent 4 }}` consistently
- **Annotations**: Include `{{ include "common.annotations" . | indent 4 }}` on all resources
- **Conditionals**: Use `.Values.component.enabled` checks before rendering optional resources
- **Indentation**: Match template indentation precisely: `{{ toYaml .Values.config | indent 4 }}`

### OTEL Configuration Patterns
- **Section Order**: Always follow: exporters → extensions → processors → connectors → receivers → service
- **Pipeline Naming**: Use descriptive type-prefixed names: `metrics/discovery`, `logs/container`, `traces`
- **Environment Variables**: Reference via `${VAR_NAME}` in configs (e.g., `${OTEL_ENVOY_ADDRESS}`)
- **Processor Chains**: Group logically: filtering → transformation → attribute manipulation → batching
- **Filter Naming**: Descriptive suffixes: `filter/receiver`, `filter/remove_internal`, `filter/keep-entity-state-events`
- **Include Helpers**: Use `{{ include "common-config.filter-remove-internal" . | nindent 2 }}` for reusable processor blocks

### Python Testing Patterns
- **Naming**: `test_<feature>_<scenario>` for test functions
- **Setup/Teardown**: Use `setup_function()` and `teardown_function()` for test lifecycle
- **Assertions**: Clear helper functions like `assert_test_log_found(content)`, `print_failure(content)`
- **Retry Logic**: `retry_until_ok(url, func, print_failure, timeout=600)` pattern for eventual consistency
- **Type Hints**: Use typing module: `def get_all_bodies(log_bulk: dict) -> List[str]:`
- **Docstrings**: PEP 257 style with parameter and return documentation

### YAML Formatting
- **Indentation**: 2 spaces for YAML, never tabs
- **Line Length**: Keep under 120 characters
- **Key Naming**: camelCase for K8s resources, snake_case for OTEL configs
- **String Quoting**: Quote ambiguous values (versions, special chars)
- **Multi-line**: Use block scalars (`|`, `|-`) for readability

## Component Configuration Discovery

When working with OTEL components:
1. Check https://github.com/solarwinds/solarwinds-otel-collector-releases/blob/main/distributions/k8s/manifest.yaml for available components
2. Find component path from `gomod` line (e.g., `github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector`)
3. Navigate to component repo, read README.md for configuration
4. Reference component without suffix in config (use `solarwindsentity:` not `solarwindsentityconnector:`)

## Code Quality Standards

### Maintainability
- Small focused helper templates (match patterns in `_helpers.tpl`)
- Reusable configuration blocks in `_common-config.tpl`
- Clear comments explaining complex OTTL transformations
- Descriptive processor and pipeline names

### Testing
- Integration tests in Python verify actual telemetry collection
- Test against running mock services (`timeseries-mock-service`)
- Use `kubectl wait` patterns for pod readiness
- Collect pod logs on failure for debugging

### Documentation
- Inline comments for non-obvious OTEL transformations
- Section headers in YAML configs: `# ===== Receivers Configuration =====`
- Document regex patterns and complex conditionals
- Update CHANGELOG.md for breaking changes

## Backward Compatibility Requirements

1. **Never Remove Values**: Deprecate old values, maintain support with warnings
2. **Schema Validation**: Update `values.schema.json` with all value changes
3. **Feature Flags**: Use `.Values.feature.enabled` for new functionality
4. **API Versions**: Support multiple K8s API versions when needed
5. **Default Behavior**: Preserve existing default behaviors unless explicitly changing

## Adding New Code

- **New Processors**: Add to appropriate config file, create reusable helper in `_common-config.tpl` if used multiple times
- **New Pipelines**: Follow type-prefix naming convention, document in this file's architecture section
- **New Templates**: Use helper functions, apply standard labels/annotations
- **New Values**: Add to `values.yaml` with comments, document in `values.schema.json`, maintain backward compatibility
- **New Tests**: Follow setup/teardown pattern, use retry logic, create assertion helpers

## Patterns to Avoid

- **Hardcoded Names**: Use `{{ include "common.fullname" . }}` always
- **Version Assumptions**: Detect from values, don't assume specific K8s versions
- **Deep Template Nesting**: Refactor complex logic to helper templates
- **Duplicated Config**: Use include statements for repeated processor blocks
- **Missing Schema**: Every value must be in `values.schema.json`

## When In Doubt

1. Search for similar existing code in the repository
2. Check `.github/instructions/` for technology-specific guidance
3. Mirror the style and structure of surrounding code
4. Prioritize consistency with existing patterns over external best practices
5. Test against integration test suite to verify behavior  
