# swo-k8s-collector

Assets to monitor Kubernetes infrastructure using [SolarWinds Observability](https://documentation.solarwinds.com/en/success_center/observability/default.htm#cshid=gh-k8s-collector)

## Table of contents

- [About](#about)
- [Installation](#installation)
- [Limitations](#limitations)
- [Customization](#customization)

Development documentation: [development.md](doc/development.md)

## About

This repository contains:

- Source files for Helm chart `swo-k8s-collector`, used for collecting metrics, events and logs and exporting them to SolarWinds Observability platform.
- Dockerfile for an image published to Docker hub, that is deployed as part of Kubernetes monitoring
- All related sources that are built into that:
  - Custom OpenTelemetry Collector processors
  - OpenTelemetry Collector configuration

Components that are being deployed:

- Service account - identity of deployed pods
- Deployments - customized OpenTelemetry Collector deployments, configured to poll Kubernetes metrics and events
- ConfigMap - configuration of OpenTelemetry Collector
- DaemonSet - customized OpenTelemetry Collector deployment, configured to poll container logs and Prometheus metrics exposed by k8s workloads

## Installation

Walk through Add Kubernetes wizard in [SolarWinds Observability](https://documentation.solarwinds.com/en/success_center/observability/default.htm#cshid=gh-k8s-collector)

## Limitations

- Each Kubernetes version is supported for 15 months after its initial release. For example, version 1.27 released on April 11, 2023 is supported until July 11, 2024. For release dates for individual Kubernetes versions, see [Patch Releases](https://kubernetes.io/releases/patch-releases/#detailed-release-history-for-active-branches) in Kubernetes documentation.
  - Local Kubernetes deployments (e.q. Minikube, Docker Desktop) are not supported (although most of the functionality may be working).
  - Note: since Kubernetes v1.24 Docker container runtime will not be reporting pod level network metrics (`kubenet` and other network plumbing was removed from upstream as part of the dockershim removal/deprecation)
- Supported architectures: Linux x86-64 (`amd64`), Linux ARM (`arm64`), Windows x86-64 (`amd64`).

## Customization

The [Helm chart](deploy/helm/Chart.yaml) that you are about to deploy to your cluster has various configuration options. The full list, including the default settings, is available in [values.yaml](deploy/helm/values.yaml).

Internally, it contains [OpenTelemetry Collector configuration](https://opentelemetry.io/docs/collector/configuration/), which defines the metrics and logs to be monitored as well as their preprocessing.

**WARNING: Custom modifications to OpenTelemetry Collector configurations can lead to unexpected `swo-k8s-collector` behavior, data loss, and subsequent entity ingestion failures on the Solarwinds Observability platform side.**

### Mandatory configuration

In order to deploy the Helm chart, you need to prepare:

- A secret called `solarwinds-api-token` with API token for sending data to SolarWinds Observability
- A Helm chart configuration:

  ```yaml
  otel:
      endpoint: <solarwinds-observability-otel-endpoint>
  cluster:
      name: <cluster-display-name>
      uid: <unique-cluster-identifier>
  ```

### Metrics

By default, the `swo-k8s-collector` collects a subset of `kube-state-metrics` metrics and metrics exposed by workloads that are annotated with `prometheus.io/scrape: true`.

To configure the autodiscovery, see settings in the `otel.metrics.autodiscovery.prometheusEndpoints` section of the [values.yaml](deploy/helm/values.yaml).

Once deployed to a Kubernetes cluster, the metrics collection and processing configuration is stored as a ConfigMap under the `metrics.config` key.

In order to reduce the size of the collected data, the `swo-k8s-collector` collects only selected `kube-state-metrics` metrics that are key for successful entity ingestion on the SolarWinds Observability side. The list of metrics collected by default: [exported_metrics.md](doc/exported_metrics.md)

Native Kubernetes metrics are in a format that requires additional processing on the collector side to produce meaningful time series data that can later be consumed and displayed by the Solarwinds Observability platform.

### Logs

Once deployed to a Kubernetes cluster, the logs collection and processing configuration is stored as a ConfigMap under the `logs.config` key.

#### Version v3.x

The `swo-k8s-collector` collects container logs only in `kube-*` namespaces, which means it only collects logs from the internal Kubernetes container. This behavior can be modified by setting `otel.logs.filter` value. An example for scraping logs from all namespaces:

```yaml
otel:
  logs:
    filter:
      include:
        match_type: regexp
        record_attributes:
          - key: k8s.namespace.name
            value: ^.*$
```

#### Version v4.x

The `swo-k8s-collector` collects all logs by default which might be intensive. To avoid processing an excessive amount of data, the `swo-k8s-collector` can define filter which will drop all unwanted logs. This behavior can be modified by setting `otel.logs.filter` value.

An example for scraping container logs only from `kube-*` namespace:

```yaml
otel:
  logs:
    filter:
      log_record:
        - not(IsMatch(resource.attributes["k8s.namespace.name"], "^kube-.*$"))
```

An example for scraping container logs only from namespaces `my-custom-namespace` or `my-other-namespace` while excluding istio logs related to some of the successful HTTP requests:

```yaml
otel:
  logs:
    filter:
      log_record:
        - not(IsMatch(resource.attributes["k8s.namespace.name"], "(^my-custom-namespace$)|(^my-other-namespace$)"))
        - |
          resource.attributes["k8s.container.name"] == "istio-proxy" and
          IsMatch(body, "\\[[^\\]]*\\] \"\\S+ \\S+ HTTP/\\d(\\.\\d)*\" 200.*")
```

## Receive 3rd party metrics

SWO K8s Collector has an OTEL service endpoint which is able to forward metrics and logs into SolarWinds Observability. All incoming data is properly associated with current cluster. Additionally, metrics are decorated with prefix `k8s.`.

Service endpoint is provided in format

```text
"<chart-name>-metrics-collector.<namespace>.svc.cluster.local:4317"
```

### OpenTelemetry Collector configuration example

In case you want to send data from your own [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector-contrib) into SWO you can either send them directly into [public OTLP endpoint](https://documentation.solarwinds.com/en/success_center/observability/content/configure/configure-otel-directly.htm) or you can send them via our `swo-k8s-collector` to have better binding to other data available in SolarWinds Observability. To do that add following exporter into your configuration.

```yaml
config:
  exporters:
    otlp:
      endpoint: <chart-name>-metrics-collector.<namespace>.svc.cluster.local:4317
```

### Telegraf configuration example

[Telegraf](https://github.com/influxdata/telegraf) is a plugin-driven server agent used for collecting and reporting metrics.

Telegraf metrics can be sent into our endpoint by adding following fragment to your `values.yaml`

```yaml
config:
  outputs:
    - opentelemetry:
        service_address: <chart-name>-metrics-collector.<namespace>.svc.cluster.local:4317
```
