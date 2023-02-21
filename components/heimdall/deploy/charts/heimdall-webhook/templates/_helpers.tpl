{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "nasp-webhook.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "nasp-webhook.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "nasp-webhook.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}


{{- define "nasp-webhook-controller.fullname" -}}
{{ include "nasp-webhook.fullname" . }}-controller
{{- end }}

{{- define "nasp-webhook-controller.name" -}}
{{ include "nasp-webhook.name" . }}-controller
{{- end }}

{{- define "nasp-webhook-controller.labels" }}
app: {{ include "nasp-webhook-controller.fullname" . }}
app.kubernetes.io/name: {{ include "nasp-webhook-controller.name" . }}
helm.sh/chart: {{ include "nasp-webhook.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | replace "+" "_" }}
app.kubernetes.io/component: nasp-webhook-controller
app.kubernetes.io/part-of: {{ include "nasp-webhook.name" . }}
{{- end }}

{{- define "nasp-webhook-controller.selectorLabels" -}}
app.kubernetes.io/name: {{ include "nasp-webhook-controller.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
