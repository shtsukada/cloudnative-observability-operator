{{/*
Common template helpers
*/}}

{{/* Chart name */}}
{{- define "cloudnative-observability-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Fully qualified release name */}}
{{- define "cloudnative-observability-operator.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name (include "cloudnative-observability-operator.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/* Chart label */}}
{{- define "cloudnative-observability-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name (.Chart.Version | replace "+" "_") | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Standard labels */}}
{{- define "cloudnative-observability-operator.labels" -}}
helm.sh/chart: {{ include "cloudnative-observability-operator.chart" . }}
app.kubernetes.io/name: {{ include "cloudnative-observability-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Chart.AppVersion }}
app.kubernetes.io/version: {{ . | quote }}
{{- end }}
{{- end -}}

{{/* Selector labels (must be immutable) */}}
{{- define "cloudnative-observability-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "cloudnative-observability-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/* ServiceAccount name */}}
{{- define "cloudnative-observability-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "cloudnative-observability-operator.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}
