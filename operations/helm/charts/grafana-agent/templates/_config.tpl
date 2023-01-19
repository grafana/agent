{{/*
Retrieve configMap name from the name of the chart or the ConfigMap the user
specified.
*/}}
{{- define "grafana-agent.config-map.name" -}}
{{- if .Values.agent.configMap.name -}}
{{- .Values.agent.configMap.name }}
{{- else -}}
{{- include "grafana-agent.fullname" . }}
{{- end }}
{{- end }}

{{/*
The name of the config file is the default or the key the user specified in the
ConfigMap.
*/}}
{{- define "grafana-agent.config-map.key" -}}
{{- if .Values.agent.configMap.key -}}
{{- .Values.agent.configMap.key }}
{{- else if eq .Values.agent.mode "flow" -}}
config.river
{{- else if eq .Values.agent.mode "static" -}}
config.yaml
{{- end }}
{{- end }}
