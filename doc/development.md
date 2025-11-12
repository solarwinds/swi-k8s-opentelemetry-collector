# Development

## Table of contents

- [Contribution Guidelines](#contribution-guidelines)
- [Prerequisites](#prerequisites)
- [Deployment](#deployment)
- [Develop against remote prometheus](#develop-against-remote-prometheus)
- [Helm Unit tests](#helm-unit-tests)
- [Integration tests](#integration-tests)
- [Performance profiling](#performance-profiling)
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

- [Skaffold](https://skaffold.dev) at least [v2.15.0](https://github.com/GoogleContainerTools/skaffold/releases/tag/v2.15.0)
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

- deploy `SWO K8s Collector` helm chart
- deploy Prometheus
- deploy OtelEndpoint mock (to see that customized Otel Collector is sending metrics correctly)

By default it will deploy `SWO K8s Collector` with features that are enabled by default. It is possible to opt-in/opt-out some features using skaffold profiles:

- `operator` - include certmanager, operator and all the features that are related to it (discovery_collector, CRDs)
- `auto-instrumentation` - include OTEL Demo services and enable auto-instrumentation. It requires `operator` to be enabled
- `no-logs` - exclude log collection
- `no-topology` - exclude network topology
- `no-metrics` - exclude metrics collection
- `no-events` - exclude events collection
- `no-tests` - exclude integration tests
- `no-prometheus` - exclude prometheus
- `swo` - send metrics to SWO. This requires following envrionment variables to be set (e.g. using [skaffold.env file](https://skaffold.dev/docs/environment/env-file/)):
  - `SOLARWINDS_OTEL_ENDPOINT` - SWO ingestion endpoint
  - `SOLARWINDS_API_TOKEN` - SWO ingestion API token
- `build-collector` - see [Rebuild `solarwinds otel collector` from sources](#rebuild-solarwinds-otel-collector-from-sources)
- `push` - push built images to remote docker registry (see [Image Repository Handling](https://skaffold.dev/docs/environment/image-registries/))

Example:
```
skaffold dev -p operator,obi
```

Read more about profiles [here](https://skaffold.dev/docs/environment/profiles/)


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

## Rebuild `solarwinds otel collector` from sources
To run the collector in a local environment with `solarwinds otel collector` image build from sources, execute:

```shell
skaffold dev -p=build-collector
```

You may need to update relative path to `solarwinds otel collector` sources in `skaffold.yaml`:

```
    # Path to cloned https://github.com/solarwinds/solarwinds-otel-collector.git repo
    context: ../solarwinds-otel-collector
```


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

## Performance profiling

The `k8s collector` can be configured to enable performance profiling with `pprof`.

### Prerequisites

- `pprof` - to analyze the profiles
- `graphviz` - to render the data as graphs

  ```shell
  choco install graphviz
  ```

- `curl` - to fetch the profiles on computers without `pprof`

### Connecting pprof to the analyzed process on a local machine

1. Deploy the `k8s collector` with setting:

    ```yaml
    diagnostics:
     profiling:
       enabled: true
    ```

2. Port-forward the `pprof` port on any of the `k8s collector`'s Pods:

    ```shell
    kubectl -n <namespace> port-forward pod/<pod-name> 1777:pprof
    ```

3. On a machine with Go runtime, run command:

    ```shell
    go tool pprof http://localhost:1777/debug/pprof/<command>
    ```

    Alternatively, get the `pprof` binary already compiled and use that.

    It will connect to the port-forwarded instance of the `k8s collector`, get the requested data, store it in a local file (path will be mentioned on the screen) and open the file in an interactive mode.

    For documentation about available commands, see [net/http/pprof](https://pkg.go.dev/net/http/pprof).

### Go memory investigation on a remote computer without pprof

1. Deploy the `k8s collector` with setting:

    ```yaml
    diagnostics:
      profiling:
        enabled: true
    ```

2. Wait until some of its instances consumes too much memory.
3. Port-forward the pprof port on the instance locally:

    ```shell
    kubectl -n <namespace> port-forward pod/<pod-name> 1777:pprof
    ```

4. Fetch a memory heap profile:

    ```shell
    curl -s http://localhost:1777/debug/pprof/heap > ./heap.out
    ```

    For ideal results, collect multiple such heap profiles, with enough time between them.

5. Open the heap profile:

    ```shell
    go tool pprof -http=:8080 ./heap.out
    ```

    This will start a local `pprof` web server listening on `http://localhost:8080`.

    If the `pprof` (or Go) is not available on computer, the file can be open on another computer.

6. When done:
   - Stop port-forwarding
   - Close the browser window
   - Stop the `pprofs`s webserver
   - Disable exposing diagnostics in the `k8s collector` using the `diagnostics.profiling.enabled` setting

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

### Helm Chart

1. Create PR with `version` change into [Chart.yaml](../deploy/helm/Chart.yaml). If you need to update an image, change `appVersion` as well.
1. Once the PR is merged, GitHub Action will be triggered to build the release and open a PR to the `gh-pages` branch.
1. Review the PR created for changes to the `gh-pages` branch (which hosts the Helm charts), and merge it.
1. Once the PR is merged, the Helm chart is published to <https://helm.solarwinds.com>.
