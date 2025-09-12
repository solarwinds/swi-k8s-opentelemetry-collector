# Collector pipeline

The `swo-k8s-collector` Helm chart deploys several k8s workflows. The following chart shows dataflows inside the deployed OTEL collectors: `MetricsCollector Deployment`, `MetricsDiscovery Deployment`, `EventsCollector Deployment`, `NodeCollector DaemonSet`, and `Gateway Collector Deployment`.

Note: Some of the pipelines may not be actually utilized, depending on the environment and the Helm chart's settings provided during its installation.


## Pipelines
![Gateway Config](./collector-image.png)


[![Gateway Config](./collector-image.png)](./collector-image.png)


<a href="./collector-image.png" target="_blank">
  <img src="./collector-image.png" alt="Gateway Config" />
</a>
