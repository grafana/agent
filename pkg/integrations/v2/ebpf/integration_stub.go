//go:build !linux || !ebpf_enabled
// +build !linux !ebpf_enabled

package ebpf

import (
	"github.com/grafana/agent/pkg/integrations/v2"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func init() {
	integrations.Register(&config{}, integrations.TypeSingleton)
}

type config struct{}

func (c *config) ApplyDefaults(globals integrations.Globals) error        { return nil }
func (c *config) Identifier(globals integrations.Globals) (string, error) { return c.Name(), nil }
func (c *config) Name() string                                            { return "ebpf" }
func (c *config) NewIntegration(l log.Logger, globals integrations.Globals) (integrations.Integration, error) {
	level.Warn(l).Log("msg", "the ebpf integration only works on linux; enabling it on other platforms will do nothing")
	return integrations.NoOpIntegration, nil
}
