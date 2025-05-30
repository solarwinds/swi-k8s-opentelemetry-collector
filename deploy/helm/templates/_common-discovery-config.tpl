{{- define "common-discovery-config.processors" -}}
{{- if .Values.otel.metrics.autodiscovery.prometheusEndpoints.filter }}
filter/metrics-discovery:
  metrics:
{{ toYaml .Values.otel.metrics.autodiscovery.prometheusEndpoints.filter | indent 4 }}
{{- end }}

metricstransform/rename/discovery:
  transforms:
    # add `k8s.` prefix to all metrics
    - include: ^(.*)$$
      match_type: regexp
      action: update
      new_name: {{ .Values.otel.metrics.autodiscovery.prefix }}$${1}

{{- if ne .Values.otel.metrics.autodiscovery.prefix "k8s." }}
  # in case the prefix differs from "k8s." we need to copy the required metrics
  # so that SWO built-in dashboards works correctly
{{- $arrayOfRequiredMetrics := list 
  "etcd_disk_backend_commit_duration_seconds"
  "etcd_disk_wal_fsync_duration_seconds" 
  "etcd_network_client_grpc_received_bytes_total" 
  "etcd_network_client_grpc_sent_bytes_total" 
  "etcd_network_peer_received_bytes_total" 
  "etcd_network_peer_sent_bytes_total" 
  "etcd_server_leader_changes_seen_total" 
  "etcd_server_proposals_applied_total" 
  "etcd_server_proposals_committed_total"
  "etcd_server_proposals_failed_total"
  "etcd_server_proposals_pending"
  "etcd_server_has_leader"
  "etcd_mvcc_db_total_size_in_bytes"
  "process_resident_memory_bytes"
  "grpc_server_started_total"
  "grpc_server_handled_total"
  "rest_client_request_duration_seconds"
  "rest_client_requests_total"
  "workqueue_adds_total"
  "workqueue_depth"
  "workqueue_queue_duration_seconds"
}}
metricstransform/copy-required-metrics:
  transforms:
  {{- $root := . }}
  {{- range $index, $metric := $arrayOfRequiredMetrics }}
    - include: {{ $root.Values.otel.metrics.autodiscovery.prefix }}{{ $metric }}
      action: insert
      new_name: k8s.{{ $metric }}
  {{- end }}
{{- end }}

{{- if .Values.otel.metrics.autodiscovery.prometheusEndpoints.customTransformations.counterToRate }}
cumulativetodelta/discovery:
  include:
    metrics:
{{- range .Values.otel.metrics.autodiscovery.prometheusEndpoints.customTransformations.counterToRate }}
      - {{ . }}
{{- end }}
    match_type: strict
deltatorate/discovery:
  metrics:
{{- range .Values.otel.metrics.autodiscovery.prometheusEndpoints.customTransformations.counterToRate }}
    - {{ . }}
{{- end }}
{{- end }}

metricstransform/istio-metrics:
  transforms:
    - include: {{ .Values.otel.metrics.autodiscovery.prefix }}istio_request_bytes_sum
      action: insert
      new_name: k8s.istio_request_bytes.rate
    - include: {{ .Values.otel.metrics.autodiscovery.prefix }}istio_response_bytes_sum
      action: insert
      new_name: k8s.istio_response_bytes.rate
    - include: {{ .Values.otel.metrics.autodiscovery.prefix }}istio_requests_total
      action: insert
      new_name: k8s.istio_requests.rate
    - include: {{ .Values.otel.metrics.autodiscovery.prefix }}istio_tcp_sent_bytes_total
      action: insert
      new_name: k8s.istio_tcp_sent_bytes.rate
    - include: {{ .Values.otel.metrics.autodiscovery.prefix }}istio_tcp_received_bytes_total
      action: insert
      new_name: k8s.istio_tcp_received_bytes.rate
    - include: k8s.istio_request_bytes.rate
      action: insert
      new_name: k8s.istio_request_bytes.delta
    - include: k8s.istio_response_bytes.rate
      action: insert
      new_name: k8s.istio_response_bytes.delta
    - include: k8s.istio_requests.rate
      action: insert
      new_name: k8s.istio_requests.delta
    - include: k8s.istio_tcp_sent_bytes.rate
      action: insert
      new_name: k8s.istio_tcp_sent_bytes.delta
    - include: k8s.istio_tcp_received_bytes.rate
      action: insert
      new_name: k8s.istio_tcp_received_bytes.delta

cumulativetodelta/istio-metrics:
  include:
    metrics:
      - k8s.istio_request_bytes.rate
      - k8s.istio_response_bytes.rate
      - k8s.istio_request_duration_milliseconds_sum_temp
      - k8s.istio_request_duration_milliseconds_count_temp
      - k8s.istio_requests.rate
      - k8s.istio_tcp_sent_bytes.rate
      - k8s.istio_tcp_received_bytes.rate
      - k8s.istio_request_bytes.delta
      - k8s.istio_response_bytes.delta
      - k8s.istio_requests.delta
      - k8s.istio_tcp_sent_bytes.delta
      - k8s.istio_tcp_received_bytes.delta
    match_type: strict

deltatorate/istio-metrics:
  metrics:
    - k8s.istio_request_bytes.rate
    - k8s.istio_response_bytes.rate
    - k8s.istio_request_duration_milliseconds_sum_temp
    - k8s.istio_request_duration_milliseconds_count_temp
    - k8s.istio_requests.rate
    - k8s.istio_tcp_sent_bytes.rate
    - k8s.istio_tcp_received_bytes.rate

metricsgeneration/istio-metrics:
  rules:
    - name: k8s.istio_request_duration_milliseconds.rate
      type: calculate
      metric1: k8s.istio_request_duration_milliseconds_sum_temp
      metric2: k8s.istio_request_duration_milliseconds_count_temp
      operation: divide

transform/istio-metrics:
  metric_statements:
    - statements:
        - extract_sum_metric(true) where (metric.name == "{{ .Values.otel.metrics.autodiscovery.prefix }}istio_request_bytes" or metric.name == "{{ .Values.otel.metrics.autodiscovery.prefix }}istio_response_bytes" or metric.name == "{{ .Values.otel.metrics.autodiscovery.prefix }}istio_request_duration_milliseconds")
        - extract_count_metric(true) where (metric.name == "{{ .Values.otel.metrics.autodiscovery.prefix }}istio_request_duration_milliseconds")
        - set(metric.name, "k8s.istio_request_duration_milliseconds_sum_temp") where metric.name == "{{ .Values.otel.metrics.autodiscovery.prefix }}istio_request_duration_milliseconds_sum"
        - set(metric.name, "k8s.istio_request_duration_milliseconds_count_temp") where metric.name == "{{ .Values.otel.metrics.autodiscovery.prefix }}istio_request_duration_milliseconds_count"

swok8sworkloadtype/istio:
  workload_mappings:
    - name_attr: source_workload
      namespace_attr: source_workload_namespace
      workload_type_attr: source_workload_type
      expected_types:
        - deployments
        - daemonsets
        - statefulsets
    - name_attr: destination_workload
      namespace_attr: destination_workload_namespace
      workload_type_attr: destination_workload_type
      expected_types:
        - deployments
        - daemonsets
        - statefulsets
    - name_attr: destination_service_name
      namespace_attr: destination_service_namespace
      workload_type_attr: destination_service_type
      expected_types:
        - services

groupbyattrs/istio-relationships:
  keys:
    - sw.k8s.cluster.uid
    - source.k8s.deployment.name
    - source.k8s.statefulset.name
    - source.k8s.daemonset.name
    - source.k8s.namespace.name
    - dest.k8s.deployment.name
    - dest.k8s.statefulset.name
    - dest.k8s.daemonset.name
    - dest.k8s.namespace.name
    - dest.k8s.service.name

filter/keep-workload-workload-relationships:
  error_mode: ignore
  metrics:
    metric:
      - name != "{{ .Values.otel.metrics.autodiscovery.prefix }}istio_request_bytes_sum"
    datapoint:
      - datapoint.attributes["source_workload_type"] == nil or datapoint.attributes["destination_workload_type"] == nil or datapoint.attributes["source_workload_type"] == "" or datapoint.attributes["destination_workload_type"] == ""

filter/keep-workload-service-relationships:
  error_mode: ignore
  metrics:
    metric:
      - name != "{{ .Values.otel.metrics.autodiscovery.prefix }}istio_request_bytes_sum"
    datapoint:
      - datapoint.attributes["source_workload_type"] == nil or datapoint.attributes["destination_service_type"] == nil or datapoint.attributes["source_workload_type"] == "" or datapoint.attributes["destination_service_type"] == ""

transform/istio-workload-workload:
  metric_statements:
    - keep_keys(datapoint.attributes, ["source_workload", "source_workload_namespace", "destination_workload", "destination_workload_namespace", "source_workload_type", "destination_workload_type"])
    - set(datapoint.attributes["source.k8s.deployment.name"], datapoint.attributes["source_workload"]) where datapoint.attributes["source_workload_type"] == "Deployment"
    - set(datapoint.attributes["source.k8s.statefulset.name"], datapoint.attributes["source_workload"]) where datapoint.attributes["source_workload_type"] == "StatefulSet"
    - set(datapoint.attributes["source.k8s.daemonset.name"], datapoint.attributes["source_workload"]) where datapoint.attributes["source_workload_type"] == "DaemonSet"
    - set(datapoint.attributes["source.k8s.namespace.name"], datapoint.attributes["source_workload_namespace"])
    - set(datapoint.attributes["dest.k8s.deployment.name"], datapoint.attributes["destination_workload"]) where datapoint.attributes["destination_workload_type"] == "Deployment"
    - set(datapoint.attributes["dest.k8s.statefulset.name"], datapoint.attributes["destination_workload"]) where datapoint.attributes["destination_workload_type"] == "StatefulSet"
    - set(datapoint.attributes["dest.k8s.daemonset.name"], datapoint.attributes["destination_workload"]) where datapoint.attributes["destination_workload_type"] == "DaemonSet"
    - set(datapoint.attributes["dest.k8s.namespace.name"], datapoint.attributes["destination_workload_namespace"])

transform/istio-workload-service:
  metric_statements:
    - keep_keys(datapoint.attributes, ["source_workload", "source_workload_namespace", "destination_service_name", "destination_service_namespace", "source_workload_type", "destination_service_type"])
    - set(datapoint.attributes["source.k8s.deployment.name"], datapoint.attributes["source_workload"]) where datapoint.attributes["source_workload_type"] == "Deployment"
    - set(datapoint.attributes["source.k8s.statefulset.name"], datapoint.attributes["source_workload"]) where datapoint.attributes["source_workload_type"] == "StatefulSet"
    - set(datapoint.attributes["source.k8s.daemonset.name"], datapoint.attributes["source_workload"]) where datapoint.attributes["source_workload_type"] == "DaemonSet"
    - set(datapoint.attributes["source.k8s.namespace.name"], datapoint.attributes["source_workload_namespace"])
    - set(datapoint.attributes["dest.k8s.service.name"], datapoint.attributes["destination_service_name"]) where datapoint.attributes["destination_service_type"] == "Service"
    - set(datapoint.attributes["dest.k8s.namespace.name"], datapoint.attributes["destination_service_namespace"])

transform/only-relationship-resource-attributes:
  metric_statements:
    - keep_keys(resource.attributes, ["sw.k8s.cluster.uid", "source.k8s.deployment.name", "source.k8s.statefulset.name", "source.k8s.daemonset.name", "source.k8s.job.name", "source.k8s.cronjob.name", "source.k8s.namespace.name", "dest.k8s.deployment.name", "dest.k8s.statefulset.name", "dest.k8s.daemonset.name", "dest.k8s.job.name", "dest.k8s.cronjob.name", "dest.k8s.service.name", "dest.k8s.namespace.name"])

batch/stateevents:
  send_batch_size: 1024
  timeout: 1s
  send_batch_max_size: 1024
{{- end }}


{{- define "common-discovery-config.connectors" -}}
forward/relationship-state-events-workload-workload:
forward/relationship-state-events-workload-service:
routing/discovered_metrics:
  default_pipelines: [metrics/discovery-custom]
  table:
    - context: metric
      pipelines: [metrics/discovery-istio]
      condition: |
        IsMatch(name, "{{ .Values.otel.metrics.autodiscovery.prefix }}istio_")

solarwindsentity/istio-workload-workload:
  source_prefix: "source."
  destination_prefix: "dest."
  schema:
    entities:
      - entity: KubernetesDeployment
        id:
          - sw.k8s.cluster.uid
          - k8s.namespace.name
          - k8s.deployment.name
      - entity: KubernetesStatefulSet
        id:
          - sw.k8s.cluster.uid
          - k8s.namespace.name
          - k8s.statefulset.name
      - entity: KubernetesDaemonSet
        id:
          - sw.k8s.cluster.uid
          - k8s.namespace.name
          - k8s.daemonset.name

    events:
      relationships:
        # source KubernetesDeployment
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesDeployment
          destination_entity: KubernetesDeployment
          attributes:
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesDeployment
          destination_entity: KubernetesStatefulSet
          attributes:
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesDeployment
          destination_entity: KubernetesDaemonSet
          attributes:
        # source KubernetesStatefulSet
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesStatefulSet
          destination_entity: KubernetesDeployment
          attributes:
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesStatefulSet
          destination_entity: KubernetesStatefulSet
          attributes:
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesStatefulSet
          destination_entity: KubernetesDaemonSet
          attributes:
        # source KubernetesDaemonSet
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesDaemonSet
          destination_entity: KubernetesDeployment
          attributes:
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesDaemonSet
          destination_entity: KubernetesStatefulSet
          attributes:
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesDaemonSet
          destination_entity: KubernetesDaemonSet
          attributes:


solarwindsentity/istio-workload-service:
  source_prefix: "source."
  destination_prefix: "dest."
  schema:
    entities:
      - entity: KubernetesDeployment
        id:
          - sw.k8s.cluster.uid
          - k8s.namespace.name
          - k8s.deployment.name
      - entity: KubernetesStatefulSet
        id:
          - sw.k8s.cluster.uid
          - k8s.namespace.name
          - k8s.statefulset.name
      - entity: KubernetesDaemonSet
        id:
          - sw.k8s.cluster.uid
          - k8s.namespace.name
          - k8s.daemonset.name
      - entity: KubernetesJob
        id:
          - sw.k8s.cluster.uid
          - k8s.namespace.name
          - k8s.job.name
      - entity: KubernetesCronJob
        id:
          - sw.k8s.cluster.uid
          - k8s.namespace.name
          - k8s.cronjob.name
      - entity: KubernetesService
        id:
          - sw.k8s.cluster.uid
          - k8s.namespace.name
          - k8s.service.name
    events:
      relationships:
        # source KubernetesDeployment
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesDeployment
          destination_entity: KubernetesService
          attributes:
        # source KubernetesStatefulSet
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesStatefulSet
          destination_entity: KubernetesService
          attributes:
        # source KubernetesDaemonSet
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesDaemonSet
          destination_entity: KubernetesService
          attributes:
{{- end }}

{{- define "common-discovery-config.pipelines" -}}
{{- $context := index . 0 -}}
{{- $entryReceiver := index . 1 -}}
{{- $metricExporter := index . 2 -}}
metrics/discovery-scrape:
  exporters:
    - routing/discovered_metrics
  processors:
    - memory_limiter
{{- if $context.Values.otel.metrics.autodiscovery.prometheusEndpoints.filter }}
    - filter/metrics-discovery
{{- end }}
{{- if $context.Values.otel.metrics.autodiscovery.prefix }}
    - metricstransform/rename/discovery
{{- end }}
{{- if ne $context.Values.otel.metrics.autodiscovery.prefix "k8s." }}
    - metricstransform/copy-required-metrics
{{- end }}
  receivers:
    - {{ $entryReceiver }}

metrics/discovery-istio:
  exporters:
    - {{ $metricExporter }}
    - forward/relationship-state-events-workload-workload
    - forward/relationship-state-events-workload-service
  processors:
    - memory_limiter
    - swok8sworkloadtype/istio
    - transform/istio-metrics
    - metricstransform/istio-metrics
    - cumulativetodelta/istio-metrics
    - deltatorate/istio-metrics
    - metricsgeneration/istio-metrics
    - groupbyattrs/common-all
    - resource/all
  receivers:
    - routing/discovered_metrics

metrics/relationship-state-events-workload-workload-preparation:
  exporters:
    - solarwindsentity/istio-workload-workload
  processors:
    - memory_limiter
    - filter/keep-workload-workload-relationships
    - transform/istio-workload-workload
    - groupbyattrs/istio-relationships
    - transform/only-relationship-resource-attributes
  receivers:
    - forward/relationship-state-events-workload-workload

metrics/relationship-state-events-workload-service-preparation:
  exporters:
    - solarwindsentity/istio-workload-service
  processors:
    - memory_limiter
    - filter/keep-workload-service-relationships
    - transform/istio-workload-service
    - groupbyattrs/istio-relationships
    - transform/only-relationship-resource-attributes
  receivers:
    - forward/relationship-state-events-workload-service

logs/stateevents:
  exporters:
    - otlp
  processors:
    - memory_limiter
    - transform/scope
    - batch/stateevents
  receivers:
    - solarwindsentity/istio-workload-workload
    - solarwindsentity/istio-workload-service

metrics/discovery-custom:
  exporters:
    - {{ $metricExporter }}
  processors:
    - memory_limiter
{{- if $context.Values.otel.metrics.autodiscovery.prometheusEndpoints.customTransformations.counterToRate }}
    - cumulativetodelta/discovery
    - deltatorate/discovery
{{- end }}
    - groupbyattrs/common-all
    - resource/all
  receivers:
    - routing/discovered_metrics
{{- end }}
