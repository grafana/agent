// Package process_exporter embeds https://github.com/ncabatoff/process-exporter
package process_exporter //nolint:golint

import (
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	loki "github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/tempo"

	exporter_config "github.com/ncabatoff/process-exporter/config"
)

// DefaultConfig holds the default settings for the process_exporter integration.
var DefaultConfig = Config{
	ProcFSPath: "/proc",
	Children:   true,
	Threads:    true,
	SMaps:      true,
	Recheck:    false,
}

// Config controls the process_exporter integration.
type Config struct {
	Common          config.Common                `yaml:",inline"`
	ProcessExporter exporter_config.MatcherRules `yaml:"process_names,omitempty"`

	ProcFSPath string `yaml:"procfs_path,omitempty"`
	Children   bool   `yaml:"track_children,omitempty"`
	Threads    bool   `yaml:"track_threads,omitempty"`
	SMaps      bool   `yaml:"gather_smaps,omitempty"`
	Recheck    bool   `yaml:"recheck_on_scrape,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *Config) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string {
	return "process_exporter"
}

// CommonConfig returns the set of common settings shared across all integrations.
func (c *Config) CommonConfig() config.Common {
	return c.Common
}

// NewIntegration converts this config into an instance of an integration.
func (c *Config) NewIntegration(l log.Logger, loki *loki.Logs, tempo *tempo.Tempo) (integrations.Integration, error) {
	return New(l, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
}
