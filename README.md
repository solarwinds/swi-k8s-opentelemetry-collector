# SolarWinds Observability Kubernetes Collector

Assets to monitor Kubernetes infrastructure using [SolarWinds Observability](https://documentation.solarwinds.com/en/success_center/observability/default.htm#cshid=gh-k8s-collector)


## About

This repository contains:

- Source files for [Helm chart](deploy/helm/README.md) `swo-k8s-collector`, used for collecting metrics, events and logs and exporting them to SolarWinds Observability platform.
- All related sources that are built into that:
  - OpenTelemetry Collector configuration


## Installation

Walk through `Add a Kubernetes cluster` in [SolarWinds Observability](https://documentation.solarwinds.com/en/success_center/observability/default.htm#cshid=gh-k8s-collector)


## Contibutions

Development documentation: [development.md](doc/development.md)

## License

The SolarWinds Kubernetes Collector is licensed under the [Apache License, Version 2.0](LICENSE).
