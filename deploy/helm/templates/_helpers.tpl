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
*/}}
{{- define "common.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Common template labels
*/}}
{{- define "common.template-labels" -}}
app.kubernetes.io/name: {{ template "common.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "common.labels" -}}
{{ include "common.template-labels" . }}
{{- if .Chart.AppVersion }}
helm.sh/chart: {{ include "common.chart" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
{{- if .Values.commonLabels}}
{{ toYaml .Values.commonLabels }}
{{- end }}
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
{{- if index . 2 }}
    annotations:
      - key_regex: (.*)
        tag_name: k8s.{{ index . 1 }}.annotations.$$1
        from: {{ index . 1 }}
{{- end }}
{{- if index . 3 }}
    labels:
      - key_regex: (.*)
        tag_name: k8s.{{ index . 1 }}.labels.$$1
        from: {{ index . 1 }}
{{- end }}
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
{{- if index . 2 }}
    annotations:
      - key_regex: (.*)
        tag_name: k8s.{{ index . 1 }}.annotations.$$1
        from: {{ index . 1 }}
{{- end }}
{{- if index . 3 }}
    labels:
      - key_regex: (.*)
        tag_name: k8s.{{ index . 1 }}.labels.$$1
        from: {{ index . 1 }}
{{- end }}
  association:
  - sources:
      - from: resource_attribute
        name: k8s.{{ index . 1 }}.name
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
{{- if index . 1 }}
  annotations:
    - key_regex: (.*)
      tag_name: k8s.pod.annotations.$$1
      from: pod
    - key_regex: (.*)
      tag_name: k8s.namespace.annotations.$$1
      from: namespace
{{- end }}
{{- if index . 2 }}
  labels:
    - key_regex: (.*)
      tag_name: k8s.pod.labels.$$1
      from: pod
    - key_regex: (.*)
      tag_name: k8s.namespace.labels.$$1
      from: namespace
{{- end }}
pod_association:
  - sources:
      - from: resource_attribute
        name: k8s.pod.name
      - from: resource_attribute
        name: k8s.namespace.name
{{ include "common.k8s-instrumentation.resource.namespaced" (tuple . "deployment" (index . 1) (index . 2)) }}
{{ include "common.k8s-instrumentation.resource.namespaced" (tuple . "statefulset" (index . 1) (index . 2)) }}
{{ include "common.k8s-instrumentation.resource.namespaced" (tuple . "replicaset" (index . 1) (index . 2)) }}
{{ include "common.k8s-instrumentation.resource.namespaced" (tuple . "daemonset" (index . 1) (index . 2)) }}
{{ include "common.k8s-instrumentation.resource.namespaced" (tuple . "job" (index . 1) (index . 2)) }}
{{ include "common.k8s-instrumentation.resource.namespaced" (tuple . "cronjob" (index . 1) (index . 2)) }}
{{ include "common.k8s-instrumentation.resource" (tuple . "persistentvolume" (index . 1) (index . 2)) }}
{{ include "common.k8s-instrumentation.resource.namespaced" (tuple . "persistentvolumeclaim" (index . 1) (index . 2)) }}
{{ include "common.k8s-instrumentation.resource" (tuple . "node" (index . 1) (index . 2)) }}
{{- end -}}