package blackbox

import (
	"time"

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
		Build:   exporter.NewMultiTarget(createExporter, "blackbox", buildBlackboxTargets),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	cfg := args.(Config)
	return cfg.Convert().NewIntegration(opts.Logger)
}

// buildBlackboxTargets creates the exporter's discovery targets based on the defined blackbox targets.
func buildBlackboxTargets(baseTarget discovery.Target, args component.Arguments) []discovery.Target {
	var targets []discovery.Target

	cfg := args.(Config)
	for _, tgt := range cfg.Targets {
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
	ProbeTimeoutOffset: 500 * time.Millisecond,
}

// BlackboxTarget defines a target to be used by the exporter.
type BlackboxTarget struct {
	Name   string `river:",label"`
	Target string `river:"address,attr"`
	Module string `river:"module,attr,optional"`
}

type TargetBlock []BlackboxTarget

// Convert converts the component's TargetBlock to a slice of integration's BlackboxTarget.
func (t TargetBlock) Convert() []blackbox_exporter.BlackboxTarget {
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
	ConfigFile         string        `river:"config_file,attr"`
	Targets            TargetBlock   `river:"target,block"`
	ProbeTimeoutOffset time.Duration `river:"probe_timeout_offset,attr,optional"`
}

// UnmarshalRiver implements River unmarshalling for Config.
func (c *Config) UnmarshalRiver(f func(interface{}) error) error {
	*c = DefaultConfig

	type cfg Config
	return f((*cfg)(c))
}

// Convert converts the component's Config to the integration's Config.
func (c *Config) Convert() *blackbox_exporter.Config {
	return &blackbox_exporter.Config{
		BlackboxConfigFile: c.ConfigFile,
		BlackboxTargets:    c.Targets.Convert(),
		ProbeTimeoutOffset: c.ProbeTimeoutOffset.Seconds(),
	}
}
