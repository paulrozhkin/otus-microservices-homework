{{- define "kafka-stack.fullname" -}}
{{- printf "%s-%s" .Release.Name .Values.componentName | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "kafka-stack.uiName" -}}
{{- printf "%s-ui" (include "kafka-stack.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "kafka-stack.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
app.kubernetes.io/name: {{ .Values.componentName }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/part-of: online-store
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "kafka-stack.selectorLabels" -}}
app.kubernetes.io/name: {{ .Values.componentName }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "kafka-stack.uiSelectorLabels" -}}
app.kubernetes.io/name: {{ .Values.componentName }}-ui
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
