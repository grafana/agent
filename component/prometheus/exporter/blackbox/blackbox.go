package blackbox

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/blackbox_exporter"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.blackbox",
		Args:    Config{},
		Exports: exporter.Exports{},
		Build:   exporter.New(createExporter, "blackbox", buildBlackboxTargets),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	cfg := args.(Config)
	return cfg.Convert().NewIntegration(opts.Logger)
}

func buildBlackboxTargets(baseTarget discovery.Target, args component.Arguments) []discovery.Target {
	var targets []discovery.Target

	cfg := args.(Config)
	for _, tgt := range cfg.BlackboxTargets {
		target := make(discovery.Target)
		for k, v := range baseTarget {
			target[k] = v
		}

		target["job"] = target["job"] + "/" + tgt.Name
		target["__param_target"] = tgt.Target
		if tgt.Module != "" {
			target["__param_module"] = tgt.Module
		}

		targets = append(targets, target)
	}

	return targets
}

// DefaultConfig holds non-zero default options for the Config when it is
// unmarshaled from river.
var DefaultConfig = Config{
	ProbeTimeoutOffset: 0.5,
}

// BlackboxTarget defines a target device to be used by the integration.
type BlackboxTarget struct {
	Name   string `river:"name,attr"`
	Target string `river:"address,attr"`
	Module string `river:"module,attr"`
}

type BlackboxTargets []BlackboxTarget

func (t BlackboxTargets) Convert() []blackbox_exporter.BlackboxTarget {
	targets := make([]blackbox_exporter.BlackboxTarget, 0, len(t))
	for _, target := range t {
		targets = append(targets, blackbox_exporter.BlackboxTarget{
			Name:   target.Name,
			Target: target.Target,
			Module: target.Module,
		})
	}
	return targets
}

type Config struct {
	BlackboxConfigFile string          `river:"config_file,attr"`
	BlackboxTargets    BlackboxTargets `river:"blackbox_target,block"`
	ProbeTimeoutOffset float64         `river:"probe_timeout_offset,attr,optional"`
}

// UnmarshalRiver implements River unmarshalling for Config.
func (c *Config) UnmarshalRiver(f func(interface{}) error) error {
	*c = DefaultConfig

	type cfg Config
	return f((*cfg)(c))
}

func (c *Config) Convert() *blackbox_exporter.Config {
	return &blackbox_exporter.Config{
		BlackboxConfigFile: c.BlackboxConfigFile,
		BlackboxTargets:    c.BlackboxTargets.Convert(),
		ProbeTimeoutOffset: c.ProbeTimeoutOffset,
	}
}
