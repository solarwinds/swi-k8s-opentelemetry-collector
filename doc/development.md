# Development

## Table of contents

- [Contribution Guidelines](#contribution-guidelines)
- [Prerequisites](#prerequisites)
- [Deployment](#deployment)
- [Develop against remote prometheus](#develop-against-remote-prometheus)
- [Helm Unit tests](#helm-unit-tests)
- [Integration tests](#integration-tests)
- [Updating Chart dependencies](#updating-chart-dependencies)
- [Updating Chart configuration](#updating-chart-configuration)
- [Release](#release)

## Contribution Guidelines

1. Pull the latest changes from the `master` branch.
2. Create your new (local) work branch.
3. Work on/commit your changes, see Development bellow.
4. Add an entry to the `Unreleased` section of `deploy/helm/CHANGELOG.md` describing your changes.
5. Push your branch to GitHub and create a Pull Request.
6. Once approved, the Pull Request can be merged with the `master` branch.

## Prerequisites

- [Skaffold](https://skaffold.dev) at least [v2.0.3](https://github.com/GoogleContainerTools/skaffold/releases/tag/v2.0.3)
  - On windows, do not install it using choco due to [this issue](https://github.com/GoogleContainerTools/skaffold/issues/4058)
- [Kustomize](https://kustomize.io):

  ```shell
  choco install kustomize
  ```

- [Helm](https://helm.sh):

  ```shell
  choco install kubernetes-helm
  ```

- Prometheus community Helm repo - it hosts a dependency for the collector's chart:

  ```shell
  helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
  ```

- [Docker desktop](https://www.docker.com/products/docker-desktop) with Kubernetes enabled

## Deployment

To run the collector in a local environment, execute:

```shell
skaffold dev
```

That will:

- build customized Otel Collector image
- deploy Prometheus
- deploy OtelEndpoint mock (to see that customized Otel Collector is sending metrics correctly)
- deploy customized Otel Collector

Possible issues:

- if you get errors like:

  ```text
  Error: INSTALLATION FAILED: failed to download https://github.com/prometheus-community/helm-charts...
  ```

  or

  ```text
  Error: INSTALLATION FAILED: no cached repo found. (try 'helm repo update'): open C:\Users\<user>\AppData\Local\Temp\helm\repository\stable-index.yaml: The system cannot find the file specified.
  ```

  you need to update Helm repo:

  ```shell
  helm repo update
  ```

- if you get error like

  ```text
  ...Unable to get an update from the "stable" chart repository (https://kubernetes-charts.storage.googleapis.com/):
          failed to fetch https://kubernetes-charts.storage.googleapis.com/index.yaml : 403 Forbidden
  ```

  you need to update path to a Helm repository:

  ```shell
  helm repo add "stable" "https://charts.helm.sh/stable" --force-update
  ```

### How can you analyze exported telemetry

#### Metrics

- You can look at `http://localhost:8088/metrics.json` (each line is JSON as bulk sent by OTEL collector)
- You can also look at local Prometheus which collects all the outputs with metric names prefixed with `output_` at `http://localhost:8080`

#### Logs

You can look at `http://localhost:8088/logs.json` (each line is JSON as bulk sent by OTEL collector)

#### Events

You can look at `http://localhost:8088/events.json` (each line is JSON as bulk sent by OTEL collector)

#### Manifests

You can look at `http://localhost:8088/manifests.json` (each line is JSON as bulk sent by OTEL collector)

#### Entity State events

You can look at `http://localhost:8088/entitystateevents.json` (each line is JSON as bulk sent by OTEL collector).

## Develop against remote cluster
* Make sure that you have working kubeContext, set this to `test-cluster` profile section in `skaffold.yaml`:
```
    - op: replace
      path: /deploy/kubeContext
      value: "<your kube context here>"
```
* Create `skaffold.env` file with two environment variables which indicates your private space in the cluster, e.q.:
```
TEST_CLUSTER_NAMESPACE=my-dev-namespace
TEST_CLUSTER_RELEASE_NAME=swi-my-dev-release
```
* Make sure that you have ECR repository where image `swi-opentelemetry-collector` will be pushed
* Login to ECR repository. E.q.: `aws ecr get-login-password --region us-east-1 --profile <your AWS profile> | docker login --username AWS --password-stdin <Your ECR repository>`
* (optionally) setup skaffold values to better target what you intent to develop. E.q.:
```
ebpfNetworkMonitoring.enabled: true
otel.logs.enabled: false
otel.events.enabled: false
otel.metrics.batch.send_batch_size: 1024
otel.metrics.batch.send_batch_max_size: 1024
```
* Run `skaffold dev -p=test-cluster --default-repo=<Your ECR repository>`

## Develop against remote prometheus

You can port forward Prometheus server to localhost:9090 and run

```shell
skaffold dev -p=remote-prometheus
```

In order to change Prometheus endpoint that is hosted on HTTPS you can adjust skaffold.yaml file:

- add `otel.metrics.prometheus.scheme: https`
- update `otel.metrics.prometheus.url: <remote prometheus>`

## Helm Unit tests

Helm Unit tests are located in `deploy/helm/tests` and are supposed to verify how Helm chart is rendered.

### Setup

Run in bash (or Git Bash):

```shell
helm plugin install https://github.com/helm-unittest/helm-unittest.git
```

### Run tests locally

```shell
helm unittest deploy/helm
```

### Refresh snapshot tests

```shell
helm unittest -u deploy/helm
```

### Integration with VS Code

To enable code completion when writing new tests, install a VS Code extension providing a YAML Language server, like `redhat.vscode-yaml`.

## Integration tests
Integration tests are located in `tests/integration` and are supposed to verify if metric processing is delivering expected outcome.

### Prerequisites
Deploy cluster locally using `skaffold dev`
### Run tests locally
* Install all dependencies: `pip install --user -r tests/integration/requirements.txt` 
* Can be run in Visual Studio Code by opening individual tests and run `Python: Pytest` debug configuration
* You can run it directly in cluster by manually triggering `integration-test` CronJob

### Updating utils used for testing

Whenever there is a need to improve the test tooling, eg. the script for scraping test data from a Prometheus (`utils/cleanup_mocked_prometheus_response.py`), or data comparison code, or versions or Python packages, ..., it should always happen in a separate PR. Do not mix changes to the test framework with changes to the k8s collector itself. Otherwise a change to the testing framework might hide an unintentional change to the collector code.

## Updating Chart dependencies

To update a dependency of the Helm chart:

1. Update the `dependencies` section in [deploy/helm/Chart.yaml](../deploy/helm/Chart.yaml).
2. Run

    ```shell
    helm dependency update deploy/helm
    ```

3. Commit changes in [deploy/helm/Chart.lock](../deploy/helm/Chart.lock).
4. *(Optional)* Delete `*.tgz` files in `/deploy/helm/charts/` - they will be re-downloaded automatically as needed.

## Updating Chart configuration

First and foremost, any changes to the default `values.yaml` or how the configuration required by the Helm templates must be backwards compatible. Breaking existing configurations prepared for previous versions of the software should be avoided, if possible.

The Helm chart contains a JSON schema for the validation of the provided configuration [values.schema.json](../deploy/helm/values.schema.json).

To use it during development, reference it by your YAML parser. For example, for software that supports the language server protocol, add `# yaml-language-server: $schema=values.schema.json` as a first line in your `values.yaml` file, adjusting the path (local, or URL) accordingly.

To verify that the changes are compatible with the current schema, it's suggested to run also:

```shell
helm lint -f <values_yaml_for_testing> .\deploy\helm\ --with-subcharts
```

Basic linting is part of the build pipeline, though.

The Helm chart is bundled also in AKS/EKS addons. Make sure that any changes are reflected there, too.

## Release

### Docker image

1. Create tag you want to release and push it to origin

    ```shell
    git tag 0.11.5
    git push origin 0.11.5
    ```

1. GitHub Action will be triggered, building the release and awaiting manual approval for publishing.
1. Once approved, it will be published to Docker Hub repository: [solarwinds/swi-opentelemetry-collector](https://hub.docker.com/repository/docker/solarwinds/swi-opentelemetry-collector).

### Helm Chart

1. Create PR with version change into [Chart.yaml](../deploy/helm/Chart.yaml)
1. Once the PR is merged, GitHub Action will be triggered to build the release and open a PR to the `gh-pages` branch.
1. Review the PR created for changes to the `gh-pages` branch (which hosts the Helm charts), and merge it.
1. Once the PR is merged, the Helm chart is published to <https://helm.solarwinds.com>.
