{{ define "packages" }}

{{ with .packages}}
---
aliases:
- /docs/agent/latest/operator/crd/
title: Custom Resource Definition Reference
weight: 500
---
# Custom Resource Definition Reference

## Packages:

{{ range . }}
*   [{{ packageDisplayName . }}](#{{- packageAnchorID . -}})
{{ end }}

{{ end}} 

{{ range .packages }}

## [{{ packageDisplayName . }}](#{{- packageAnchorID . -}})

{{ with (index .GoPackages 0 )}}
{{ with .DocComments }}

{{ safe (renderComments .) }}

{{ end }} 
{{ end }} 

## Resource Types:

{{- range (visibleTypes (sortedTypes .Types)) -}} 
{{ if isExportedType . -}}
*   [{{ typeDisplayName . }}]({{ linkForType . }})
{{- end }} 
{{- end -}}

{{ range (visibleTypes (sortedTypes .Types))}} 
{{ template "type" . }} 
{{ end }}
{{ end }}

_Generated with `gen-crd-api-reference-docs` {{ with .gitCommit }} on git commit `{{ . }}`{{end}}._

{{ end }}
