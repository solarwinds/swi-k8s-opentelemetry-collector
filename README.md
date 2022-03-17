# nighthawk-im-k8s-monitor

Assets to monitor kubernetes infrastructure

## Table of contents

- [About](#about)
- [Installation](#installation)

## About

This repository contains Kubernetes manifest files to collect metrics provided by existing Prometheus server, and export those metrics to SolarWinds OaaS infrastructure.
Components that are being deployed:

- Service account - identity of deployed pods
- Deployment - customized OpenTelemetry Collector deployment, configured to poll Prometheus instance(s)
- ConfigMap - configuration of OpenTelemetry Collector

## Installation

1. First decide to which namespace you want to deploy the manifest. It is recommended to deploy them to the same namespace where you Prometheus instance is deployed.
2. Store API Token to kubernetes secret called `solarwinds-api-token` (Get the token from `Settings` -> `API Tokens` -> `Create API Token` and select `Ingestion` Type)

```bash
kubectl create secret generic solarwinds-api-token -n <CHOSEN NAMESPACE> --from-literal=SOLARWINDS_API_TOKEN=<REPLACE WITH TOKEN>
```

3. Adjust Prometheus instance(s) in the manifest (look for `PROMETHEUS_URL` in the manifest or in case of multiple instances adjust OtelCollector configuration in `receivers` -> `prometheus` -> `config` -> `scrape_configs` -> `job_name: prometheus` -> `static_configs` -> `targets`)
4. Set current active namespace to namespace from step 1

```
kubectl config set-context --current --namespace=<insert-namespace-name-here>
```

5. Deploy the manifest

```
kubectl apply -f deploy/k8s/manifest.yaml
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

## Publishing
customized Otel Collector image is getting published to https://hub.docker.com/repository/docker/solarwinds/swi-opentelemetry-collector 

Steps to publish new version:
* Create GitHub release selecting the Tag/branch you want to release with description of changes
  * use tag in semver format, it is the tag which Docker hub image will have publicly
  * publish release
* GitHub action will be triggered that will build the release and wait for publish approval
* after CODEOWNERS approve it, it will be published to Dockerhub public repository