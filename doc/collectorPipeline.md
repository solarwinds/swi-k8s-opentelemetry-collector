# Collector pipeline

The `swo-k8s-collector` Helm chart deploys several k8s workflows. The following chart shows dataflows inside the deployed OTEL collectors: `MetricsCollector Deployment`, `MetricsDiscovery Deployment`, `EventsCollector Deployment`, `NodeCollector DaemonSet`, and `Gateway Collector Deployment`.

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
    mc_metricsPrometheusNodeMetricsPipeline --> mc_metricsPrometheusPipeline
    mc_metricsPrometheusServerPipeline --> mc_metricsPrometheusPipeline
  }

  metricsDiscoveryDeployment: MetricsDiscovery Deployment
  state metricsDiscoveryDeployment {
    md_metricsDiscoveryPipeline: 'metrics/discovery' pipeline
    state md_metricsDiscoveryPipeline {
      md_r1: 'receiver_creator/discovery' receiver
      md_e1: 'forward/metric-exporter' connector
      md_r1 --> md_e1 : processors
    }

    md_metricsPipeline: 'metrics' pipeline
    state md_metricsPipeline {
      md_r2: 'forward/metric-exporter' connector
      md_e2: 'otlp' exporter
      md_r2 --> md_e2 : processors
    }

    md_metricsDiscoveryPipeline --> md_metricsPipeline
  }

  eventsCollectorDeployment: EventsCollector Deployment
  state eventsCollectorDeployment {

  ec_logsPipeline: 'logs' pipeline
    state ec_logsPipeline {
      ec_r1: 'k8s_events' receiver
      ec_e1: 'otlp' exporter
      ec_r1 --> ec_e1 : processors
    }

  ec_manifestsPipeline: 'logs/manifests' pipeline
    state ec_manifestsPipeline {
      ec_r2: 'k8sobjects' receiver
      ec_e2: 'otlp' exporter
      ec_r2 --> ec_e2 : processors
    }

  }

  gatewayCollectorDeployment: Gateway Collector Deployment
  state gatewayCollectorDeployment {
    gw_metricsPipeline: 'metrics' pipeline
    state gw_metricsPipeline {
      gw_mr: 'otlp' receiver
      gw_me: 'otlp' exporter
      gw_mr --> gw_me : processors
    }

    gw_logsPipeline: 'logs' pipeline
    state gw_logsPipeline {
      gw_lr: 'otlp' receiver
      gw_le: 'otlp' exporter
      gw_lr --> gw_le : processors
    }

    gw_tracesPipeline: 'traces' pipeline
    state gw_tracesPipeline {
      gw_tr: 'otlp' receiver
      gw_te: 'otlp' exporter
      gw_tr --> gw_te : processors
    }
  }

  kubestatemetricsDeployment: KubeStateMetrics Deployment
  ebpfreducerDeployment: eBPF Reducer Deployment
  beylaComponent: Beyla Auto-Instrumentation

  kubestatemetricsDeployment --> mc_metricsKubestatemetricsPipeline
  ebpfreducerDeployment --> mc_metricsOtlpPipeline
  beylaComponent --> gatewayCollectorDeployment : traces/metrics

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
