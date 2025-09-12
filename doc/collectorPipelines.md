## Viewing collector pipelines in [OTelBin](https://www.otelbin.io/)
- First you need to render the pipeline defined by the helm chart
- Go to the `./deploy` folder
- Create `values.yaml` file with options you want to use for rendering the helm chart. Based on the options, some parts of the pipeline can be added or removed. Below is example of simple `values.yaml`
```shell
cluster:
    name: test-cluster
otel:
    endpoint: example.endpoint.com:443 
```
- Update helm dependencies
```shell
helm dependency update ./helm
```
- Render the helm chart of the pipeline/config map you are interested in. Here is example of rendering `metrics-collector-config-map.yaml` and outputing results to `output.yaml`
```shell
helm template my-release ./helm -f values.yaml --show-only templates/metrics-collector-config-map.yaml > output.yaml
```
- Open the `output.yaml` with config map definition and copy configuration part to clipboard
- Open the [OTelBin](https://www.otelbin.io/) website and replace the example OTel configuration with contents of the clipboard

