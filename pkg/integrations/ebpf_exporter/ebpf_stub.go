//go:build !linux || !amd64 || noebpf
// +build !linux !amd64 noebpf

package ebpf

import (
	"github.com/grafana/agent/pkg/integrations"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// NewIntegration creates a new ebpf_exporter.
func (c *Config) NewIntegration(logger log.Logger) (integrations.Integration, error) {
	level.Warn(logger).Log("msg", "the ebpf integration only works on linux; enabling it on other platforms will do nothing")
	return &integrations.StubIntegration{}, nil
}
