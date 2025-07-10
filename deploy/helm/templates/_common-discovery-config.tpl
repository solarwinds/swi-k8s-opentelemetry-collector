{{- define "common-discovery-config.processors" -}}
{{- if .Values.otel.metrics.autodiscovery.prometheusEndpoints.filter }}
filter/metrics-discovery:
  metrics:
{{ toYaml .Values.otel.metrics.autodiscovery.prometheusEndpoints.filter | indent 4 }}
{{- end }}

logdedup/solarwindsentity: {}

filter/keep-entity-state-events:
  logs:
    log_record:
      - not(attributes["otel.entity.event.type"] == "entity_state")

filter/keep-relationship-state-events:
  logs:
    log_record:
      - not(attributes["otel.entity.event.type"] == "entity_relationship_state")

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
        - set(resource.attributes["istio"], "true") 

transform/istio-metric-datapoints:
  metric_statements:
    - statements:
        - set(datapoint.attributes["dest.sw.server.address.fqdn"], datapoint.attributes["destination_service"]) where metric.name == "{{ .Values.otel.metrics.autodiscovery.prefix }}istio_request_bytes_sum" and IsMatch(datapoint.attributes["destination_service"], "^(https?://)?[a-zA-Z0-9][-a-zA-Z0-9]*\\.[a-zA-Z0-9][-a-zA-Z0-9\\.]*(:\\d+)?$") and not(IsMatch(datapoint.attributes["destination_service"], ".*\\.cluster\\.local$")) and not(IsMatch(datapoint.attributes["destination_service"], "^(https?://)?\\d+\\.\\d+\\.\\d+\\.\\d+(:\\d+)?$"))

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
    - dest.sw.server.address.fqdn

filter/keep-workload-workload-relationships:
  error_mode: ignore
  metrics:
    datapoint:
      - datapoint.attributes["source_workload_type"] == nil or datapoint.attributes["destination_workload_type"] == nil or datapoint.attributes["source_workload_type"] == "" or datapoint.attributes["destination_workload_type"] == ""

filter/keep-workload-service-relationships:
  error_mode: ignore
  metrics:
    datapoint:
      - datapoint.attributes["source_workload_type"] == nil or datapoint.attributes["source_workload_type"] == "" or ((datapoint.attributes["destination_service_type"] == "" or datapoint.attributes["destination_service_type"] == nil) and (datapoint.attributes["dest.sw.server.address.fqdn"] == "" or datapoint.attributes["dest.sw.server.address.fqdn"] == nil))

# filter is used to keep only metrics that are not workload-to-workload or workload-to-service
filter/keep-not-relationships:
  error_mode: ignore
  metrics:
    datapoint:
      - not(datapoint.attributes["source_workload_type"] == nil or datapoint.attributes["destination_workload_type"] == nil or datapoint.attributes["source_workload_type"] == "" or datapoint.attributes["destination_workload_type"] == "" or ((datapoint.attributes["destination_service_type"] == "" or datapoint.attributes["destination_service_type"] == nil) and (datapoint.attributes["dest.sw.server.address.fqdn"] == "" or datapoint.attributes["dest.sw.server.address.fqdn"] == nil)))



transform/istio-workload-workload:
  metric_statements:
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
    - set(datapoint.attributes["source.k8s.deployment.name"], datapoint.attributes["source_workload"]) where datapoint.attributes["source_workload_type"] == "Deployment"
    - set(datapoint.attributes["source.k8s.statefulset.name"], datapoint.attributes["source_workload"]) where datapoint.attributes["source_workload_type"] == "StatefulSet"
    - set(datapoint.attributes["source.k8s.daemonset.name"], datapoint.attributes["source_workload"]) where datapoint.attributes["source_workload_type"] == "DaemonSet"
    - set(datapoint.attributes["source.k8s.namespace.name"], datapoint.attributes["source_workload_namespace"])
    - set(datapoint.attributes["dest.k8s.service.name"], datapoint.attributes["destination_service_name"]) where datapoint.attributes["destination_service_type"] == "Service"
    - set(datapoint.attributes["dest.k8s.namespace.name"], datapoint.attributes["destination_service_namespace"])

transform/only-relationship-resource-attributes:
  metric_statements:
    # Temporary, to be removed when solarwindsentityconnector supports creation of entities from attributes with prefixes
    - set(resource.attributes["sw.server.address.fqdn"], resource.attributes["dest.sw.server.address.fqdn"]) where resource.attributes["dest.sw.server.address.fqdn"] != nil

batch/stateevents:
  send_batch_size: 1024
  timeout: 1s
  send_batch_max_size: 1024

resource/clean-temporary-attributes:
  attributes:      
    - key: istio
      action: delete
{{- end }}


{{- define "common-discovery-config.connectors" -}}
forward/relationship-state-events-workload-workload: {}
forward/relationship-state-events-workload-service: {}
forward/not-relationship-state-events: {}
forward/discovery-istio-metrics-clean: {}
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
          conditions: []
          context: "metric"
          attributes:
            - istio
          action: "update"
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesDeployment
          destination_entity: KubernetesStatefulSet
          conditions: []
          context: "metric"
          attributes:
            - istio
          action: "update"
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesDeployment
          destination_entity: KubernetesDaemonSet
          conditions: []
          context: "metric"
          attributes:
            - istio
          action: "update"
        # source KubernetesStatefulSet
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesStatefulSet
          destination_entity: KubernetesDeployment
          conditions: []
          context: "metric"
          attributes:
            - istio
          action: "update"
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesStatefulSet
          destination_entity: KubernetesStatefulSet
          conditions: []
          context: "metric"
          attributes:
            - istio
          action: "update"
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesStatefulSet
          destination_entity: KubernetesDaemonSet
          conditions: []
          context: "metric"
          attributes:
            - istio
          action: "update"
        # source KubernetesDaemonSet
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesDaemonSet
          destination_entity: KubernetesDeployment
          conditions: []
          context: "metric"
          attributes:
            - istio
          action: "update"
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesDaemonSet
          destination_entity: KubernetesStatefulSet
          conditions: []
          context: "metric"
          attributes:
            - istio
          action: "update"
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesDaemonSet
          destination_entity: KubernetesDaemonSet
          conditions: []
          context: "metric"
          attributes:
            - istio
          action: "update"


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
      - entity: PublicNetworkLocation
        id:
          - sw.server.address.fqdn
    events:
      entities:
        - type: "PublicNetworkLocation"
          context: "metric"
          action: "update"
      relationships:
        # source KubernetesDeployment
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesDeployment
          destination_entity: KubernetesService
          conditions: []
          context: "metric"
          attributes:
            - istio
          action: "update"
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesDeployment
          destination_entity: PublicNetworkLocation
          conditions: []
          context: "metric"
          attributes:
            - istio
          action: "update"
        # source KubernetesStatefulSet
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesStatefulSet
          destination_entity: KubernetesService
          conditions: []
          context: "metric"
          attributes:
            - istio
          action: "update"
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesStatefulSet
          destination_entity: PublicNetworkLocation
          conditions: []
          context: "metric"
          attributes:
            - istio
          action: "update"
        # source KubernetesDaemonSet
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesDaemonSet
          destination_entity: KubernetesService
          conditions: []
          context: "metric"
          attributes:
            - istio
          action: "update"
        - type: KubernetesCommunicatesWith
          source_entity: KubernetesDaemonSet
          destination_entity: PublicNetworkLocation
          conditions: []
          context: "metric"
          attributes:
            - istio
          action: "update"
{{- end }}

{{- define "common-discovery-config.pipelines" -}}
{{- $context := index . 0 -}}
{{- $entryReceiver := index . 1 -}}
{{- $metricExporter := index . 2 -}}
metrics/discovery-scrape:
  receivers:
    - {{ $entryReceiver }}
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
  exporters:
    - routing/discovered_metrics

metrics/discovery-istio:
  receivers:
    - routing/discovered_metrics
  processors:
    - memory_limiter
    - swok8sworkloadtype/istio
    - transform/istio-metrics
    - transform/istio-metric-datapoints
    - metricstransform/istio-metrics
    - cumulativetodelta/istio-metrics
    - deltatorate/istio-metrics
    - metricsgeneration/istio-metrics
    - groupbyattrs/common-all
    - resource/all
  exporters:
    - forward/relationship-state-events-workload-workload
    - forward/relationship-state-events-workload-service
    - forward/not-relationship-state-events

metrics/relationship-state-events-workload-workload-preparation:
  receivers:
    - forward/relationship-state-events-workload-workload
  processors:
    - memory_limiter
    - filter/keep-workload-workload-relationships
    - transform/istio-workload-workload
    - groupbyattrs/istio-relationships
    - transform/only-relationship-resource-attributes
  exporters:
    - forward/discovery-istio-metrics-clean
    - solarwindsentity/istio-workload-workload

metrics/relationship-state-events-workload-service-preparation:
  receivers:
    - forward/relationship-state-events-workload-service
  processors:
    - memory_limiter
    - filter/keep-workload-service-relationships
    - transform/istio-workload-service
    - groupbyattrs/istio-relationships
    - transform/only-relationship-resource-attributes
  exporters:
    - forward/discovery-istio-metrics-clean
    - solarwindsentity/istio-workload-service

metrics/not-relationship-state-events-preparation:
  receivers:
    - forward/not-relationship-state-events
  processors:
    - memory_limiter
    - filter/keep-not-relationships
  exporters:
    - forward/discovery-istio-metrics-clean

metrics/discovery-istio-clean:
  receivers:
    - forward/discovery-istio-metrics-clean
  processors:
    - memory_limiter
    - resource/clean-temporary-attributes
  exporters:
    - {{ $metricExporter }}

# Current SWO pipeline cannot process state events and relationships events together,
# so we need to split them into two separate pipelines.
# TODO - merge them into one pipeline when SWO supports it.
logs/stateevents-entities:
  receivers:
    - solarwindsentity/istio-workload-workload
    - solarwindsentity/istio-workload-service
  processors:
    - memory_limiter
    - filter/keep-entity-state-events
    - transform/scope
    - logdedup/solarwindsentity
    - batch/stateevents
  exporters:
    - otlp

logs/stateevents-relationships:
  receivers:
    - solarwindsentity/istio-workload-workload
    - solarwindsentity/istio-workload-service
  processors:
    - memory_limiter
    - filter/keep-relationship-state-events
    - transform/scope
    - logdedup/solarwindsentity
    - batch/stateevents
  exporters:
    - otlp

metrics/discovery-custom:
  receivers:
    - routing/discovered_metrics
  processors:
    - memory_limiter
{{- if $context.Values.otel.metrics.autodiscovery.prometheusEndpoints.customTransformations.counterToRate }}
    - cumulativetodelta/discovery
    - deltatorate/discovery
{{- end }}
    - groupbyattrs/common-all
    - resource/all
  exporters:
    - {{ $metricExporter }}
{{- end }}
