# Collector pipeline

The `swo-k8s-collector` Helm chart deploys several k8s workflows. The following chart shows dataflows inside the deployed OTEL collectors: `MetricsCollector Deployment`, `EventsCollector Deployment` and `NodeCollector DaemonSet`

Note: Some of the pipelines may not be actually utilized, depending on the environment and the Helm chart's settings provided during its installation.

```mermaid
stateDiagram-v2

  metricsCollectorDeployment: MetricsCollector Deployment
  state metricsCollectorDeployment {

    mc_metricsPipeline: 'metrics' pipeline
    state mc_metricsPipeline {
      mc_r1: 'forward/metric-exporter' connector
      mc_e1: 'otlp' exporter
      mc_r1 --> mc_e1 : processors
    }

    mc_metricsKubestatemetricsPipeline: 'metrics/kubestatemetrics' pipeline
    state mc_metricsKubestatemetricsPipeline {
      mc_r2: 'prometheus/kube-state-metrics' receiver
      mc_e2: 'forward/prometheus' connector
      mc_r2 --> mc_e2 : processors
    }

    mc_metricsOpencostPipeline: 'metrics/opencost' pipeline
    state mc_metricsOpencostPipeline {
      mc_r3: 'prometheus/opencost' receiver
      mc_e3: 'forward/prometheus' connector
      mc_r3 --> mc_e3 : processors
    }

    mc_metricsOtlpPipeline: 'metrics/otlp' pipeline
    state mc_metricsOtlpPipeline {
      mc_r4: 'otlp' receiver
      mc_e4: 'forward/metric-exporter' connector
      mc_r4 --> mc_e4 : processors
    }

    mc_metricsPrometheusPipeline: 'metrics/prometheus' pipeline
    state mc_metricsPrometheusPipeline {
      mc_r5: 'forward/prometheus' connector
      mc_e5: 'forward/metric-exporter' connector
      mc_r5 --> mc_e5 : processors
    }

    mc_metricsPrometheusNodeMetricsPipeline: 'metrics/prometheus-node-metrics' pipeline
    state mc_metricsPrometheusNodeMetricsPipeline {
      mc_r6: 'prometheus/node-metrics' receiver
      mc_e6: 'forward/prometheus' connector
      mc_r6 --> mc_e6 : processors
    }

    mc_metricsPrometheusServerPipeline: 'metrics/prometheus-server' pipeline
    state mc_metricsPrometheusServerPipeline {
      mc_r7: 'prometheus/prometheus-server' receiver
      mc_e7: 'forward/prometheus' connector
      mc_r7 --> mc_e7 : processors
    }

    mc_metricsOtlpPipeline --> mc_metricsPipeline
    mc_metricsPrometheusPipeline --> mc_metricsPipeline

    mc_metricsKubestatemetricsPipeline --> mc_metricsPrometheusPipeline
    mc_metricsOpencostPipeline --> mc_metricsPrometheusPipeline
    mc_metricsPrometheusNodeMetricsPipeline --> mc_metricsPrometheusPipeline
    mc_metricsPrometheusServerPipeline --> mc_metricsPrometheusPipeline
  }

  eventsCollectorDeployment: EventsCollector Deployment
  state eventsCollectorDeployment {

  ec_logsPipeline: 'logs' pipeline
    state ec_logsPipeline {
      ec_r1: 'k8s_events' receiver
      ec_e1: 'otlp' exporter
      ec_r1 --> ec_e1 : processors
    }

  }

  kubestatemetricsDeployment: KubeStateMetrics Deployment
  opencostDeployment: OpenCost Deployment
  ebpfreducerDeployment: eBPF Reducer Deployment

  kubestatemetricsDeployment --> mc_metricsKubestatemetricsPipeline
  opencostDeployment --> mc_metricsOpencostPipeline
  ebpfreducerDeployment --> mc_metricsOtlpPipeline

  nodeCollectorDaemonset: NodeCollector DaemonSet
  state nodeCollectorDaemonset {

  nc_logsPipeline: 'logs' pipeline
    state nc_logsPipeline {
      nc_r1: 'forward/logs-exporter' connector
      nc_e1: 'otlp' exporter
      nc_r1 --> nc_e1 : processors
    }

    nc_logsContainerPipeline: 'logs/container' pipeline
    state nc_logsContainerPipeline {
      nc_r2: 'filelog' receiver
      nc_e2: 'forward/logs-exporter' connector
      nc_r2 --> nc_e2 : processors
    }

    nc_logsJournalPipeline: 'logs/journal' pipeline
    state nc_logsJournalPipeline {
      nc_r3: 'journald' receiver
      nc_e3: 'forward/logs-exporter' connector
      nc_r3 --> nc_e3 : processors
    }

    nc_metricsPipeline: 'metrics' pipeline
    state nc_metricsPipeline {
      nc_r4: 'forward/metric-exporter' connector
      nc_e4: 'otlp' exporter
      nc_r4 --> nc_e4 : processors
    }

    nc_metricsDiscoveryPipeline: 'metrics/discovery' pipeline
    state nc_metricsDiscoveryPipeline {
      nc_r5: 'receiver_creator/discovery' receiver
      nc_e5: 'forward/metric-exporter' connector
      nc_r5 --> nc_e5 : processors
    }

    nc_metricsNodePipeline: 'metrics/node' pipeline
    state nc_metricsNodePipeline {
      nc_r6: 'receiver_creator/node' receiver
      nc_e6: 'forward/metric-exporter' connector
      nc_r6 --> nc_e6 : processors
    }

    nc_logsContainerPipeline --> nc_logsPipeline
    nc_logsJournalPipeline --> nc_logsPipeline

    nc_metricsDiscoveryPipeline --> nc_metricsPipeline
    nc_metricsNodePipeline --> nc_metricsPipeline

  }

```
