{{- define "microservice.fullname" -}}
{{- printf "%s-%s" .Release.Name .Values.componentName | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "microservice.databaseName" -}}
{{- printf "%s-db" (include "microservice.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "microservice.secretName" -}}
{{- printf "%s-secrets" (include "microservice.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "microservice.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
app.kubernetes.io/name: {{ .Values.componentName }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/part-of: online-store
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "microservice.selectorLabels" -}}
app.kubernetes.io/name: {{ .Values.componentName }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "microservice.databaseSelectorLabels" -}}
app.kubernetes.io/name: {{ .Values.componentName }}-db
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
