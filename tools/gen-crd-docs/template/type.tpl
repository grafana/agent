{{ define "type" }}

### {{.Name.Name}}{{ if eq .Kind "Alias" }}(`{{.Underlying}}` alias){{ end }}

{{ with (typeReferences .) }}
(*Appears on: * {{- $prev := "" -}}{{ range . -}}{{- if $prev -}}, {{ end -}}{{- $prev = . -}}[{{ typeDisplayName . }}]({{ linkForType . }}){{ end }}{{ end }}{{ .CommentLines }})

{{ with (constantsOfType .) }}

#### Values

|Value|Description|
|-|-|
{{- range . -}}
|{{ typeDisplayName . }}|{{ safe (renderComments .CommentLines) }}|
{{- end -}}
{{ end }}

{{ if .Members }}

#### Fields

|Field|Description|
|-|-|
{{ if isExportedType . }}
|apiVersion|string<br/>`{{apiGroup .}}`|
|kind|string<br/>`{{.Name.Name}}`|
{{ end }}
{{ template "members" .}}
{{ end }}

{{ end }}
