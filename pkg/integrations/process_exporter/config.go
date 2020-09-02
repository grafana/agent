// Package process_exporter embeds https://github.com/ncabatoff/process-exporter
package process_exporter //nolint:golint

import (
	"github.com/grafana/agent/pkg/integrations/config"

	exporter_config "github.com/ncabatoff/process-exporter/config"
)

var (
	DefaultConfig Config = Config{
		ProcFSPath: "/proc",
		Children:   true,
		Threads:    true,
		SMaps:      true,
		Recheck:    false,
	}
)

// Config controls the process_exporter integration.
type Config struct {
	// Enabled enables the process_exporter integration.
	Enabled bool `yaml:"enabled"`

	CommonConfig    config.Common                `yaml:",inline"`
	ProcessExporter exporter_config.MatcherRules `yaml:"process_names"`

	ProcFSPath string `yaml:"procfs_path"`
	Children   bool   `yaml:"track_children"`
	Threads    bool   `yaml:"track_threads"`
	SMaps      bool   `yaml:"gather_smaps"`
	Recheck    bool   `yaml:"recheck_on_scrape"`
}

func (c *Config) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}
