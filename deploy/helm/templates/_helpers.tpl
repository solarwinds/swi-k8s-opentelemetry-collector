{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "common.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
And depending on the resources the name is completed with an extension.
If release name contains chart name it will be used as a full name.
Usages: 
  no suffix: {{ include "common.fullname" . }}
  with suffix: {{ include "common.fullname" (tuple . "-node-collector") }}
  with custom max length: {{ include "common.fullname" (tuple . "-node-collector" 50) }}
*/}}
{{- define "common.fullname" -}}
{{- $context := . -}}
{{- $suffix := "" -}}
{{- $defaultMaxLength := 63 -}}
{{- $maxLength := $defaultMaxLength -}}
{{- if eq (kindOf .) "slice" -}}
  {{- $context = index . 0 -}}
  {{- $suffix = index . 1 | default "" -}}
  {{- if gt (len .) 2 -}}
    {{- $paramMax := index . 2 | default $defaultMaxLength -}}
    {{- $maxLength = sub $paramMax (len $suffix) -}}
  {{- else -}}
    {{- $maxLength = sub $defaultMaxLength (len $suffix) -}}
  {{- end -}}
{{- end -}}

{{- $maxLengthStr := printf "%d" $maxLength -}}
{{- $maxLengthInt := $maxLengthStr | atoi -}}
{{- $releaseNameMax := int (div $maxLengthInt 2) -}}
{{- $releaseName := $context.Release.Name | trunc $releaseNameMax | trimSuffix "-" -}}
{{- $result := "" -}}

{{- if $context.Values.fullnameOverride -}}
  {{- $result = $context.Values.fullnameOverride | trunc $maxLengthInt | trimSuffix "-" -}}
{{- else -}}
  {{- $name := default $context.Chart.Name $context.Values.nameOverride -}}
  {{- if contains $name $releaseName -}}
    {{- $result = $releaseName | trunc $maxLengthInt | trimSuffix "-" -}}
  {{- else -}}
    {{- $result = printf "%s-%s" $releaseName $name | trunc $maxLengthInt | trimSuffix "-" -}}
  {{- end -}}
{{- end -}}
{{- printf "%s%s" $result $suffix -}}
{{- end -}}

{{/*
Get cluster UID based on name and uid provided in .Values.cluster.
Usage:
  {{ include "common.cluster-uid" . }}
*/}}
{{- define "common.cluster-uid" -}}
{{ default .Values.cluster.name .Values.cluster.uid }}
{{- end -}}

{{/*
Common pod labels - those labels are included on every pod in the chart
*/}}
{{- define "common.pod-labels" -}}
{{- if .Values.aks }}
azure-extensions-usage-release-identifier: {{ .Release.Name }}
{{- end -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "common.labels" -}}
app.kubernetes.io/part-of: swo-k8s-collector
app.kubernetes.io/instance: {{ template "common.fullname" . }}
app.kubernetes.io/managed-by: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
helm.sh/chart: {{ include "common.chart" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
{{- if .Values.commonLabels}}
{{ toYaml .Values.commonLabels }}
{{- end }}
{{- end -}}

{{/*
Common annotations
*/}}
{{- define "common.annotations" -}}
swo.cloud.solarwinds.com/cluster-uid: {{ (include "common.cluster-uid" .) }}
{{- end -}}

{{/*
Event which are considered as error
*/}}
{{- define "common.events-error-conditions" -}}
log.attributes["k8s.event.reason"] == "Failed"
or log.attributes["k8s.event.reason"] == "BackOff"
or log.attributes["k8s.event.reason"] == "FailedKillPod"
or log.attributes["k8s.event.reason"] == "FailedCreatePodContainer"
or log.attributes["k8s.event.reason"] == "NetworkNotReady"
or log.attributes["k8s.event.reason"] == "InspectFailed"
or log.attributes["k8s.event.reason"] == "ErrImageNeverPull"
or log.attributes["k8s.event.reason"] == "NodeNotReady"
or log.attributes["k8s.event.reason"] == "NodeNotSchedulable"
or log.attributes["k8s.event.reason"] == "KubeletSetupFailed"
or log.attributes["k8s.event.reason"] == "FailedAttachVolume"
or log.attributes["k8s.event.reason"] == "FailedMount"
or log.attributes["k8s.event.reason"] == "VolumeResizeFailed"
or log.attributes["k8s.event.reason"] == "FileSystemResizeFailed"
or log.attributes["k8s.event.reason"] == "FailedMapVolume"
or log.attributes["k8s.event.reason"] == "ContainerGCFailed"
or log.attributes["k8s.event.reason"] == "ImageGCFailed"
or log.attributes["k8s.event.reason"] == "FailedNodeAllocatableEnforcement"
or log.attributes["k8s.event.reason"] == "FailedCreatePodSandBox"
or log.attributes["k8s.event.reason"] == "FailedPodSandBoxStatus"
or log.attributes["k8s.event.reason"] == "FailedMountOnFilesystemMismatch"
or log.attributes["k8s.event.reason"] == "InvalidDiskCapacity"
or log.attributes["k8s.event.reason"] == "FreeDiskSpaceFailed"
or log.attributes["k8s.event.reason"] == "FailedSync"
or log.attributes["k8s.event.reason"] == "FailedValidation"
or log.attributes["k8s.event.reason"] == "FailedPostStartHook"
or log.attributes["k8s.event.reason"] == "FailedPreStopHook"
{{- end -}}

{{/*
Event which are considered as warning
*/}}
{{- define "common.events-warning-conditions" -}}
log.attributes["k8s.event.reason"] == "ProbeWarning"
or log.attributes["k8s.event.reason"] == "Unhealthy"
{{- end -}}

{{- define "common.k8s-instrumentation" -}}
auth_type: "serviceAccount"
passthrough: false
extract:
  metadata:
    - k8s.deployment.name
    - k8s.replicaset.name
    - k8s.daemonset.name
    - k8s.job.name
    - k8s.cronjob.name
    - k8s.statefulset.name
    - k8s.node.name
pod_association:
  - sources:
      - from: resource_attribute
        name: k8s.pod.name
      - from: resource_attribute
        name: k8s.namespace.name
{{- end -}}

{{/*
common.image - Helper template to determine the image path based on various conditions.

Usage:
{{ include "common.image" (tuple $root $path $nameObj $defaultFullImage $defaultTag) }}

Where:
- $root: The root context of the chart (usually passed as '.' from the calling template).
- $path: The path within .Values where the image information is located.
- $nameObj: The key name for the image configuration. This can be either a string or a slice.
  - If a string, it is used as the key name for both otel and Azure image configurations.
  - If a slice, it expects two elements:
    - The first element is the key name for the otel image configuration.
    - The second element is the key name for the Azure image configuration.
- $defaultFullImage: (Optional) A default image (including tag) to use if the specified image is not found. 
  - Expected format: "repository/image:tag".
- $defaultTag: (Optional) A default tag to use if no tag is specified in the image configuration.

Details:
- The template first checks if $nameObj is a slice to handle different keys for otel and Azure configurations.
- It then prepares the image path based on whether the chart is configured to use Azure (`$root.Values.aks`) or custom settings.
- For Azure configurations, it constructs the image path using Azure registry, image, and digest details.
- For non-Azure configurations, it uses the repository and tag from the specified path in .Values, falling back to default values if necessary.

Example:
{{ include "common.image" (tuple . .Values.otel "image" "myrepo/myimage:v1.0.0" "v1.0.0") }}
- This example uses "image" as the key for both otel and Azure configurations, with default image "myrepo/myimage:v1.0.0" and default tag "v1.0.0".
*/}}
{{- define "common.image" -}}
{{- $root := index . 0 -}}
{{- $path := index . 1 -}}
{{- $nameObj := index . 2 -}}
{{- $name := "" -}}
{{- $azureName := "" -}}
{{- if eq (kindOf $nameObj) "slice" -}}
  {{- $name = index $nameObj 0 -}}
  {{- $azureName = index $nameObj 1 -}}
{{- else -}}
  {{- $name = $nameObj -}}
  {{- $azureName = $nameObj -}}
{{- end -}}

{{- $defaultFullImage := "" -}}
{{- $defaultImage := "" -}}
{{- $defaultTag := "" -}}
{{- if gt (len .) 3 -}}
  {{- $defaultFullImage = index . 3 -}}
{{- end -}}

{{- if gt (len .) 4 -}}
  {{- $defaultTag = index . 4 -}}
{{- end -}}

{{- if $defaultFullImage -}}
  {{- $defaultImageParts := split ":" $defaultFullImage -}}
  {{- $defaultImage = $defaultImageParts._0 -}}
  {{- if gt (len $defaultImageParts) 1 -}}
    {{- $defaultTag = $defaultImageParts._1 -}}
  {{- end -}}
{{- end -}}

{{- $azure := false -}}
{{- if and $root.Values.aks $root.Values.global -}}
{{- if $root.Values.global.azure -}}
{{- if $root.Values.global.azure.images -}}
{{- if index $root.Values.global.azure.images $azureName -}}
  {{- $azure = true -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- if $azure -}}
  {{- $azurePath := $root.Values.global.azure.images -}}
  {{- $azureImageObj := index $azurePath $azureName -}}
  {{- $azureDigest := index $azureImageObj "digest" -}}
  {{- $azureImage := index $azureImageObj "image" -}}
  {{- $azureRegistry := index $azureImageObj "registry" -}}
  {{- printf "%s/%s@%s" $azureRegistry $azureImage $azureDigest -}}
{{- else -}}
  {{- $valuesPath := index $path $name -}}
  {{- $valuesRepository := index $valuesPath "repository" -}}
  {{- if eq $valuesRepository "solarwinds/swi-opentelemetry-collector" -}}
    {{- $valuesRepository = "solarwinds/solarwinds-otel-collector" -}}
  {{- end -}}
  {{- $valuesTag := index $valuesPath "tag" -}}
  {{- $repo := $valuesRepository | default $defaultImage -}}
  {{- $tag := $valuesTag | default $defaultTag -}}
  {{- if $tag -}}
    {{- printf "%s:%s" $repo $tag -}}
  {{- else -}}
    {{- printf "%s" $repo -}}
  {{- end -}}
{{- end -}}

{{- end -}}

{{/*
Define name for the Secret
*/}}
{{- define "common.secret" -}}
{{- if .Values.otel.api_token }}
{{- include "common.fullname" (tuple . "-secret") }}
{{- else }}
{{- "solarwinds-api-token" }}
{{- end }}
{{- end -}}

{{/*
Check the used filtering version

Usage:
{{ isDeprecatedFilterSyntax (.Values.otel.events.filter) }}
*/}}
{{- define "isDeprecatedFilterSyntax" -}}
{{- if . -}}
{{- if or (index . "include") (index . "exclude") -}}
true
{{- else -}}
false
{{- end -}}
{{- else -}}
false
{{- end -}}
{{- end -}}

{{- define "defaultDeprecatedLogsFilter" -}}
include:
  match_type: regexp
  # a log has to match all expressions in the list to be included
  # see https://github.com/google/re2/wiki/Syntax for regexp syntax
  record_attributes:
    # allow only system namespaces (kube-system, kube-public)
    - key: k8s.namespace.name
      value: ^kube-.*$
{{- end }}

{{- define "defaultLogsFilter" -}}
{{- end }}

{{/*
Get the log filter.
The filter is a merge from the default filter and the user defined one.
The default filter's syntax is chosen based on the syntax of the user defined filter.

Usage:
{{ include "logsFilter" . }}

Returns:
YAML with the filter.
*/}}
{{- define "logsFilter" -}}

{{- $defaultFilter := (include "defaultLogsFilter" .) -}}
{{- if eq (include "isDeprecatedFilterSyntax" .Values.otel.logs.filter) "true" -}}
{{- $defaultFilter = (include "defaultDeprecatedLogsFilter" .) -}}
{{- end -}}

{{- $filter := dict -}}
{{- if .Values.otel.logs.filter -}}
{{- $filter = deepCopy .Values.otel.logs.filter -}}
{{- end -}}

{{- if or $defaultFilter $filter -}}
{{- merge $filter (fromYaml $defaultFilter) | toYaml -}}
{{- end -}}

{{- end -}}

{{/*
Check whether the SWI endpoint check is enabled

Usage:
{{ isSwiEndpointCheckEnabled . }}
*/}}
{{- define "isSwiEndpointCheckEnabled" -}}
{{- ternary "true" "" (and .Values.otel.swi_endpoint_check.enabled (ternary true .Values.otel.metrics.swi_endpoint_check (eq .Values.otel.metrics.swi_endpoint_check nil))) -}}
{{- end -}}


{{/*
Check whether namespace filter is enabled

Usage:
{{- if eq (include "isNamespacesFilterEnabled" .) "true" }}
*/}}
{{- define "isNamespacesFilterEnabled" -}}
{{- if or (not (empty .Values.cluster.filter.exclude_namespaces)) (not (empty .Values.cluster.filter.exclude_namespaces_regex)) (not (empty .Values.cluster.filter.include_namespaces)) (not (empty .Values.cluster.filter.include_namespaces_regex)) -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Returns namespace filters in filter processor's format

Usage:
{{- include "namespacesFilter" . | nindent 8 }}
*/}}
{{- define "namespacesFilter" -}}
{{- range .Values.cluster.filter.exclude_namespaces }}
- resource.attributes["k8s.namespace.name"] == "{{ . }}"
{{- end }}
{{- range .Values.cluster.filter.exclude_namespaces_regex }}
- IsMatch(resource.attributes["k8s.namespace.name"], "{{ . }}")
{{- end }}
# include namespaces have to be merged to one condition with ORs
{{- if or (not (empty .Values.cluster.filter.include_namespaces)) (not (empty .Values.cluster.filter.include_namespaces_regex)) -}}
{{- $conditions := list }}
{{- range .Values.cluster.filter.include_namespaces }}
  {{- $value := . }}
  {{- $condition := printf `resource.attributes["k8s.namespace.name"] == "%s"` $value }}
  {{- $conditions = append $conditions $condition }}
{{- end }}
{{- range .Values.cluster.filter.include_namespaces_regex }}
  {{- $value := . }}
  {{- $condition := printf `IsMatch(resource.attributes["k8s.namespace.name"], "%s")` $value }}
  {{- $conditions = append $conditions $condition }}
{{- end }}
{{- $conditions = append $conditions (printf `resource.attributes["k8s.namespace.name"] == nil`) }}
{{- $joinedConditions := join " or " $conditions }}
- not({{ $joinedConditions }}) 
{{- end -}}
{{- end -}}

{{/*
Calculate max_staleness for cumulativetodelta processor
If max_staleness is explicitly set, use it. Otherwise, default to twice the scrape_interval
Input should be the Values object with prometheus configuration
Output will be a duration string like "120s", "4m", etc.
*/}}
{{- define "common.maxStaleness" -}}
{{- if .Values.otel.metrics.prometheus.max_staleness -}}
  {{- .Values.otel.metrics.prometheus.max_staleness -}}
{{- else -}}
  {{- include "common.doubleInterval" .Values.otel.metrics.prometheus.scrape_interval -}}
{{- end -}}
{{- end -}}

{{/*
Calculate max_staleness as twice the scrape_interval
Input should be a duration string like "60s", "5m", etc.
Output will be the same format but doubled
*/}}
{{- define "common.doubleInterval" -}}
{{- $interval := . -}}
{{- if hasSuffix "s" $interval -}}
  {{- $num := $interval | trimSuffix "s" | int -}}
  {{- printf "%ds" (mul $num 2) -}}
{{- else if hasSuffix "m" $interval -}}
  {{- $num := $interval | trimSuffix "m" | int -}}
  {{- printf "%dm" (mul $num 2) -}}
{{- else if hasSuffix "h" $interval -}}
  {{- $num := $interval | trimSuffix "h" | int -}}
  {{- printf "%dh" (mul $num 2) -}}
{{- else -}}
  {{- /* Default case - assume seconds if no suffix */ -}}
  {{- $num := $interval | int -}}
  {{- printf "%ds" (mul $num 2) -}}
{{- end -}}
{{- end -}}

{{- define "common.prometheus.relabelconfigs" -}}
metric_relabel_configs:
  - source_labels: [service_name]
    regex: (.+)
    target_label: job
    replacement: $1
    action: replace
  - regex: ^service_name$
    action: labeldrop
{{- end -}}