{{ define "type" }}

# {{.Name.Name}}{{ if eq .Kind "Alias" }}(<code>{{.Underlying}}</code> alias){{ end }}
{{ with (typeReferences .) }}
*Appears on:*
{{- $prev := "" -}}
{{- range . -}}
  {{- if $prev -}}, {{ end -}}
  {{- $prev = . -}}
  [{{ typeDisplayName . }}]({{ linkForType . }})
{{- end -}}
{{ end }}

{{ safe (renderComments .CommentLines) }}

{{ with (constantsOfType .) }}
## Values
|Value|Description|
|-|-|
{{- range . -}}
|{{ typeDisplayName . }}|{{ safe (renderComments .CommentLines) }}|
{{- end -}}
{{ end }}

{{ if .Members }}
## Fields
|Field|Description|
|-|-|
{{ if isExportedType . }}
|apiVersion|string<br/><code>{{apiGroup .}}</code>|
|kind|string<br/><code>{{.Name.Name}}</code>|
{{ end }}
{{ template "members" .}}
{{ end }}

{{ end }}
