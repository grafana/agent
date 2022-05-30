//go:build !linux || !amd64 || noebpf
// +build !linux !amd64 noebpf

package ebpf

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/integrations/v2"
)

// NewIntegration builds a no-op ebpf-integration for non-Linux systems.
func (c *Config) NewIntegration(l log.Logger, globals integrations.Globals) (integrations.Integration, error) {
	level.Warn(l).Log("msg", "the ebpf integration only works on linux; enabling it on other platforms will do nothing")
	return integrations.NoOpIntegration, nil
}
