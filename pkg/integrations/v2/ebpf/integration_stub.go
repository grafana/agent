//go:build !linux
// +build !linux

package ebpf

import (
	"github.com/grafana/agent/pkg/integrations/v2"

	"github.com/go-kit/log"
)

func init() {
	integrations.Register(&config{}, integrations.TypeSingleton)
}

type config struct{}

func (c *config) ApplyDefaults(globals integrations.Globals) error        { return nil }
func (c *config) Identifier(globals integrations.Globals) (string, error) { return c.Name(), nil }
func (c *config) Name() string                                            { return "noop-ebpf" }

func (c *config) NewIntegration(l log.Logger, globals integrations.Globals) (integrations.Integration, error) {
	return integrations.NoOpIntegration, nil
}
