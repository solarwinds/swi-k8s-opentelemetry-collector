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
{{- if not .Values.externalRenderer}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "common.labels" -}}
{{ include "common.template-labels" . }}
{{- if .Chart.AppVersion }}
{{- if not .Values.externalRenderer}}
helm.sh/chart: {{ include "common.chart" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
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