# swi-k8s-opentelemetry-collector

Assets to monitor kubernetes infrastructure in SolarWinds Observability

## Table of contents

- [About](#about)
- [Installation](#installation)
- [Development](#development)

## About

This repository contains:
* Kubernetes manifest files to collect metrics provided by existing Prometheus server, events and logs and export it to SolarWinds Observability infrastructure.
* Dockerfile for images published to Docker hub that is deployed as part of Kubernetes monitoring
* All related sources that are built into that:
  * Custom OpenTelemetry collector processors  
  * OpenTelemetry collector configuration

Components that are being deployed:

- Service account - identity of deployed pods
- Deployment - customized OpenTelemetry Collector deployment, configured to poll Prometheus instance(s)
- ConfigMap - configuration of OpenTelemetry Collector
- DaemonSet - customized OpenTelemetry Collector deployment, configured to poll container logs

## Installation
Walk through Add Kubernetes wizard in SolarWinds Observability

## Limitations
* Supported kubernetes version: 1.18 and higher.
  * Local kubernetes deployments (e.q. Minikube) are not supported (although most of the functionality may be working).
* Supported kube-state-metrics: 1.5.0 and higher.
* Supported architectures: Linux x86-64 (`amd64`).

## Customization
The [manifest](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/blob/master/deploy/k8s/manifest.yaml) that you are about to deploy to your cluster using the Add Kubernetes wizard contains [OpenTelemetry configuration](https://opentelemetry.io/docs/collector/configuration/) which defines the metrics and logs to be monitored. It allows you to customize the list of metrics and logs to be monitored, as well as their preprocessing.

**WARNING: Custom modifications to OpenTelemetry collector configurations can lead to unexpected Kubernetes agent behavior, data loss, and subsequent entity ingestion failures on the Solarwinds Observability platform side.**

### Metrics

The metrics collection and processing configuration is included in the manifest as a ConfigMap under the `metrics.config` key.

In order to reduce the size of the collected data, the swi-k8s-opentelemetry-collector whitelists only selected metrics that are key for successful entity ingestion on the Solarwinds Observability side. The list of observed metrics can be easily modified by simply adding or removing the desired metrics from the list located in the `scrape_configs` section of the collector configuration.

Default metrics monitored by swi-k8s-opentelemetry-collector:
- container_cpu_usage_seconds_total
- container_spec_cpu_quota
- container_spec_cpu_period
- container_memory_working_set_bytes
- container_spec_memory_limit_bytes
- container_cpu_cfs_throttled_periods_total
- container_cpu_cfs_periods_total
- kube_node_info
- kube_node_created
- kube_node_status_capacity
- kube_node_status_condition
- kube_pod_created
- kube_pod_info
- kube_pod_start_time
- kube_pod_completion_time
- kube_pod_status_phase
- kube_pod_start_time
- kube_resourcequota
- kube_pod_container_status_restarts_total
- kube_node_status_allocatable
- kube_pod_container_resource_requests
- '{\_\_name\_\_=~"kube_pod_container_.*"}'

Native Kubernetes metrics are in a format that requires additional processing on the collector side to produce meaningful time series data that can later be consumed and displayed by the Solarwinds Observability platform. 

Processors included in the collector:
  - [batch](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/batchprocessor)
  - [cumulativetodelta](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/cumulativetodeltaprocessor)
  - [deltatorate](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/deltatorateprocessor)
  - [experimental_metricsgeneration](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/metricsgenerationprocessor)
  - [groupbyattrs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/groupbyattrsprocessor)
  - [memory_limiter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/memorylimiterprocessor)
  - [metricstransform](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/metricstransformprocessor)
  - [prometheustypeconvert](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/tree/master/src/processor/prometheustypeconverterprocessor)
  - [resource](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/resourceprocessor)
  - [swmetricstransform](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/tree/master/src/processor/swmetricstransformprocessor)

### Logs

The logs collection and processing configuration is included in the manifest as a ConfigMap under the `logs.config` key.

To reduce the overall size of the data created during log collection, the collector whitelists container logs only on `kube-*` namespaces, which means it only collects logs from the internal Kubernetes container. Otherwise, the size of the collected data would lead to infrastructure overload. This behavior can be modified in the `filter` section of the log collection configuration.

To collect all logs remove the `filter` section
```diff
processors:
  # For more all the options about the filtering see https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/filterprocessor
- filter:
-   logs:
-     include:
-         match_type: regexp
-         record_attributes:
-             # allow only system namespaces (kube-system, kube-public)
-             - key: k8s.namespace.name
-               value: ^kube-.*$
```

To collect logs for a specific namespace, change the filter value with the name of the namespace to be monitored
```diff
filter:
  logs:
    include:
        match_type: regexp
        record_attributes:
            # allow only system namespaces (kube-system, kube-public)
            - key: k8s.namespace.name
-              value: ^kube-.*$
+              value: <NAMESPACE_NAME>
```

## Development

### Prerequisites

- [Skaffold](https://skaffold.dev) at least [v1.31.0](https://github.com/GoogleContainerTools/skaffold/releases/tag/v1.31.0)
  - On windows, do not install it using choco due to [this issue](https://github.com/GoogleContainerTools/skaffold/issues/4058)
- [Kustomize](https://kustomize.io): `choco install kustomize`
- [Helm](https://helm.sh): `choco install kubernetes-helm`
- [Docker desktop](https://www.docker.com/products/docker-desktop) with Kubernetes enabled

### Deployment

To run local environment run: `skaffold dev` command.

That will:

- build customized Otel Collector image
- deploy Prometheus
- deploy OtelEndpoint mock (to see that customized Otel Collector is sending metrics correctly)
- deploy customized Otel Collector

Possible issues:

- if you get error like `Error: INSTALLATION FAILED: failed to download https://github.com/prometheus-community/helm-charts...`, you need to update helm repo: `helm repo update`
- if you get error like

  ```text
  ...Unable to get an update from the "stable" chart repository (https://kubernetes-charts.storage.googleapis.com/):
          failed to fetch https://kubernetes-charts.storage.googleapis.com/index.yaml : 403 Forbidden
  ```

  you need to update path to a helm repository:

  ```shell
  helm repo add "stable" "https://charts.helm.sh/stable" --force-update
  ```

## Publishing
customized Otel Collector image is getting published to https://hub.docker.com/repository/docker/solarwinds/swi-opentelemetry-collector 

Steps to publish new version:
* Create GitHub release selecting the Tag/branch you want to release with description of changes
  * use tag in semver format, it is the tag which Docker hub image will have publicly
  * publish release
* GitHub action will be triggered that will build the release and wait for publish approval
* after CODEOWNERS approve it, it will be published to Dockerhub public repository