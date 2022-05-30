//go:build linux && amd64 && !noebpf
// +build linux,amd64,!noebpf

package ebpf

import (
	"fmt"

	ebpf_config "github.com/cloudflare/ebpf_exporter/config"
	"github.com/cloudflare/ebpf_exporter/exporter"
	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations"
)

// New sets up the ebpf exporter.
func New(logger log.Logger, c *Config) (integrations.Integration, error) {

	exp, err := exporter.New(ebpf_config.Config{Programs: c.Programs})
	if err != nil {
		return nil, fmt.Errorf("failed to create ebpf exporter with input config: %s", err)
	}

	err = exp.Attach()
	if err != nil {
		return nil, fmt.Errorf("failed to attach ebpf exporter: %s", err)
	}

	return integrations.NewCollectorIntegration(
		c.Name(),
		integrations.WithCollectors(exp),
	), nil
}

// NewIntegration creates a new ebpf_exporter instance.
func (c *Config) NewIntegration(logger log.Logger) (integrations.Integration, error) {
	return New(logger, c)
}
