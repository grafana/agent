//go:build !linux
// +build !linux

package ebpf

import (
	ebpf_config "github.com/cloudflare/ebpf_exporter/config"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/integrations/v2"
)

func init() {
	integrations.Register(&Config{}, integrations.TypeSingleton)
}

type Config struct {
	Programs []ebpf_config.Program `yaml:"programs,omitempty"`
}

func (c *Config) ApplyDefaults(globals integrations.Globals) error        { return nil }
func (c *Config) Identifier(globals integrations.Globals) (string, error) { return c.Name(), nil }
func (c *Config) Name() string                                            { return "ebpf" }
func (c *Config) NewIntegration(l log.Logger, globals integrations.Globals) (integrations.Integration, error) {
	level.Warn(l).Log("msg", "the ebpf integration only works on linux; enabling it on other platforms will do nothing")
	return integrations.NoOpIntegration, nil
}
