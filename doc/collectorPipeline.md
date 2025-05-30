# Collector pipeline

The `swo-k8s-collector` Helm chart deploys several k8s workflows. The following chart shows dataflows inside the deployed OTEL collectors: `MetricsCollector Deployment`, `MetricsDiscovery Deployment`, `EventsCollector Deployment`, `NodeCollector DaemonSet`, and `Gateway Collector Deployment`.

Note: Some of the pipelines may not be actually utilized, depending on the environment and the Helm chart's settings provided during its installation.

```mermaid
stateDiagram-v2
%%
%% ──────────────────────────────────────────────────────────────────────────────
%%  METRICS COLLECTOR
%% ──────────────────────────────────────────────────────────────────────────────
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
      mc_r3: 'otlp' receiver
      mc_e3: 'forward/metric-exporter' connector
      mc_r3 --> mc_e3 : processors
    }

    mc_metricsPrometheusPipeline: 'metrics/prometheus' pipeline
    state mc_metricsPrometheusPipeline {
      mc_r4: 'forward/prometheus' connector
      mc_e4: 'forward/metric-exporter' connector
      mc_r4 --> mc_e4 : processors
    }

    mc_metricsPrometheusNodeMetricsPipeline: 'metrics/prometheus-node-metrics' pipeline
    state mc_metricsPrometheusNodeMetricsPipeline {
      mc_r5: 'prometheus/node-metrics' receiver
      mc_e5: 'forward/prometheus' connector
      mc_r5 --> mc_e5 : processors
    }

    mc_metricsPrometheusServerPipeline: 'metrics/prometheus-server' pipeline
    state mc_metricsPrometheusServerPipeline {
      mc_r6: 'prometheus/prometheus-server' receiver
      mc_e6: 'forward/prometheus' connector
      mc_r6 --> mc_e6 : processors
    }

    mc_metricsOtlpPipeline --> mc_metricsPipeline
    mc_metricsPrometheusPipeline --> mc_metricsPipeline
    mc_metricsKubestatemetricsPipeline --> mc_metricsPrometheusPipeline
    mc_metricsPrometheusNodeMetricsPipeline --> mc_metricsPrometheusPipeline
    mc_metricsPrometheusServerPipeline --> mc_metricsPrometheusPipeline
  }

%% ──────────────────────────────────────────────────────────────────────────────
%%  METRICS-DISCOVERY COLLECTOR
%% ──────────────────────────────────────────────────────────────────────────────
  metricsDiscoveryDeployment: MetricsDiscovery Deployment
  state metricsDiscoveryDeployment {

    md_metricsDiscoveryScrapePipeline: 'metrics/discovery-scrape' pipeline
    state md_metricsDiscoveryScrapePipeline {
      md_r1: 'receiver_creator/discovery' receiver
      md_e1: 'routing/discovered_metrics' connector
      md_r1 --> md_e1 : processors
    }

    md_metricsDiscoveryIstioPipeline: 'metrics/discovery-istio' pipeline
    state md_metricsDiscoveryIstioPipeline {
      md_r2: 'routing/discovered_metrics' connector
      md_e2a: 'forward/metric-exporter' connector
      md_e2b: 'forward/relationship-state-events-workload-workload' connector
      md_e2c: 'forward/relationship-state-events-workload-service' connector
      md_r2 --> md_e2a : processors
      md_r2 --> md_e2b : processors
      md_r2 --> md_e2c : processors
    }

    md_metricsRelationshipWorkloadPipeline: 'metrics/relationship-state-events-workload-workload-preparation' pipeline
    state md_metricsRelationshipWorkloadPipeline {
      md_r3: 'forward/relationship-state-events-workload-workload' connector
      md_e3: 'solarwindsentity/istio-workload-workload' connector
      md_r3 --> md_e3 : processors
    }

    md_metricsRelationshipServicePipeline: 'metrics/relationship-state-events-workload-service-preparation' pipeline
    state md_metricsRelationshipServicePipeline {
      md_r4: 'forward/relationship-state-events-workload-service' connector
      md_e4: 'solarwindsentity/istio-workload-service' connector
      md_r4 --> md_e4 : processors
    }

    md_logsStateEventsPipeline: 'logs/stateevents' pipeline
    state md_logsStateEventsPipeline {
      md_r5a: 'solarwindsentity/istio-workload-workload' connector
      md_r5b: 'solarwindsentity/istio-workload-service' connector
      md_e5: 'otlp' exporter
      md_r5a --> md_e5 : processors
      md_r5b --> md_e5 : processors
    }

    md_metricsDiscoveryCustomPipeline: 'metrics/discovery-custom' pipeline
    state md_metricsDiscoveryCustomPipeline {
      md_r6: 'routing/discovered_metrics' connector
      md_e6: 'forward/metric-exporter' connector
      md_r6 --> md_e6 : processors
    }

    md_metricsDiscoveryScrapePipeline --> md_metricsDiscoveryCustomPipeline
    md_metricsDiscoveryScrapePipeline --> md_metricsDiscoveryIstioPipeline
    md_metricsDiscoveryIstioPipeline --> md_metricsRelationshipWorkloadPipeline
    md_metricsDiscoveryIstioPipeline --> md_metricsRelationshipServicePipeline
    md_metricsRelationshipWorkloadPipeline --> md_logsStateEventsPipeline
    md_metricsRelationshipServicePipeline --> md_logsStateEventsPipeline
  }

%% ──────────────────────────────────────────────────────────────────────────────
%%  EVENTS COLLECTOR
%% ──────────────────────────────────────────────────────────────────────────────
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

%% ──────────────────────────────────────────────────────────────────────────────
%%  GATEWAY COLLECTOR
%% ──────────────────────────────────────────────────────────────────────────────
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

%% ──────────────────────────────────────────────────────────────────────────────
%%  OTHER COMPONENTS
%% ──────────────────────────────────────────────────────────────────────────────
  kubestatemetricsDeployment: KubeStateMetrics Deployment
  ebpfreducerDeployment: eBPF Reducer Deployment
  beylaComponent: Beyla Auto-Instrumentation

  kubestatemetricsDeployment --> mc_metricsKubestatemetricsPipeline
  ebpfreducerDeployment --> mc_metricsOtlpPipeline
  beylaComponent --> gatewayCollectorDeployment : traces/metrics

%% ──────────────────────────────────────────────────────────────────────────────
%%  NODE COLLECTOR
%% ──────────────────────────────────────────────────────────────────────────────
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

    nc_metricsDiscoveryScrapePipeline: 'metrics/discovery-scrape' pipeline
    state nc_metricsDiscoveryScrapePipeline {
      nc_r5: 'receiver_creator/discovery' receiver
      nc_e5: 'routing/discovered_metrics' connector
      nc_r5 --> nc_e5 : processors
    }

    nc_metricsDiscoveryIstioPipeline: 'metrics/discovery-istio' pipeline
    state nc_metricsDiscoveryIstioPipeline {
      nc_r6: 'routing/discovered_metrics' connector
      nc_e6a: 'forward/metric-exporter' connector
      nc_e6b: 'forward/relationship-state-events-workload-workload' connector
      nc_e6c: 'forward/relationship-state-events-workload-service' connector
      nc_r6 --> nc_e6a : processors
      nc_r6 --> nc_e6b : processors
      nc_r6 --> nc_e6c : processors
    }

    nc_metricsRelationshipWorkloadPipeline: 'metrics/relationship-state-events-workload-workload-preparation' pipeline
    state nc_metricsRelationshipWorkloadPipeline {
      nc_r7: 'forward/relationship-state-events-workload-workload' connector
      nc_e7: 'solarwindsentity/istio-workload-workload' connector
      nc_r7 --> nc_e7 : processors
    }

    nc_metricsRelationshipServicePipeline: 'metrics/relationship-state-events-workload-service-preparation' pipeline
    state nc_metricsRelationshipServicePipeline {
      nc_r8: 'forward/relationship-state-events-workload-service' connector
      nc_e8: 'solarwindsentity/istio-workload-service' connector
      nc_r8 --> nc_e8 : processors
    }

    nc_logsStateEventsPipeline: 'logs/stateevents' pipeline
    state nc_logsStateEventsPipeline {
      nc_r9a: 'solarwindsentity/istio-workload-workload' connector
      nc_r9b: 'solarwindsentity/istio-workload-service' connector
      nc_e9: 'otlp' exporter
      nc_r9a --> nc_e9 : processors
      nc_r9b --> nc_e9 : processors
    }

    nc_metricsDiscoveryCustomPipeline: 'metrics/discovery-custom' pipeline
    state nc_metricsDiscoveryCustomPipeline {
      nc_r10: 'routing/discovered_metrics' connector
      nc_e10: 'forward/metric-exporter' connector
      nc_r10 --> nc_e10 : processors
    }

    nc_metricsNodePipeline: 'metrics/node' pipeline
    state nc_metricsNodePipeline {
      nc_r11: 'receiver_creator/node' receiver
      nc_e11: 'forward/metric-exporter' connector
      nc_r11 --> nc_e11 : processors
    }

    nc_logsContainerPipeline --> nc_logsPipeline
    nc_logsJournalPipeline --> nc_logsPipeline

    nc_metricsDiscoveryScrapePipeline --> nc_metricsDiscoveryCustomPipeline
    nc_metricsDiscoveryScrapePipeline --> nc_metricsDiscoveryIstioPipeline
    nc_metricsDiscoveryIstioPipeline --> nc_metricsRelationshipWorkloadPipeline
    nc_metricsDiscoveryIstioPipeline --> nc_metricsRelationshipServicePipeline
    nc_metricsRelationshipWorkloadPipeline --> nc_logsStateEventsPipeline
    nc_metricsRelationshipServicePipeline --> nc_logsStateEventsPipeline

    nc_metricsDiscoveryCustomPipeline --> nc_metricsPipeline
    nc_metricsNodePipeline --> nc_metricsPipeline
  }

%% ──────────────────────────────────────────────────────────────────────────────
%%  NEW DISCOVERY COLLECTOR
%% ──────────────────────────────────────────────────────────────────────────────
  discoveryCollectorDeployment: Discovery Collector Deployment
  state discoveryCollectorDeployment {

    dc_metricsPipeline: 'metrics' pipeline
    state dc_metricsPipeline {
      dc_r0: 'forward/metric-exporter' connector
      dc_e0: 'otlp' exporter
      dc_r0 --> dc_e0 : processors
    }

    dc_metricsDiscoveryScrapePipeline: 'metrics/discovery-scrape' pipeline
    state dc_metricsDiscoveryScrapePipeline {
      dc_r1: 'receiver_creator/discovery' receiver
      dc_e1: 'routing/discovered_metrics' connector
      dc_r1 --> dc_e1 : processors
    }

    dc_metricsDiscoveryIstioPipeline: 'metrics/discovery-istio' pipeline
    state dc_metricsDiscoveryIstioPipeline {
      dc_r2: 'routing/discovered_metrics' connector
      dc_e2a: 'forward/metric-exporter' connector
      dc_e2b: 'forward/relationship-state-events-workload-workload' connector
      dc_e2c: 'forward/relationship-state-events-workload-service' connector
      dc_r2 --> dc_e2a : processors
      dc_r2 --> dc_e2b : processors
      dc_r2 --> dc_e2c : processors
    }

    dc_metricsRelationshipWorkloadPipeline: 'metrics/relationship-state-events-workload-workload-preparation' pipeline
    state dc_metricsRelationshipWorkloadPipeline {
      dc_r3: 'forward/relationship-state-events-workload-workload' connector
      dc_e3: 'solarwindsentity/istio-workload-workload' connector
      dc_r3 --> dc_e3 : processors
    }

    dc_metricsRelationshipServicePipeline: 'metrics/relationship-state-events-workload-service-preparation' pipeline
    state dc_metricsRelationshipServicePipeline {
      dc_r4: 'forward/relationship-state-events-workload-service' connector
      dc_e4: 'solarwindsentity/istio-workload-service' connector
      dc_r4 --> dc_e4 : processors
    }

    dc_logsStateEventsPipeline: 'logs/stateevents' pipeline
    state dc_logsStateEventsPipeline {
      dc_r5a: 'solarwindsentity/istio-workload-workload' connector
      dc_r5b: 'solarwindsentity/istio-workload-service' connector
      dc_e5: 'otlp' exporter
      dc_r5a --> dc_e5 : processors
      dc_r5b --> dc_e5 : processors
    }

    dc_metricsDiscoveryCustomPipeline: 'metrics/discovery-custom' pipeline
    state dc_metricsDiscoveryCustomPipeline {
      dc_r6: 'routing/discovered_metrics' connector
      dc_e6: 'forward/metric-exporter' connector
      dc_r6 --> dc_e6 : processors
    }

    dc_metricsDiscoveryScrapePipeline --> dc_metricsDiscoveryCustomPipeline
    dc_metricsDiscoveryScrapePipeline --> dc_metricsDiscoveryIstioPipeline
    dc_metricsDiscoveryIstioPipeline --> dc_metricsRelationshipWorkloadPipeline
    dc_metricsDiscoveryIstioPipeline --> dc_metricsRelationshipServicePipeline
    dc_metricsRelationshipWorkloadPipeline --> dc_logsStateEventsPipeline
    dc_metricsRelationshipServicePipeline --> dc_logsStateEventsPipeline

    dc_metricsDiscoveryCustomPipeline --> dc_metricsPipeline
  }
```
