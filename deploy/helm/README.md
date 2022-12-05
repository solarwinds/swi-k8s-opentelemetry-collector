# SolarWinds Observability - Kubernetes integration

SolarWinds Observability is a full-stack observability product. The Kubernetes integration installs the OpenTelemetry
Collector to retrieve metrics from an existing Prometheus installation. To get started quickly with a basic Prometheus
installation, you can use the [Prometheus community helm chart](https://github.com/prometheus-community/helm-charts/tree/main/charts/prometheus)
to get a running, compatible version of Prometheus.

## Prerequisites

* Kubernetes 1.18+
* Prometheus 0.15.0+
    * With kube-state-metrics 1.5.0+
* amd64 architecture
