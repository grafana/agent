package process_exporter

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/integration"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/process_exporter"
	exporter_config "github.com/ncabatoff/process-exporter/config"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.integration.process_exporter",
		Args:    Config{},
		Exports: integration.Exports{},
		Build:   integration.New(createIntegration, "process_exporter"),
	})
}

func createIntegration(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	cfg := args.(Config)
	return cfg.Convert().NewIntegration(opts.Logger)
}

// DefaultConfig holds the default settings for the process_exporter integration.
var DefaultConfig = Config{
	ProcFSPath: "/proc",
	Children:   true,
	Threads:    true,
	SMaps:      true,
	Recheck:    false,
}

// Config is the base config for this component.
type Config struct {
	ProcessExporter MatcherRules `river:"process_names,block,optional"`

	ProcFSPath string `river:"procfs_path,attr,optional"`
	Children   bool   `river:"track_children,attr,optional"`
	Threads    bool   `river:"track_threads,attr,optional"`
	SMaps      bool   `river:"gather_smaps,attr,optional"`
	Recheck    bool   `river:"recheck_on_scrape,attr,optional"`
}

// MatcherGroup and MatcherRules taken and converted to River from github.com/ncabatoff/process-exporter/config
type MatcherGroup struct {
	Name         string   `river:"name,attr,optional"`
	CommRules    []string `river:"comm,attr,optional"`
	ExeRules     []string `river:"exe,attr,optional"`
	CmdlineRules []string `river:"cmdline,attr,optional"`
}
type MatcherRules []MatcherGroup

// UnmarshalRiver implements River unmarshalling for Config.
func (c *Config) UnmarshalRiver(f func(interface{}) error) error {
	*c = DefaultConfig

	type cfg Config
	return f((*cfg)(c))
}

func (c *Config) Convert() *process_exporter.Config {
	return &process_exporter.Config{
		ProcessExporter: c.ProcessExporter.Convert(),
		ProcFSPath:      c.ProcFSPath,
		Children:        c.Children,
		Threads:         c.Threads,
		SMaps:           c.SMaps,
		Recheck:         c.Recheck,
	}
}

func (m MatcherRules) Convert() exporter_config.MatcherRules {
	var out exporter_config.MatcherRules
	for _, v := range m {
		out = append(out, exporter_config.MatcherGroup(v))
	}
	return out
}
