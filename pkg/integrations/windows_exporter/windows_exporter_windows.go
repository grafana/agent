package windows_exporter //nolint:golint

import (
	"github.com/grafana/agent/pkg/integrations"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus-community/windows_exporter/exporter"
)

// New creates a new node_exporter integration.
func New(log log.Logger, c *Config) (integrations.Integration, error) {
	cm := c.ConvertToMap()
	wc, err := exporter.NewWindowsCollector(c.Name(), c.EnabledCollectors, cm)
	if err != nil {
		return nil, err
	}
	level.Info(log).Log("msg", "Enabled windows_exporter collectors")
	return integrations.NewCollectorIntegration(c.Name(), integrations.WithCollectors(wc)), nil
}
