# nighthawk-im-k8s-monitor

Assets to monitor kubernetes infrastructure

## Table of contents

  * [About](#about)
  * [Installation](#installation)

## About

This repository contains Kubernetes manifest files to collect metrics provided by existing Prometheus server, and export those metrics to SolarWinds OaaS infrastructure. 
Components that are being deployed:
* Service account - identity of deployed pods
* Deployment - customized OpenTelemetry Collector deployment, configured to poll Prometheus instance(s)
* ConfigMap - configuration of OpenTelemetry Collector 

## Installation

1. First decide to which namespace you want to deploy the manifest. It is recommended to deploy them to the same namespace where you Prometheus instance is deployed. 
2. Store API Token to kubernetes secret called `solarwinds-api-token` (Get the token from `Settings` -> `API Tokens` -> `Create API Token` and select `Ingestion` Type)
``` bash
kubectl create secret generic solarwinds-api-token -n <CHOSEN NAMESPACE> --from-literal=SOLARWINDS_API_TOKEN=<REPLACE WITH TOKEN>
```
3. Adjust Prometheus instance(s) in the manifest (look for `PROMETHEUS_URL` in the manifest or in case of multiple instances adjust OtelCollector configuration in `receivers` -> `prometheus` -> `config` -> `scrape_configs` -> `job_name: prometheus` -> `static_configs` -> `targets`)
4. Deploy the manifest
```
kubectl apply -f deploy/k8s/manifest.yaml
```
