package blackbox

import (
	"errors"
	"fmt"
	"time"

	blackbox_config "github.com/prometheus/blackbox_exporter/config"
	"gopkg.in/yaml.v2"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/blackbox_exporter"
	"github.com/grafana/river/rivertypes"
)

func init() {
	component.Register(component.Registration{
		Name:          "prometheus.exporter.blackbox",
		Args:          Arguments{},
		Exports:       exporter.Exports{},
		NeedsServices: exporter.RequiredServices(),
		Build:         exporter.NewWithTargetBuilder(createExporter, "blackbox", buildBlackboxTargets),
	})
}

func createExporter(opts component.Options, args component.Arguments, defaultInstanceKey string) (integrations.Integration, string, error) {
	a := args.(Arguments)
	return integrations.NewIntegrationWithInstanceKey(opts.Logger, a.Convert(), defaultInstanceKey)
}

// buildBlackboxTargets creates the exporter's discovery targets based on the defined blackbox targets.
func buildBlackboxTargets(baseTarget discovery.Target, args component.Arguments) []discovery.Target {
	var targets []discovery.Target

	a := args.(Arguments)
	for _, tgt := range a.Targets {
		target := make(discovery.Target)
		// Set extra labels first, meaning that any other labels will override
		for k, v := range tgt.Labels {
			target[k] = v
		}
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

// DefaultArguments holds non-zero default options for Arguments when it is
// unmarshaled from river.
var DefaultArguments = Arguments{
	ProbeTimeoutOffset: 500 * time.Millisecond,
}

// BlackboxTarget defines a target to be used by the exporter.
type BlackboxTarget struct {
	Name   string            `river:",label"`
	Target string            `river:"address,attr"`
	Module string            `river:"module,attr,optional"`
	Labels map[string]string `river:"labels,attr,optional"`
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

type Arguments struct {
	ConfigFile         string                    `river:"config_file,attr,optional"`
	Config             rivertypes.OptionalSecret `river:"config,attr,optional"`
	Targets            TargetBlock               `river:"target,block"`
	ProbeTimeoutOffset time.Duration             `river:"probe_timeout_offset,attr,optional"`
	ConfigStruct       blackbox_config.Config
}

// SetToDefault implements river.Defaulter.
func (a *Arguments) SetToDefault() {
	*a = DefaultArguments
}

// Validate implements river.Validator.
func (a *Arguments) Validate() error {
	if a.ConfigFile != "" && a.Config.Value != "" {
		return errors.New("config and config_file are mutually exclusive")
	}

	err := yaml.UnmarshalStrict([]byte(a.Config.Value), &a.ConfigStruct)
	if err != nil {
		return fmt.Errorf("invalid backbox_exporter config: %s", err)
	}

	return nil
}

// Convert converts the component's Arguments to the integration's Config.
func (a *Arguments) Convert() *blackbox_exporter.Config {
	return &blackbox_exporter.Config{
		BlackboxConfigFile: a.ConfigFile,
		BlackboxConfig:     a.ConfigStruct,
		BlackboxTargets:    a.Targets.Convert(),
		ProbeTimeoutOffset: a.ProbeTimeoutOffset.Seconds(),
	}
}
