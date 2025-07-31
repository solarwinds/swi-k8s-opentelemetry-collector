# Project snapshot
*Name*: SWO K8s Collector
*Domain*: Kubernetes monitoring and observability  
*Key technologies*: OpenTelemetry, Helm, Kubernetes, Python, Prometheus, eBPF

# Testing rules
- Use Helm unit tests for verifying chart rendering.  
- Update snapshot tests with `helm unittest -u deploy/helm`.
- To run integration test you DO NOT CHECK for its configuration, DO NOT VALIDATE anything else, you simply run these commands in order:
1. Build artifacts: `skaffold build --file-output=/tmp/tags.json -v info`
  * This is needed for the first time and every time you change the integration test.
2. Deploy environment: `skaffold deploy --build-artifacts /tmp/tags.json -p beyla,operator,no-prometheus --status-check=true -v info`
  * This is needed for the first time and any time there are configuration changes. This is not needed if only integration test is changed. 
3. Run tests: `skaffold verify --build-artifacts /tmp/tags.json -v info`
4. Delete environment: `skaffold delete -v info`
  * This will cleanup everything what `skaffold deploy` created.

You must wait for all the commands that you run, avoid running them in the background.

# Directory layout
- **deploy/helm/** – Helm chart for deploying the collector
  - **templates/metrics-deployment.yaml** – MetricsCollector deployment definition
  - **templates/metrics-discovery-deployment.yaml** – MetricsDiscovery deployment definition
  - **templates/events-collector-statefulset.yaml** – EventsCollector statefulset definition
  - **templates/node-collector-daemon-set.yaml** – NodeCollector daemonset definition
  - **templates/node-collector-daemon-set-windows.yaml** – NodeCollector Windows daemonset definition
  - **templates/gateway/** – Gateway collector components
  - **templates/network/** – eBPF network monitoring components
  - **templates/beyla/** – Beyla eBPF-based auto-instrumentation
  - **templates/operator/** – OpenTelemetry operator integration
  - **templates/autoupdate/** – Components for auto-updating configurations
  - **templates/openshift/** – OpenShift-specific components
  - **metrics-collector-config.yaml** – MetricsCollector pipeline configuration
  - **metrics-discovery-config.yaml** – MetricsDiscovery pipeline configuration for discovered metrics
  - **events-collector-config.yaml** – EventsCollector pipeline configuration
  - **gateway-collector-config.yaml** – Gateway collector pipeline configuration
  - **node-collector-config.yaml** – NodeCollector pipeline configuration
  - **values.yaml** – Helm chart default values
  - **values.schema.json** – JSON schema for validating Helm values
- **doc/** – Documentation including development and metrics info
  - **collectorPipeline.md** – Description of data flow between components
  - **development.md** – Development guide
  - **exported_metrics.md** – List of metrics exported by the collector
- **tests/integration/** – Integration tests for verifying telemetry collection  
- **operator/** – OpenTelemetry operator customization
- **utils/** – Utility scripts for development and testing  

# OTEL Collector Pipelines
The SWO K8s Collector consists of five main components, each with specific pipelines:

## MetricsCollector Deployment
- **metrics pipeline**: Main pipeline that exports metrics via OTLP
- **metrics/kubestatemetrics pipeline**: Collects metrics from the Kube State Metrics service
- **metrics/otlp pipeline**: Receives metrics via OTLP protocol and forwards them to the main metrics pipeline
- **metrics/prometheus pipeline**: Processes Prometheus-formatted metrics and forwards them to the main metrics pipeline
- **metrics/prometheus-node-metrics pipeline**: Collects node-level metrics in Prometheus format
- **metrics/prometheus-server pipeline**: Collects metrics from the Prometheus server

## MetricsDiscovery Deployment
- **metrics/discovery pipeline**: Discovers and collects metrics from annotated pods (especially in AWS Fargate) using the receiver_creator and k8s_observer
- **metrics pipeline**: Processes discovered metrics and exports them via OTLP

## EventsCollector Deployment
- **logs pipeline**: Collects Kubernetes events (pod creations, deletions, etc.) via the k8s_events receiver
- **logs/manifests pipeline**: Collects Kubernetes object manifests via the k8sobjects receiver

## NodeCollector DaemonSet
- **logs pipeline**: Main pipeline for logs that exports them via OTLP
- **logs/container pipeline**: Collects container logs from files using the filelog receiver
- **logs/journal pipeline**: Collects system logs from journald
- **metrics pipeline**: Main pipeline for node-level metrics that exports them via OTLP
- **metrics/discovery pipeline**: Uses receiver_creator to discover and collect metrics from discoverable endpoints
- **metrics/node pipeline**: Collects metrics specific to the node using receiver_creator

## Gateway Collector Deployment
- **traces pipeline**: Receives traces via OTLP protocol and exports them
- **metrics pipeline**: Receives metrics via OTLP protocol and exports them
- **logs pipeline**: Receives logs via OTLP protocol and exports them

# Component configuration
- OTEL uses SolarWinds built OTEL Collector
- Use `github` MCP server to access GitHub related content
- You will find all the components available in this file https://github.com/solarwinds/solarwinds-otel-collector-releases/blob/main/distributions/k8s/manifest.yaml
    - Location of component configuration can be infered from the `gomod` line
    - For example if you find `  - gomod: github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector v0.123.7` you will look for configuration settings into `https://github.com/solarwinds/solarwinds-otel-collector-contrib/tree/main/connector/solarwindsentityconnector`. 
    - You look first into README.md, if that does not contain configuration sample or example you have to look at the code of the component. 
- Generally the component is referenced in the configuration typically without receiver/exporter/processor/connector suffux (e.g. in configuration use `solarwindsentity` key instead of `solarwindsentityconnector`)

# Coding conventions 
1. Maintain backward compatibility when updating chart configurations.  
2. Follow JSON schema validation for Helm values through `values.schema.json`.  
