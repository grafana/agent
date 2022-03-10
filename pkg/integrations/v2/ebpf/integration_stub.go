//go:build !linux
// +build !linux

package ebpf

import (
	"github.com/grafana/agent/pkg/integrations/v2"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type config struct{}

func (c *config) Name() string                             { return "ebpf no-op" }
func (c *config) ApplyDefaults(integrations.Globals) error { return nil }
func (c *config) Identifier(integrations.Globals) (string, error) {
	return "stub ebpf integration", nil
}

// NewIntegration creates a new no-op ebpf integration for non-Linux platforms
func (c *config) NewIntegration(logger log.Logger, globals integrations.Globals) (integrations.Integration, error) {
	level.Warn(logger).Log("msg", "the ebpf integration only works on linux; enabling it on other platforms will do nothing")
	return integrations.NoOpIntegration, nil
}
