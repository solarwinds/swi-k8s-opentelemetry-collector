# swi-k8s-opentelemetry-collector

Assets to monitor kubernetes infrastructure in SolarWinds Observability

## Table of contents

- [About](#about)
- [Installation](#installation)
- [Development](#development)

## About

This repository contains:
* Kubernetes manifest files to collect metrics provided by existing Prometheus server, events and logs and export it to SolarWinds OaaS infrastructure.
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