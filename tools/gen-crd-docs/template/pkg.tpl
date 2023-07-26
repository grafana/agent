{{ define "packages" }}

{{ with .packages}}
---
aliases:
- /docs/agent/latest/operator/crd/
canonical: https://grafana.com/docs/agent/latest/operator/api/
title: Custom Resource Definition Reference
weight: 500
---
# Custom Resource Definition Reference
{{ end}} 

{{ range .packages }}
{{ with (index .GoPackages 0 )}}
{{ with .DocComments }}
{{ . }}
{{ end }} 
{{ end }} 

## Resource Types:
{{ range (visibleTypes (sortedTypes .Types)) }} 
{{ if isExportedType . -}}
* [{{ typeDisplayName . }}]({{ linkForType . }}) 
{{- end }} 
{{ end }}

{{ range (visibleTypes (sortedTypes .Types))}} 
{{ template "type" . }} 
{{ end }}
{{ end }}
{{ end }}
