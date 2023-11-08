# Development

## Table of contents

- [Contribution Guidelines](#contribution-guidelines)
- [Prerequisites](#prerequisites)
- [Deployment](#deployment)
- [Develop against remote prometheus](#develop-against-remote-prometheus)
- [Integration tests](#integration-tests)
- [Updating Chart dependencies](#updating-chart-dependencies)
- [Publishing](#publishing)

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
* You can look at `http://localhost:8088/metrics.json` (each line is JSON as bulk sent by OTEL collector)
* You can also look at local Prometheus which collects all the outputs with metric names prefixed with `output_` at `http://localhost:8080`

#### Logs
You can look at `http://localhost:8088/logs.json` (each line is JSON as bulk sent by OTEL collector)

#### Events
You can look at `http://localhost:8088/events.json` (each line is JSON as bulk sent by OTEL collector)

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

## Integration tests
Integration tests are located in `tests/integration` and are supposed to verify if metric processing is delivering expected outcome.

### Prerequisites
Deploy cluster locally using `skaffold dev -p=only-mock` (configured to poll mocked data)
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

## Publishing

### Docker image

Customized Otel Collector image is getting published to <https://hub.docker.com/repository/docker/solarwinds/swi-opentelemetry-collector>.

Steps to publish new version:

1. Create GitHub release selecting the Tag/branch you want to release with description of changes
   - use tag in semver format, it is the tag which Docker hub image will have publicly
   - publish release
2. GitHub action will be triggered that will build the release and wait for publish approval
3. after CODEOWNERS approve it, it will be published to Dockerhub public repository

### Helm Chart

Helm chart is published to <https://helm.solarwinds.com>.

1. Update property `version` in [deploy/helm/Chart.yaml](../deploy/helm/Chart.yaml). (follow the [SemVer 2](https://semver.org/spec/v2.0.0.html) format).
2. Update [deploy/helm/CHANGELOG.md](../deploy/helm/CHANGELOG.md):
   1. Create release record with the right version and the date.
   2. Write all changes recorded in `Unreleased` section into the release.
3. Create PR for the changes to the `master` branch and merge them.
4. Run "Release Helm Chart" GitHub action workflow.
5. Find relevant release in GitHub, edit it and write all changes recorded into [CHANGELOG.md](../deploy/helm/CHANGELOG.md) into its description.
6. Review PR that was created for the changes to the `gh-pages` branch and merge them.
