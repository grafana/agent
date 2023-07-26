{{ define "type" }}

### {{.Name.Name}}{{ if eq .Kind "Alias" }}(`{{.Underlying}}` alias){{ end }} <a name="{{ anchorIDForType . }}"></a>

{{ with (typeReferences .) }}
(Appears on: {{- $prev := "" -}}{{ range . -}}{{- if $prev -}}, {{ end -}}{{- $prev = . -}} [{{ typeDisplayName . }}]({{ linkForType . }}){{ end }}){{ end }}

{{ range .CommentLines }}{{ . }} {{ end }}

{{ with (constantsOfType .) }}

#### Values

|Value|Description|
|-|-|
{{- range . -}}
|{{ typeDisplayName . }}|{{ range .CommentLines }}{{ . }} {{ end }}|
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
