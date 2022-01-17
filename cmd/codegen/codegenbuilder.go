package main

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/grafana/agent/pkg/integrations/shared"
)

type codeGen struct {
}

func (c *codeGen) createV1Config(configs []configMeta) string {
	integrationTemplate, err := template.New("shared").Parse(`

type {{.Name}} struct {
  {{.ConfigStruct}} ` + "`yaml:\",omitempty,inline\"`" + `
  shared.Common ` + "`yaml:\",omitempty,inline\"`" + `
}

func (c *{{ .Name }}) Cfg() shared.Config {
	return &c.Config
}

func (c *{{ .Name }}) Cmn() shared.Common {
	return c.Common
}

{{ if .DefaultConfig }}
func (c *{{ .Name }}) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = {{ .PackageName }}.DefaultConfig
	type plain {{ .Name }}
	return unmarshal((*plain)(c))
}
{{ end }}
`)
	v1Template, err := template.New("v1").Parse(`
type V1Integration struct {
  {{ range $index, $element := . -}}
	{{ $element.Name }} *{{$element.Name}} ` + "`yaml:\"{{ $element.PackageName }},omitempty\"`\n" +
		`{{ end -}}
   TestConfigs []shared.V1IntegrationConfig ` + "`yaml:\"-,omitempty\"`\n" + `
}

func (v *V1Integration) ActiveConfigs() []shared.V1IntegrationConfig {
    activeConfigs := make([]shared.V1IntegrationConfig,0)
	{{ range $index, $element := . -}}
	if v.{{ $element.Name }} != nil {
        activeConfigs = append(activeConfigs, v.{{ $element.Name}})
    }
	{{ end -}}
    for _, i := range v.TestConfigs {
        activeConfigs = append(activeConfigs, i)
    }
    return activeConfigs
}
`)
	if err != nil {
		panic(err)
	}
	v1ConfigBuilder := strings.Builder{}
	v1ConfigBuilder.WriteString("package v1\n")
	v1ConfigBuilder.WriteString(`
import (
	"github.com/grafana/agent/pkg/integrations/shared"
	"github.com/grafana/agent/pkg/integrations/v1/agent"
	"github.com/grafana/agent/pkg/integrations/v1/cadvisor"
	"github.com/grafana/agent/pkg/integrations/v1/consul_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/dnsmasq_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/elasticsearch_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/github_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/kafka_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/memcached_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/mongodb_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/mysqld_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/node_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/postgres_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/process_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/redis_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/statsd_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/windows_exporter"
)
`)
	v1Buffer := bytes.Buffer{}
	err = v1Template.Execute(&v1Buffer, configs)
	if err != nil {
		panic(err)
	}

	v1ConfigBuilder.WriteString(v1Buffer.String())

	for _, cfg := range configs {
		bf := bytes.Buffer{}
		err = integrationTemplate.Execute(&bf, cfg)
		if err != nil {
			panic(err)
		}
		v1ConfigBuilder.WriteString(bf.String())
	}
	return v1ConfigBuilder.String()
}

func (c *codeGen) createV2Config(configs []configMeta) string {
	integrationTemplate, err := template.New("shared").Parse(`

type {{.Name}} struct {
  {{.ConfigStruct}} ` + "`yaml:\",omitempty,inline\"`" + `
  Cmn common.MetricsConfig  ` + "`yaml:\",inline\"`" + `
}

func (c *{{ .Name }}) Cfg() Config {
	return c
}

func (c *{{ .Name }}) Common() common.MetricsConfig {
	return c.Cmn
}

{{ if .DefaultConfig }}
func (c *{{ .Name }}) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = {{ .PackageName }}.DefaultConfig
	type plain {{ .Name }}
	return unmarshal((*plain)(c))
}
{{ end }}

{{ if eq .IsNativeV2 false -}}

func (c*{{ .Name }}) ApplyDefaults(globals Globals) error {
	c.Cmn.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Cmn.InstanceKey = &id
	}
	return nil
}

func (c *{{ .Name }}) Identifier(globals Globals) (string, error) {
	if c.Cmn.InstanceKey != nil {
		return *c.Cmn.InstanceKey, nil
	}
	return c.Config.InstanceKey(globals.AgentIdentifier)
}

func (c *{{ .Name }}) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return newIntegration(c, logger, globals, c.Config.NewIntegration)
}

{{ end -}}

{{ if eq .IsNativeV2 true -}}

func (c*{{ .Name }}) ApplyDefaults(globals Globals) error {
	return c.Config.ApplyDefaults(globals)
}

func (c *{{ .Name }}) Identifier(globals Globals) (string, error) {
	return c.Config.Identifier(globals)
}

func (c *{{ .Name }}) NewIntegration(logger log.Logger, globals Globals) (Integration, error) {
	return c.Config.NewIntegration(logger,globals)
}

{{ end -}}
`)
	// Type: 1 = Singleton, 2 = Multiplex
	v2template, err := template.New("v1").Parse(`
type Integrations struct {
  {{ range $index, $element := . -}}
	{{ if eq .Type 0 -}}
		{{ $element.Name }} *{{$element.Name}} ` + "`yaml:\"{{ $element.PackageName }},omitempty\"`" + `
    {{ end -}}
    {{ if eq .Type  1 -}}
       {{ $element.Name }}Configs []*{{$element.Name}} ` + "`yaml:\"{{ $element.PackageName }}_configs,omitempty\"`" + `
    {{ end -}}
    {{ end -}}
   TestConfigs []Config  ` + "`yaml:\"-,omitempty\"`\n" + `

}

func (v *Integrations) ActiveConfigs() []Config {
    activeConfigs := make([]Config,0)
	{{ range $index, $element := . -}}
    {{ if eq .Type  0 -}}
	if v.{{ $element.Name }} != nil {
        activeConfigs = append(activeConfigs, v.{{ $element.Name}})
    }
    {{ end -}}
	{{ if eq .Type  1 -}}
	for _, i := range v.{{ $element.Name}}Configs {
		activeConfigs = append(activeConfigs, i)
	}
    {{ end -}}
	{{ end -}}
    for _, i := range v.TestConfigs {
        activeConfigs = append(activeConfigs, i)
    }
    return activeConfigs
}
`)
	if err != nil {
		panic(err)
	}
	v2ConfigBuilder := strings.Builder{}
	v2ConfigBuilder.WriteString("package v2\n")
	v2ConfigBuilder.WriteString(`
import (
"context"
	"errors"
	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/shared"
	"github.com/grafana/agent/pkg/integrations/v1/agent"
    "github.com/grafana/agent/pkg/integrations/v1/cadvisor"
	"github.com/grafana/agent/pkg/integrations/v1/consul_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/dnsmasq_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/elasticsearch_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/github_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/kafka_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/memcached_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/mongodb_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/mysqld_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/node_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/postgres_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/process_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/redis_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/statsd_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/windows_exporter"
	"github.com/grafana/agent/pkg/integrations/v2/common"
)
`)
	v1Buffer := bytes.Buffer{}
	err = v2template.Execute(&v1Buffer, configs)
	if err != nil {
		panic(err)
	}

	v2ConfigBuilder.WriteString(v1Buffer.String())

	for _, cfg := range configs {
		bf := bytes.Buffer{}
		err = integrationTemplate.Execute(&bf, cfg)
		if err != nil {
			panic(err)
		}
		v2ConfigBuilder.WriteString(bf.String())
	}
	v2ConfigBuilder.WriteString(`
func newIntegration(c IntegrationConfig, logger log.Logger, globals Globals, newInt func(l log.Logger) (shared.Integration, error)) (Integration, error) {

	v1Integration, err := newInt(logger)
	if err != nil {
		return nil, err
	}

	id, err := c.Cfg().Identifier(globals)
	if err != nil {
		return nil, err
	}

	// Generate our handler. Original integrations didn't accept a prefix, and
	// just assumed that they would be wired to /metrics somewhere.
	handler, err := v1Integration.MetricsHandler()
	if err != nil {
		return nil, fmt.Errorf("generating http handler: %w", err)
	} else if handler == nil {
		handler = http.NotFoundHandler()
	}

	// Generate targets. Original integrations used a static set of targets,
	// so this mapping can always be generated just once.
	//
	// Targets are generated from the result of ScrapeConfigs(), which returns a
	// tuple of job name and relative metrics path.
	//
	// Job names were prefixed at the subsystem level with integrations/, so we
	// will retain that behavior here.
	v1ScrapeConfigs := v1Integration.ScrapeConfigs()
	targets := make([]handlerTarget, 0, len(v1ScrapeConfigs))
	for _, sc := range v1ScrapeConfigs {
		targets = append(targets, handlerTarget{
			MetricsPath: sc.MetricsPath,
			Labels: model.LabelSet{
				model.JobLabel: model.LabelValue("integrations/" + sc.JobName),
			},
		})
	}

	// Convert he run function. Original integrations sometimes returned
	// ctx.Err() on exit. This isn't recommended anymore, but we need to hide the
	// error if it happens, since the error was previously ignored.
	runFunc := func(ctx context.Context) error {
		err := v1Integration.Run(ctx)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, context.Canceled) && ctx.Err() != nil:
			// Hide error that no longer happens in newer integrations.
			return nil
		default:
			return err
		}
	}

	// Aggregate our converted settings into a v2 integration.
	return &metricsHandlerIntegration{
		integrationName: c.Cfg().Name(),
		instanceID:      id,

		common:  c.Common(),
		globals: globals,
		handler: handler,
		targets: targets,

		runFunc: runFunc,
	}, nil
}
`)
	return v2ConfigBuilder.String()
}

type configMeta struct {
	Name          string
	ConfigStruct  string
	DefaultConfig string
	PackageName   string
	Type          shared.Type
	IsNativeV2    bool
}
