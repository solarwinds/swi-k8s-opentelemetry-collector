## Viewing Collector Pipelines in [OTelBin](https://www.otelbin.io/)
- First, render the pipeline defined by the Helm chart.
- Go to the `./deploy` folder.
- Create a `values.yaml` file with the options you want to use for rendering the Helm chart. Based on these options, some parts of the pipeline may be added or removed. Below is an example of a simple `values.yaml`:
```shell
cluster:
    name: test-cluster
otel:
    endpoint: example.endpoint.com:443 
```
- Update the Helm dependencies:
```shell
helm dependency update ./helm
```
- Render the Helm chart for the pipeline/ConfigMap you are interested in. Here is an example of rendering `metrics-collector-config-map.yaml` and outputting the results to `output.yaml`:
```shell
helm template my-release ./helm -f values.yaml --show-only templates/metrics-collector-config-map.yaml > output.yaml
```
- Open `output.yaml` with the ConfigMap definition and copy the configuration section to your clipboard.
- Open the [OTelBin](https://www.otelbin.io/) website and replace the example OTel configuration with the contents of your clipboard.

