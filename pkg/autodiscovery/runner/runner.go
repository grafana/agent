package runner

import (
	"os"
	"text/template"

	"github.com/grafana/agent/pkg/autodiscovery"
	"github.com/grafana/agent/pkg/autodiscovery/mysql"
)

type templateInput struct {
	Components     []string
	MetricsExports []string
	MetricsTargets []string
	LogsTargets    []string
}

// Do do doo do doo.
func Do() {
	var results []*autodiscovery.Result
	mysql, err := mysql.New()
	if err != nil {
		panic(err)
	}

	res, err := mysql.Run()
	if err != nil {
		panic(err)
	} else {
		results = append(results, res)
	}

	input := BuildTemplateInput(results)

	RenderConfig(input)
}

// BuildTemplateInput ...
func BuildTemplateInput(input []*autodiscovery.Result) templateInput {
	res := templateInput{}

	for _, r := range input {
		res.Components = append(res.Components, r.RiverConfig)
		if r.MetricsExport != "" {
			res.MetricsExports = append(res.MetricsExports, r.MetricsExport)
		}
		if len(r.MetricsTargets) > 0 {
			for _, mt := range r.MetricsTargets {
				res.MetricsTargets = append(res.MetricsTargets, mt.RiverString())
			}
		}
		if len(r.LogfileTargets) > 0 {
			for _, lt := range r.LogfileTargets {
				res.LogsTargets = append(res.LogsTargets, lt.RiverString())
			}
		}
	}

	return res
}

func RenderConfig(input templateInput) {
	tmpl := template.New("cfg.river")
	tmpl = template.Must(tmpl.Parse(templateStr))

	tmpl.Execute(os.Stdout, input)
}

var templateStr = `
prometheus.scrape "default" {
  targets = concat({{if .MetricsExports }}{{range .MetricsExports}}
  {{if .}}  {{.}},{{end}}{{end}}{{end}}
  {{if .MetricsTargets }} [{{range .MetricsTargets}}
      {{.}},{{end}}
    ],{{end}}
  )
  forward_to = [prometheus.remote_write.default.receiver]
}

prometheus.remote_write "default" {
  endpoint {
    url = env("GRAFANACLOUD_METRICS_URL")

    basic_auth {
      username = env("GRAFANACLOUD_METRICS_USER")
      password = env("GRAFANACLOUD_APIKEY")
    }
  }
}

{{if .LogsTargets }}
loki.source.file "default" {
  targets = [{{range .LogsTargets}}
      {{.}},{{end}}
  ]

  forward_to = [loki.write.default.receiver]
}

loki.write "default" {
  endpoint {
    url = env("GRAFANACLOUD_LOGS_URL")
    basic_auth {
      username = env("GRAFANACLOUD_LOGS_USER")
      password = env("GRAFANACLOUD_APIKEY")
    }
  }
}
{{end}}

{{ range .Components}} {{if .}}
{{.}}
{{ end }}{{end}}
`
