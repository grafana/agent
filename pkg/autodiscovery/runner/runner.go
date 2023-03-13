package runner

import (
	"bytes"
	"io"
	"text/template"

	"github.com/grafana/agent/pkg/autodiscovery"
	"github.com/grafana/agent/pkg/autodiscovery/apache"
	"github.com/grafana/agent/pkg/autodiscovery/consul"
	"github.com/grafana/agent/pkg/autodiscovery/kubernetes"
	"github.com/grafana/agent/pkg/autodiscovery/mysql"
	"github.com/grafana/agent/pkg/autodiscovery/postgres"
	"github.com/grafana/agent/pkg/autodiscovery/prometheus"
	"github.com/grafana/agent/pkg/autodiscovery/redis"
	"github.com/grafana/agent/pkg/river/parser"
	"github.com/grafana/agent/pkg/river/printer"
)

type templateInput struct {
	Components     []string
	MetricsExports []string
	MetricsTargets []string
	LogsTargets    []string
}

type Discoverer interface {
	Run() (*autodiscovery.Result, error)
}

func initComponent[T any](v T, err error) T {
	if err != nil {
		//TODO: Log a warning? Or panic?
		panic(err)
	}
	return v
}

// Do do doo do doo.
func Do(wr io.Writer) {
	discFuncs := []Discoverer{
		initComponent(mysql.New()),
		initComponent(postgres.New()),
		initComponent(consul.New()),
		initComponent(prometheus.New()),
		initComponent(kubernetes.New()),
		initComponent(redis.New()),
		initComponent(apache.New()),
	}

	var results []*autodiscovery.Result

	for _, f := range discFuncs {
		res, err := f.Run()
		if err != nil {
			panic(err)
		} else {
			results = append(results, res)
		}
	}

	input := BuildTemplateInput(results)

	//TODO: Check RenderConfig for errors?
	// We don't have to return it. Maybe log a warning and continue silently?
	RenderConfig(wr, input)

	//TODO: If the agent already has a River config, can we merge this new one and the existing one?
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

func RenderConfig(wr io.Writer, input templateInput) error {
	tmpl := template.New("cfg.river")
	tmpl = template.Must(tmpl.Parse(templateStr))

	rawBuf := new(bytes.Buffer)
	tmpl.Execute(rawBuf, input)

	return PretifyRiver(wr, rawBuf.Bytes())
}

// TODO: The main formatting logic was copied form riverfmt. Place this
// function in a shared package, and have riverfmt use it from the same shared place.
func PretifyRiver(wr io.Writer, riverCfg []byte) error {
	ast, err := parser.ParseFile("", riverCfg)
	if err != nil {
		return err
	}

	var prettyBuf bytes.Buffer
	if err := printer.Fprint(&prettyBuf, ast); err != nil {
		return err
	}

	wr.Write(prettyBuf.Bytes())
	return nil
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
