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
*/}}
{{- define "common.fullname" -}}
{{- $context := . -}}
{{- $suffix := "" -}}
{{- $maxLength := 63 -}}
{{- if eq (kindOf .) "slice" -}}
{{- $context = index . 0 -}}
{{- $suffix = index . 1 | default "" -}}
{{- $maxLength = sub 63 (len $suffix) -}}
{{- end -}}

{{- $maxLengthStr := printf "%d" $maxLength -}}
{{- $maxLengthInt := $maxLengthStr | atoi -}}
{{- $releaseName := $context.Release.Name | trunc 30 | trimSuffix "-" -}}
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
attributes["k8s.event.reason"] == "Failed"
or attributes["k8s.event.reason"] == "BackOff"
or attributes["k8s.event.reason"] == "FailedKillPod"
or attributes["k8s.event.reason"] == "FailedCreatePodContainer"
or attributes["k8s.event.reason"] == "NetworkNotReady"
or attributes["k8s.event.reason"] == "InspectFailed"
or attributes["k8s.event.reason"] == "ErrImageNeverPull"
or attributes["k8s.event.reason"] == "NodeNotReady"
or attributes["k8s.event.reason"] == "NodeNotSchedulable"
or attributes["k8s.event.reason"] == "KubeletSetupFailed"
or attributes["k8s.event.reason"] == "FailedAttachVolume"
or attributes["k8s.event.reason"] == "FailedMount"
or attributes["k8s.event.reason"] == "VolumeResizeFailed"
or attributes["k8s.event.reason"] == "FileSystemResizeFailed"
or attributes["k8s.event.reason"] == "FailedMapVolume"
or attributes["k8s.event.reason"] == "ContainerGCFailed"
or attributes["k8s.event.reason"] == "ImageGCFailed"
or attributes["k8s.event.reason"] == "FailedNodeAllocatableEnforcement"
or attributes["k8s.event.reason"] == "FailedCreatePodSandBox"
or attributes["k8s.event.reason"] == "FailedPodSandBoxStatus"
or attributes["k8s.event.reason"] == "FailedMountOnFilesystemMismatch"
or attributes["k8s.event.reason"] == "InvalidDiskCapacity"
or attributes["k8s.event.reason"] == "FreeDiskSpaceFailed"
or attributes["k8s.event.reason"] == "FailedSync"
or attributes["k8s.event.reason"] == "FailedValidation"
or attributes["k8s.event.reason"] == "FailedPostStartHook"
or attributes["k8s.event.reason"] == "FailedPreStopHook"
{{- end -}}

{{/*
Event which are considered as warning
*/}}
{{- define "common.events-warning-conditions" -}}
attributes["k8s.event.reason"] == "ProbeWarning"
or attributes["k8s.event.reason"] == "Unhealthy"
{{- end -}}

{{- define "common.k8s-instrumentation.resource.namespaced" -}}
{{ index . 1 }}:
  extract:
  association:
  - sources:
      - from: resource_attribute
        name: k8s.{{ index . 1 }}.name
      - from: resource_attribute
        name: k8s.namespace.name
{{- end -}}

{{- define "common.k8s-instrumentation.resource" -}}
{{ index . 1 }}:
  extract:
  association:
  - sources:
      - from: resource_attribute
        name: k8s.{{ index . 1 }}.name
{{- end -}}

{{- define "common.k8s-instrumentation" -}}
auth_type: "serviceAccount"
passthrough: false
set_object_existence: true
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
{{ include "common.k8s-instrumentation.resource.namespaced" (tuple . "deployment") }}
{{ include "common.k8s-instrumentation.resource.namespaced" (tuple . "statefulset") }}
{{ include "common.k8s-instrumentation.resource.namespaced" (tuple . "replicaset") }}
{{ include "common.k8s-instrumentation.resource.namespaced" (tuple . "daemonset") }}
{{ include "common.k8s-instrumentation.resource.namespaced" (tuple . "job") }}
{{ include "common.k8s-instrumentation.resource.namespaced" (tuple . "cronjob") }}
{{ include "common.k8s-instrumentation.resource" (tuple . "persistentvolume") }}
{{ include "common.k8s-instrumentation.resource.namespaced" (tuple . "persistentvolumeclaim") }}
{{ include "common.k8s-instrumentation.resource" (tuple . "node") }}
{{ include "common.k8s-instrumentation.resource.namespaced" (tuple . "service") }}
{{- end -}}

{{/*
common.image - Helper template to determine the image path based on various conditions.

Usage:
  {{ include "common.image" (tuple $root $path $nameObj $defaultFullImage $defaultTag $defaultKey) }}

Where:
- $root: The root context of the chart (usually passed as '.' from the calling template).
- $path: The path within .Values where the image information is located.
- $nameObj: The key name for the image configuration. This can be either a string or a slice.
  - If a string, it is used as the key name for both otel and Azure image configurations.
  - If a slice, it expects two elements:
    - The first element is the key name for the otel image configuration.
    - The second element is the key name for the Azure image configuration.
- $defaultKey: A unique key that is used for retrieving default values of the image if autoupdate is enabled.
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

{{- $defaultKey := "" -}}
{{- if gt (len .) 3 -}}
  {{- $defaultKey = index . 3 -}}
{{- end -}}

{{- $defaultFullImage := "" -}}
{{- $defaultImage := "" -}}
{{- $defaultTag := "" -}}
{{- if gt (len .) 4 -}}
  {{- $defaultFullImage = index . 4 -}}
{{- end -}}
{{- if gt (len .) 5 -}}
  {{- $defaultTag = index . 5 -}}
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
  {{- $valuesTag := index $valuesPath "tag" -}}
  {{- $valuesPullPolicy := index $valuesPath "pullPolicy" -}}

  {{- if $root.Values.autoupdate.enabled }}
    {{- $defaultImagesJson := include "common.defaultImages" $root -}}
    {{- $defaultImages := fromJson $defaultImagesJson -}}
    {{- $defaultImage := index $defaultImages $defaultKey -}}
    {{- $validRepository := index $defaultImage "repository" -}}
    {{- $validTag := index $defaultImage "tag" -}}
    {{- $validPullPolicy := index $defaultImage "pullPolicy" | default "IfNotPresent" -}}

    {{- if and (ne $defaultKey "") $defaultImage }}
      {{- if or (ne $valuesRepository $validRepository) (ne $valuesTag $validTag) (ne $valuesPullPolicy $validPullPolicy) }}
        {{- fail (printf "Autoupdate is enabled, but the default image was changed for key '%s' (repo|tag|pullPolicy differ)." $defaultKey) -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
  
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
Backup validation set of default images from values.yaml for checking if user has changed something.
*/}}
{{- define "common.defaultImages" -}}
{
  "autoupdate.image": {
    "repository": "alpine/k8s",
    "tag": "1.27.16",
    "pullPolicy": "IfNotPresent"
  },
  "otel.image": {
    "repository": "solarwinds/swi-opentelemetry-collector",
    "tag": "",
    "pullPolicy": "IfNotPresent"
  },
  "otel.init_images.busy_box": {
    "repository": "busybox",
    "tag": "1.36.1",
    "pullPolicy": "IfNotPresent"
  },
  "otel.windows.image": {
    "repository": "solarwinds/swi-opentelemetry-collector",
    "tag": "",
    "pullPolicy": "IfNotPresent"
  },
  "swoagent.image": {
    "repository": "solarwinds/swo-agent",
    "tag": "v2.10.68",
    "pullPolicy": "IfNotPresent"
  },
  "ebpfNetworkMonitoring.k8sCollector.watcher.image": {
    "repository": "solarwinds/opentelemetry-ebpf-k8s-watcher",
    "tag": "v0.10.3",
    "pullPolicy": "IfNotPresent"
  },
  "ebpfNetworkMonitoring.k8sCollector.relay.image": {
    "repository": "solarwinds/opentelemetry-ebpf-k8s-relay",
    "tag": "v0.10.3",
    "pullPolicy": "IfNotPresent"
  },
  "ebpfNetworkMonitoring.reducer.image": {
    "repository": "solarwinds/opentelemetry-ebpf-reducer",
    "tag": "v0.10.3",
    "pullPolicy": "IfNotPresent"
  },
  "ebpfNetworkMonitoring.kernelCollector.image": {
    "repository": "solarwinds/opentelemetry-ebpf-kernel-collector",
    "tag": "v0.10.3",
    "pullPolicy": "IfNotPresent"
  }
}
{{- end -}}
