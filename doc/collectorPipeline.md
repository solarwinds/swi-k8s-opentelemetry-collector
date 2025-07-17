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

    mc_metricsPrometheusPipeline: 'metrics/prometheus' pipeline
    state mc_metricsPrometheusPipeline {
      mc_r3: 'forward/prometheus' connector
      mc_e3: 'forward/metric-exporter' connector
      mc_r3 --> mc_e3 : processors
    }

    mc_metricsPrometheusNodeMetricsPipeline: 'metrics/prometheus-node-metrics' pipeline
    state mc_metricsPrometheusNodeMetricsPipeline {
      mc_r4: 'prometheus/node-metrics' receiver
      mc_e4: 'forward/prometheus' connector
      mc_r4 --> mc_e4 : processors
    }

    mc_metricsPrometheusServerPipeline: 'metrics/prometheus-server' pipeline
    state mc_metricsPrometheusServerPipeline {
      mc_r5: 'prometheus/prometheus-server' receiver
      mc_e5: 'forward/prometheus' connector
      mc_r5 --> mc_e5 : processors
    }

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
      md_e2a: 'forward/relationship-state-events-workload-workload' connector
      md_e2b: 'forward/relationship-state-events-workload-service' connector
      md_e2c: 'forward/not-relationship-state-events' connector
      md_r2 --> md_e2a : processors
      md_r2 --> md_e2b : processors
      md_r2 --> md_e2c : processors
    }

    md_metricsRelationshipWorkloadPipeline: 'metrics/relationship-state-events-workload-workload-preparation' pipeline
    state md_metricsRelationshipWorkloadPipeline {
      md_r3: 'forward/relationship-state-events-workload-workload' connector
      md_e3a: 'forward/discovery-istio-metrics-clean' connector
      md_e3b: 'solarwindsentity/istio-workload-workload' connector
      md_r3 --> md_e3a : processors
      md_r3 --> md_e3b : processors
    }

    md_metricsRelationshipServicePipeline: 'metrics/relationship-state-events-workload-service-preparation' pipeline
    state md_metricsRelationshipServicePipeline {
      md_r4: 'forward/relationship-state-events-workload-service' connector
      md_e4a: 'forward/discovery-istio-metrics-clean' connector
      md_e4b: 'solarwindsentity/istio-workload-service' connector
      md_r4 --> md_e4a : processors
      md_r4 --> md_e4b : processors
    }

    md_metricsNotRelationshipPipeline: 'metrics/not-relationship-state-events-preparation' pipeline
    state md_metricsNotRelationshipPipeline {
      md_r5: 'forward/not-relationship-state-events' connector
      md_e5: 'forward/discovery-istio-metrics-clean' connector
      md_r5 --> md_e5 : processors
    }

    md_metricsDiscoveryIstioCleanPipeline: 'metrics/discovery-istio-clean' pipeline
    state md_metricsDiscoveryIstioCleanPipeline {
      md_r6: 'forward/discovery-istio-metrics-clean' connector
      md_e6: 'forward/metric-exporter' connector
      md_r6 --> md_e6 : processors
    }

    md_logsStateEventsEntitiesPipeline: 'logs/stateevents-entities' pipeline
    state md_logsStateEventsEntitiesPipeline {
      md_r7a: 'solarwindsentity/istio-workload-workload' connector
      md_r7b: 'solarwindsentity/istio-workload-service' connector
      md_e7: 'otlp' exporter
      md_r7a --> md_e7 : processors
      md_r7b --> md_e7 : processors
    }

    md_logsStateEventsRelationshipsPipeline: 'logs/stateevents-relationships' pipeline
    state md_logsStateEventsRelationshipsPipeline {
      md_r8a: 'solarwindsentity/istio-workload-workload' connector
      md_r8b: 'solarwindsentity/istio-workload-service' connector
      md_e8: 'otlp' exporter
      md_r8a --> md_e8 : processors
      md_r8b --> md_e8 : processors
    }

    md_metricsDiscoveryCustomPipeline: 'metrics/discovery-custom' pipeline
    state md_metricsDiscoveryCustomPipeline {
      md_r9: 'routing/discovered_metrics' connector
      md_e9: 'forward/metric-exporter' connector
      md_r9 --> md_e9 : processors
    }

    md_metricsPipeline: 'metrics' pipeline
    state md_metricsPipeline {
      md_r10: 'forward/metric-exporter' connector
      md_e10: 'otlp' exporter
      md_r10 --> md_e10 : processors
    }

    md_metricsDiscoveryScrapePipeline --> md_metricsDiscoveryCustomPipeline
    md_metricsDiscoveryScrapePipeline --> md_metricsDiscoveryIstioPipeline
    md_metricsDiscoveryIstioPipeline --> md_metricsRelationshipWorkloadPipeline
    md_metricsDiscoveryIstioPipeline --> md_metricsRelationshipServicePipeline
    md_metricsDiscoveryIstioPipeline --> md_metricsNotRelationshipPipeline
    md_metricsRelationshipWorkloadPipeline --> md_metricsDiscoveryIstioCleanPipeline
    md_metricsRelationshipServicePipeline --> md_metricsDiscoveryIstioCleanPipeline
    md_metricsNotRelationshipPipeline --> md_metricsDiscoveryIstioCleanPipeline
    md_metricsDiscoveryIstioCleanPipeline --> md_metricsPipeline
    md_metricsRelationshipWorkloadPipeline --> md_logsStateEventsEntitiesPipeline
    md_metricsRelationshipServicePipeline --> md_logsStateEventsEntitiesPipeline
    md_metricsRelationshipWorkloadPipeline --> md_logsStateEventsRelationshipsPipeline
    md_metricsRelationshipServicePipeline --> md_logsStateEventsRelationshipsPipeline
    md_metricsDiscoveryCustomPipeline --> md_metricsPipeline
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
      ec_r2: 'swok8sobjects' receiver
      ec_e2: 'otlp' exporter
      ec_r2 --> ec_e2 : processors
    }

    ec_manifestsKeepalivePipeline: 'logs/manifests-keepalive' pipeline
    state ec_manifestsKeepalivePipeline {
      ec_r3: 'swok8sobjects/keepalive' receiver
      ec_e3: 'solarwindsentity/keepalive' connector
      ec_r3 --> ec_e3 : processors
    }

    ec_stateEventsPipeline: 'logs/stateevents' pipeline
    state ec_stateEventsPipeline {
      ec_r4: 'solarwindsentity/keepalive' connector
      ec_e4: 'otlp' exporter
      ec_r4 --> ec_e4 : processors
    }

    ec_manifestsKeepalivePipeline --> ec_stateEventsPipeline
  }

%% ──────────────────────────────────────────────────────────────────────────────
%%  GATEWAY COLLECTOR
%% ──────────────────────────────────────────────────────────────────────────────
  gatewayCollectorDeployment: Gateway Collector Deployment
  state gatewayCollectorDeployment {

    gw_metricsCommonInPipeline: 'metrics/common_in' pipeline
    state gw_metricsCommonInPipeline {
      gw_mr1: 'otlp' receiver
      gw_me1: 'routing/metrics' connector
      gw_mr1 --> gw_me1 : processors
    }

    gw_metricsPipeline: 'metrics' pipeline
    state gw_metricsPipeline {
      gw_mr2: 'routing/metrics' connector
      gw_me2: 'forward/metrics_common' connector
      gw_mr2 --> gw_me2 : processors
    }

    gw_metricsCommonOutPipeline: 'metrics/common_out' pipeline
    state gw_metricsCommonOutPipeline {
      gw_mr3: 'forward/metrics_common' connector
      gw_me3: 'otlp' exporter
      gw_mr3 --> gw_me3 : processors
    }

    gw_metricsBeylaNetworkPipeline: 'metrics/beyla-network-entities-and-relationships' pipeline
    state gw_metricsBeylaNetworkPipeline {
      gw_mr4: 'routing/metrics' connector
      gw_me4a: 'forward/metrics_common' connector
      gw_me4b: 'solarwindsentity/beyla-relationships' connector
      gw_me4c: 'solarwindsentity/beyla-entities' connector
      gw_mr4 --> gw_me4a : processors
      gw_mr4 --> gw_me4b : processors
      gw_mr4 --> gw_me4c : processors
    }

    gw_logsBeylaStateEventsEntitiesPipeline: 'logs/beyla-stateevents-entities' pipeline
    state gw_logsBeylaStateEventsEntitiesPipeline {
      gw_lr1: 'solarwindsentity/beyla-entities' connector
      gw_le1: 'otlp' exporter
      gw_lr1 --> gw_le1 : processors
    }

    gw_logsBeylaStateEventsRelationshipsPipeline: 'logs/beyla-stateevents-relationships' pipeline
    state gw_logsBeylaStateEventsRelationshipsPipeline {
      gw_lr2: 'solarwindsentity/beyla-relationships' connector
      gw_le2: 'otlp' exporter
      gw_lr2 --> gw_le2 : processors
    }

    gw_logsPipeline: 'logs' pipeline
    state gw_logsPipeline {
      gw_lr3: 'otlp' receiver
      gw_le3: 'otlp' exporter
      gw_lr3 --> gw_le3 : processors
    }

    gw_tracesPipeline: 'traces' pipeline
    state gw_tracesPipeline {
      gw_tr: 'otlp' receiver
      gw_te: 'otlp' exporter
      gw_tr --> gw_te : processors
    }

    gw_metricsCommonInPipeline --> gw_metricsPipeline
    gw_metricsCommonInPipeline --> gw_metricsBeylaNetworkPipeline
    gw_metricsPipeline --> gw_metricsCommonOutPipeline
    gw_metricsBeylaNetworkPipeline --> gw_metricsCommonOutPipeline
    gw_metricsBeylaNetworkPipeline --> gw_logsBeylaStateEventsEntitiesPipeline
    gw_metricsBeylaNetworkPipeline --> gw_logsBeylaStateEventsRelationshipsPipeline
  }

%% ──────────────────────────────────────────────────────────────────────────────
%%  OTHER COMPONENTS
%% ──────────────────────────────────────────────────────────────────────────────
  kubestatemetricsDeployment: KubeStateMetrics Deployment
  ebpfreducerDeployment: eBPF Reducer Deployment
  beylaComponent: Beyla Auto-Instrumentation

  kubestatemetricsDeployment --> mc_metricsKubestatemetricsPipeline
  ebpfreducerDeployment --> gw_metricsCommonInPipeline
  beylaComponent --> gw_metricsCommonInPipeline : traces/metrics

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
      nc_e6a: 'forward/relationship-state-events-workload-workload' connector
      nc_e6b: 'forward/relationship-state-events-workload-service' connector
      nc_e6c: 'forward/not-relationship-state-events' connector
      nc_r6 --> nc_e6a : processors
      nc_r6 --> nc_e6b : processors
      nc_r6 --> nc_e6c : processors
    }

    nc_metricsRelationshipWorkloadPipeline: 'metrics/relationship-state-events-workload-workload-preparation' pipeline
    state nc_metricsRelationshipWorkloadPipeline {
      nc_r7: 'forward/relationship-state-events-workload-workload' connector
      nc_e7a: 'forward/discovery-istio-metrics-clean' connector
      nc_e7b: 'solarwindsentity/istio-workload-workload' connector
      nc_r7 --> nc_e7a : processors
      nc_r7 --> nc_e7b : processors
    }

    nc_metricsRelationshipServicePipeline: 'metrics/relationship-state-events-workload-service-preparation' pipeline
    state nc_metricsRelationshipServicePipeline {
      nc_r8: 'forward/relationship-state-events-workload-service' connector
      nc_e8a: 'forward/discovery-istio-metrics-clean' connector
      nc_e8b: 'solarwindsentity/istio-workload-service' connector
      nc_r8 --> nc_e8a : processors
      nc_r8 --> nc_e8b : processors
    }

    nc_metricsNotRelationshipPipeline: 'metrics/not-relationship-state-events-preparation' pipeline
    state nc_metricsNotRelationshipPipeline {
      nc_r9: 'forward/not-relationship-state-events' connector
      nc_e9: 'forward/discovery-istio-metrics-clean' connector
      nc_r9 --> nc_e9 : processors
    }

    nc_metricsDiscoveryIstioCleanPipeline: 'metrics/discovery-istio-clean' pipeline
    state nc_metricsDiscoveryIstioCleanPipeline {
      nc_r10: 'forward/discovery-istio-metrics-clean' connector
      nc_e10: 'forward/metric-exporter' connector
      nc_r10 --> nc_e10 : processors
    }

    nc_logsStateEventsEntitiesPipeline: 'logs/stateevents-entities' pipeline
    state nc_logsStateEventsEntitiesPipeline {
      nc_r11a: 'solarwindsentity/istio-workload-workload' connector
      nc_r11b: 'solarwindsentity/istio-workload-service' connector
      nc_e11: 'otlp' exporter
      nc_r11a --> nc_e11 : processors
      nc_r11b --> nc_e11 : processors
    }

    nc_logsStateEventsRelationshipsPipeline: 'logs/stateevents-relationships' pipeline
    state nc_logsStateEventsRelationshipsPipeline {
      nc_r12a: 'solarwindsentity/istio-workload-workload' connector
      nc_r12b: 'solarwindsentity/istio-workload-service' connector
      nc_e12: 'otlp' exporter
      nc_r12a --> nc_e12 : processors
      nc_r12b --> nc_e12 : processors
    }

    nc_metricsDiscoveryCustomPipeline: 'metrics/discovery-custom' pipeline
    state nc_metricsDiscoveryCustomPipeline {
      nc_r13: 'routing/discovered_metrics' connector
      nc_e13: 'forward/metric-exporter' connector
      nc_r13 --> nc_e13 : processors
    }

    nc_metricsNodePipeline: 'metrics/node' pipeline
    state nc_metricsNodePipeline {
      nc_r14: 'receiver_creator/node' receiver
      nc_e14: 'forward/metric-exporter' connector
      nc_r14 --> nc_e14 : processors
    }

    nc_logsContainerPipeline --> nc_logsPipeline
    nc_logsJournalPipeline --> nc_logsPipeline

    nc_metricsDiscoveryScrapePipeline --> nc_metricsDiscoveryCustomPipeline
    nc_metricsDiscoveryScrapePipeline --> nc_metricsDiscoveryIstioPipeline
    nc_metricsDiscoveryIstioPipeline --> nc_metricsRelationshipWorkloadPipeline
    nc_metricsDiscoveryIstioPipeline --> nc_metricsRelationshipServicePipeline
    nc_metricsDiscoveryIstioPipeline --> nc_metricsNotRelationshipPipeline
    nc_metricsRelationshipWorkloadPipeline --> nc_metricsDiscoveryIstioCleanPipeline
    nc_metricsRelationshipServicePipeline --> nc_metricsDiscoveryIstioCleanPipeline
    nc_metricsNotRelationshipPipeline --> nc_metricsDiscoveryIstioCleanPipeline
    nc_metricsDiscoveryIstioCleanPipeline --> nc_metricsPipeline
    nc_metricsRelationshipWorkloadPipeline --> nc_logsStateEventsEntitiesPipeline
    nc_metricsRelationshipServicePipeline --> nc_logsStateEventsEntitiesPipeline
    nc_metricsRelationshipWorkloadPipeline --> nc_logsStateEventsRelationshipsPipeline
    nc_metricsRelationshipServicePipeline --> nc_logsStateEventsRelationshipsPipeline

    nc_metricsDiscoveryCustomPipeline --> nc_metricsPipeline
    nc_metricsNodePipeline --> nc_metricsPipeline
  }

```
