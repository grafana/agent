package loki

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/grafana/loki/pkg/promtail/client"
	"github.com/grafana/loki/pkg/promtail/positions"
	"github.com/grafana/loki/pkg/promtail/scrapeconfig"
	"github.com/grafana/loki/pkg/promtail/targets/file"
)

// ConfigV0 controls the configuration of the Loki log scraper.
type ConfigV0 struct {
	Version string `yaml:"version"`

	ClientConfigs   []client.Config       `yaml:"clients,omitempty"`
	PositionsConfig positions.Config      `yaml:"positions,omitempty"`
	ScrapeConfig    []scrapeconfig.Config `yaml:"scrape_configs,omitempty"`
	TargetConfig    file.Config           `yaml:"target_config,omitempty"`
}

// Upgrade upgrades a ConfigV0 to ConfigV1.
func (c *ConfigV0) Upgrade() (*ConfigV1, error) {
	v1 := ConfigV1{
		Version: "v1",
		Configs: []*InstanceConfig{{
			Name:            "default",
			ClientConfigs:   c.ClientConfigs,
			PositionsConfig: c.PositionsConfig,
			ScrapeConfig:    c.ScrapeConfig,
			TargetConfig:    c.TargetConfig,
		}},
	}
	return &v1, v1.ApplyDefaults()
}

func (c *ConfigV0) RegisterFlags(f *flag.FlagSet) {
	c.PositionsConfig.RegisterFlagsWithPrefix("loki.", f)
	c.TargetConfig.RegisterFlagsWithPrefix("loki.", f)
}

// ConfigV1 controls the configuration of the Loki log scraper.
type ConfigV1 struct {
	Version string `yaml:"version"`

	PositionsDirectory string            `yaml:"positions_directory"`
	Configs            []*InstanceConfig `yaml:"configs"`
}

func (c *ConfigV1) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type configV1 ConfigV1
	err := unmarshal((*configV1)(c))
	if err != nil {
		return err
	}

	return c.ApplyDefaults()
}

func (c *ConfigV1) RegisterFlags(f *flag.FlagSet) {
	for _, ic := range c.Configs {
		ic.RegisterFlags(f)
	}
}

// ApplyDefaults applies defaults to the ConfigV1 and ensures that it is valid.
//
// Validations:
//
//   1. No two InstanceConfigs may have the same name.
//   2. No two InstanceConfigs may have the same positions path.
//   3. No InstanceConfig may have an empty name.
//   4. If InstanceConfig positions path is empty, shared PositionsDirectory
//      must not be empty.
//
// Defaults:
//
//   1. If a positions config is empty, it will be generated based on
//      the InstanceConfig name and ConfigV1.PositionsDirectory.
func (c *ConfigV1) ApplyDefaults() error {
	var (
		names     = map[string]struct{}{}
		positions = map[string]string{} // positions file name -> config using it
	)

	for idx, ic := range c.Configs {
		if ic.Name == "" {
			return fmt.Errorf("Loki config index %d must have a name", idx)
		}
		if _, ok := names[ic.Name]; ok {
			return fmt.Errorf("found two Loki configs with name %s", ic.Name)
		}
		names[ic.Name] = struct{}{}

		if ic.PositionsConfig.PositionsFile == "" {
			if c.PositionsDirectory == "" {
				return fmt.Errorf("cannot generate Loki positions file path for %s because positions_directory is not configured", ic.Name)
			}
			ic.PositionsConfig.PositionsFile = filepath.Join(c.PositionsDirectory, ic.Name+".yml")
		}
		if orig, ok := positions[ic.PositionsConfig.PositionsFile]; ok {
			return fmt.Errorf("Loki configs %s and %s must have different positions file paths", orig, ic.Name)
		}
		positions[ic.PositionsConfig.PositionsFile] = ic.Name
	}

	return nil
}

// InstanceConfig is an individual Promtail config.
type InstanceConfig struct {
	Name string `yaml:"name,omitempty"`

	ClientConfigs   []client.Config       `yaml:"clients,omitempty"`
	PositionsConfig positions.Config      `yaml:"positions,omitempty"`
	ScrapeConfig    []scrapeconfig.Config `yaml:"scrape_configs,omitempty"`
	TargetConfig    file.Config           `yaml:"target_config,omitempty"`
}

func (c *InstanceConfig) RegisterFlags(f *flag.FlagSet) {
	c.PositionsConfig.RegisterFlagsWithPrefix("loki."+c.Name+".", f)
	c.TargetConfig.RegisterFlagsWithPrefix("loki."+c.Name+".", f)
}
