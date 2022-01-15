// Package process_exporter embeds https://github.com/ncabatoff/process-exporter
package process_exporter //nolint:golint

import (
	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/shared"

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
	ProcessExporter exporter_config.MatcherRules `yaml:"process_names,omitempty"`

	ProcFSPath string `yaml:"procfs_path,omitempty"`
	Children   bool   `yaml:"track_children,omitempty"`
	Threads    bool   `yaml:"track_threads,omitempty"`
	SMaps      bool   `yaml:"gather_smaps,omitempty"`
	Recheck    bool   `yaml:"recheck_on_scrape,omitempty"`
}

// Name returns the name of the integration that this shared represents.
func (c *Config) Name() string {
	return "process_exporter"
}

// InstanceKey returns the hostname of the machine.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return agentKey, nil
}

// NewIntegration converts this shared into an instance of an integration.
func (c *Config) NewIntegration(l log.Logger) (shared.Integration, error) {
	return New(l, c)
}
