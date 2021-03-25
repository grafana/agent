package windows_exporter //nolint:golint

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/prometheus-community/windows_exporter/exporter"
)

// New creates a new windows_exporter integration.
func New(log log.Logger, c *Config) (integrations.Integration, error) {
	configMap := exporter.GenerateConfigs()
	c.applyConfig(configMap)
	wc, err := exporter.NewWindowsCollector(c.Name(), c.EnabledCollectors, configMap)
	if err != nil {
		return nil, err
	}
	_ = level.Info(log).Log("msg", "Enabled windows_exporter collectors")
	return integrations.NewCollectorIntegration(c.Name(), integrations.WithCollectors(wc)), nil
}
