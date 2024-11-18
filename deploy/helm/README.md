# swo-k8s-collector

## Table of contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Limitations](#limitations)
- [Auto Instrumentation (experimental feature)](#auto-instrumentation-experimental-feature)

## Installation

Walk through `Add a Kubernetes cluster` in [SolarWinds Observability](https://documentation.solarwinds.com/en/success_center/observability/default.htm#cshid=gh-k8s-collector)

## Configuration

The [Helm chart](Chart.yaml) that you are about to deploy to your cluster has various configuration options. The full list, including the default settings, is available in [values.yaml](values.yaml).

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

#### Version 4.1.0 and newer

Starting with version 4.1.0 setting `cluster.uid` is optional. If not provided it defaults to value of `cluster.name`.

### Metrics

By default, the `swo-k8s-collector` collects a subset of `kube-state-metrics` metrics and metrics exposed by workloads that are annotated with `prometheus.io/scrape: true`.

To configure the autodiscovery, see settings in the `otel.metrics.autodiscovery.prometheusEndpoints` section of the [values.yaml](values.yaml).

Once deployed to a Kubernetes cluster, the metrics collection and processing configuration is stored as a ConfigMap under the `metrics.config` key.

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

### Manifests

Starting with version 4.0.0, `swo-k8s-collector` observes changes in supported resources and collects their manifests.

By default, manifest collection is enabled, but it can be disabled by setting `otel.manifests.enabled` to `false`.  Manifest collection runs in the event collector, so `otel.events.enabled` must be set to `true` (default). 

Currently, the following resources are watched for changes: `pods`, `deployments`, `statefulsets`, `replicasets`, `daemonsets`, `jobs`, `cronjobs`, `nodes`, `services`, `persistentvolumes`, `persistentvolumeclaims`, `configmaps`, `ingresses` and Istio's `virtualservices`.

By default, `swo-k8s-collector` collects all manifests. You can use the `otel.manifests.filter` setting to filter out manifests that should not be collected.

An example of filter for collecting all manifests, but `configmaps` just for `kube-system` namespace.

```yaml
otel:
  manifests:
    enabled: true
    filter:
      log_record:  
        - attributes["k8s.object.kind"] == "ConfigMap" and resource.attributes["k8s.namespace.name"] != "kube-system"
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

## Limitations

- Each Kubernetes version is supported for 15 months after its initial release. For example, version 1.27 released on April 11, 2023 is supported until July 11, 2024. For release dates for individual Kubernetes versions, see [Patch Releases](https://kubernetes.io/releases/patch-releases/#detailed-release-history-for-active-branches) in Kubernetes documentation.
  - Local Kubernetes deployments (e.q. Minikube, Docker Desktop) are not supported (although most of the functionality may be working).
  - Note: since Kubernetes v1.24 Docker container runtime will not be reporting pod level network metrics (`kubenet` and other network plumbing was removed from upstream as part of the dockershim removal/deprecation)
- Supported architectures: Linux x86-64 (`amd64`), Linux ARM (`arm64`), Windows x86-64 (`amd64`).

## Auto Instrumentation (experimental feature)

This chart allows you to deploy the [OpenTelemetry Operator](https://github.com/open-telemetry/opentelemetry-operator), which can be used to auto-instrument applications with [SWO APM](https://documentation.solarwinds.com/en/success_center/observability/content/intro/services.htm).

### Setting up

#### 1. Enable deployment of the operator

Set the following option in `values.yaml`: `operator.enable=true`

#### 2. Ensure proper TLS Certificate management

The operator expects that Cert Manager is already present on the cluster. There are a few different ways you can use to generate/configure the required TLS certificate:

- Deploy `cert-manager` as part of this chart.
  - Ensure there is no cert-manager instance already present in the cluster.
  - Set `certmanager.enabled=true`.
- Or, read the OTEL Operator documentation for alternative options: https://opentelemetry.io/docs/kubernetes/helm/operator/#configuration. All OTEL Operator configuration options are available below the `operator` key in `values.yaml`.

#### 3. Create an `Instrumentation` custom resource

- Create an `Instrumentation` custom resource.
- Set `SW_APM_SERVICE_KEY` with the SWO ingestion API token. You can use the same token that is used for this chart.
- Set `SW_APM_COLLECTOR` with the APM SWO endpoint (e.g., `apm.collector.na-01.cloud.solarwinds.com`).
- If you are using a secret to store the API token, both the secret and the `Instrumentation` resource must be created in the same namespace as the applications that are to be instrumented.
- Example:

  ```yaml
  apiVersion: opentelemetry.io/v1alpha1
  kind: Instrumentation
  metadata:
    name: swo-apm-instrumentation
  spec:
    java:
      env:
        - name: SW_APM_SERVICE_KEY
          valueFrom:
            secretKeyRef:
              name: solarwinds-api-token
              key: SOLARWINDS_API_TOKEN
        - name: SW_APM_COLLECTOR
          value: apm.collector.na-01.cloud.solarwinds.com
  ```

#### 4. Instrument applications by setting the annotation

The final step is to opt your services into automatic instrumentation. This is done by updating your service’s `spec.template.metadata.annotations` to include a language-specific annotation:

- .NET: `instrumentation.opentelemetry.io/inject-dotnet: "true"`
- Go: `instrumentation.opentelemetry.io/inject-go: "true"`
- Java: `instrumentation.opentelemetry.io/inject-java: "true"`
- Node.js: `instrumentation.opentelemetry.io/inject-nodejs: "true"`
- Python: `instrumentation.opentelemetry.io/inject-python: "true"`

The possible values for the annotation can be:

- `"true"` - to inject the Instrumentation resource with the default name from the current namespace.
- `"my-instrumentation"` - to inject the Instrumentation CR instance with the name "my-instrumentation" in the current namespace.
- `"my-other-namespace/my-instrumentation"` - to inject the Instrumentation CR instance with the name "my-instrumentation" from another namespace "my-other-namespace".
- `"false"` - do not inject.

Alternatively, the annotation can be added to a namespace, which will result in all services in that namespace opting into automatic instrumentation. See the [Operator's auto-instrumentation documentation](https://github.com/open-telemetry/opentelemetry-operator/blob/main/README.md#opentelemetry-auto-instrumentation-injection) for more details.
