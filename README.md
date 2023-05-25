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

- Source files for Helm chart `swo-k8s-collector`, used for collecting metrics (provided by existing Prometheus server), events and logs and exporting them to SolarWinds Observability platform.
- Dockerfile for an image published to Docker hub, that is deployed as part of Kubernetes monitoring
- All related sources that are built into that:
  - Custom OpenTelemetry Collector processors  
  - OpenTelemetry Collector configuration

Components that are being deployed:

- Service account - identity of deployed pods
- Deployment - customized OpenTelemetry Collector deployment, configured to poll Prometheus instance(s)
- ConfigMap - configuration of OpenTelemetry Collector
- DaemonSet - customized OpenTelemetry Collector deployment, configured to poll container logs

## Installation

Walk through Add Kubernetes wizard in [SolarWinds Observability](https://documentation.solarwinds.com/en/success_center/observability/default.htm#cshid=gh-k8s-collector)

## Limitations

- Supported Kubernetes version: 1.18 and higher.
  - Local Kubernetes deployments (e.q. Minikube) are not supported (although most of the functionality may be working).
  - Note: since Kubernetes v1.24 Docker container runtime will not be reporting pod level network metrics (`kubenet` and other network plumbing was removed from upstream as part of the dockershim removal/deprecation)
- Supported architectures: Linux x86-64 (`amd64`).

## Customization

The [Helm chart](deploy/helm/Chart.yaml) that you are about to deploy to your cluster has various configuration options. The full list, including the default settings, is available in [values.yaml](deploy/helm/values.yaml).

Internally, it contains [OpenTelemetry Collector configuration](https://opentelemetry.io/docs/collector/configuration/), which defines the metrics and logs to be monitored as well as their preprocessing.

**WARNING: Custom modifications to OpenTelemetry Collector configurations can lead to unexpected `swo-k8s-collector` behavior, data loss, and subsequent entity ingestion failures on the Solarwinds Observability platform side.**

### Metrics

The `swo-k8s-collector` collects metrics from a Prometheus instance. To configure its address, set

```yaml
otel:
  metrics:
    prometheus:
      url: <some_address>
```

Alternatively, **for testing purposes**, you can also let the collector deploy a Prometheus server for you.

```yaml
prometheus:
  enabled: true
```

Once deployed to a Kubernetes cluster, the metrics collection and processing configuration is stored as a ConfigMap under the `metrics.config` key.

In order to reduce the size of the collected data, the `swo-k8s-collector` collects only selected metrics that are key for successful entity ingestion on the SolarWinds Observability side. The list of observed metrics can be extended by setting `otel.metrics.extra_scrape_metrics` value. Example:

```yaml
otel:
  metrics:
    extra_scrape_metrics:
      - node_cpu_seconds_total
      - node_cpu_guest_seconds_total
```

The list of metrics collected by default: [exported_metrics.md](doc/exported_metrics.md)

Native Kubernetes metrics are in a format that requires additional processing on the collector side to produce meaningful time series data that can later be consumed and displayed by the Solarwinds Observability platform.

Processors included in the collector:

- [attributes](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/attributesprocessor)
- [batch](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/batchprocessor)
- [cumulativetodelta](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/cumulativetodeltaprocessor)
- [deltatorate](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/deltatorateprocessor)
- [experimental_metricsgeneration](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/metricsgenerationprocessor)
- [filter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/filterprocessor)
- [groupbyattrs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/groupbyattrsprocessor)
- [memory_limiter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/memorylimiterprocessor)
- [metricstransform](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/metricstransformprocessor)
- [prometheustypeconvert](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/tree/master/src/processor/prometheustypeconverterprocessor)
- [resource](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/resourceprocessor)
- [transform](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/transformprocessor)
- [swmetricstransform](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/tree/master/src/processor/swmetricstransformprocessor)

### Logs

Once deployed to a Kubernetes cluster, the logs collection and processing configuration is stored as a ConfigMap under the `logs.config` key.

To avoid processing an excessive amount of data, the `swo-k8s-collector` collects container logs only in `kube-*` namespaces, which means it only collects logs from the internal Kubernetes container. This behavior can be modified by setting `otel.logs.filter` value. An example for scraping logs from all namespaces:

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
