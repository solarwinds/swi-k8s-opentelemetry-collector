# Development

## Prerequisites

- [Skaffold](https://skaffold.dev) at least [v2.0.3](https://github.com/GoogleContainerTools/skaffold/releases/tag/v2.0.3)
  - On windows, do not install it using choco due to [this issue](https://github.com/GoogleContainerTools/skaffold/issues/4058)
- [Kustomize](https://kustomize.io): `choco install kustomize`
- [Helm](https://helm.sh): `choco install kubernetes-helm`
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

## Develop against remote prometheus

You can port forward Prometheus server to localhost:9090 and run

```shell
skaffold dev -p=remote-prometheus
```

In order to change Prometheus endpoint that is hosted on HTTPS you can adjust skaffold.yaml file:

- add `otel.metrics.prometheus.scheme: https`
- update `otel.metrics.prometheus.url: <remote prometheus>`

## Updating manifest

Temporarily there will be `manifest.yaml` and Helm chart in the repository. In order to avoid maintaining two sources, the `manifest.yaml` is generated using `helm template` command. So please do not write directly to `manifest.yaml` file.

Update Helm chart and use following command to update the manifest:

```shell
helm template swo-k8s-collector deploy/helm -n="<NAMESPACE>" --set-string externalRenderer=true > deploy/k8s/manifest.yaml
```

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

1. Update property `version` in [deploy/helm/Chart.yaml](deploy/helm/Chart.yaml). (follow the [SemVer 2](https://semver.org/spec/v2.0.0.html) format)
2. Propagate the version change to `manifest.yaml` (see [Updating manifest](#updating-manifest) for more info):

    ```shell
    helm template swo-k8s-collector deploy/helm -n="<NAMESPACE>" --set-string externalRenderer=true > deploy/k8s/manifest.yaml
    ```

3. Create PR for the changes to the `main` branch and merge them.
4. Tag the merged commit in GitHub. Use format `helm-v<chartVersion>`, e.g. `helm-v2.0.1-alpha.1`.
5. Package the Helm chart:

    ```shell
    helm package deploy\helm\
    ```

    It will create a file with name like `swo-k8s-collector-2.0.1-alpha.1.tgz`.
6. Switch to branch `gh-pages`. The file generated in the previous step should stay in the folder.
7. Package the Helm chart:

    ```shell
    helm repo index . --url https://helm.solarwinds.com
    ```

    It will update the `index.yaml` file with a new entry for the file.
8. Create PR for the changes to the `gh-pages` branch and merge them.
