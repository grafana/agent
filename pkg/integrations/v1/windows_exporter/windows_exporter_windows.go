package windows_exporter //nolint:golint

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/grafana/agent/pkg/integrations/shared"

	"github.com/go-kit/log"
	"github.com/prometheus-community/windows_exporter/collector"
	"github.com/prometheus/statsd_exporter/pkg/level"
)

// New creates a new windows_exporter integration.
func New(log log.Logger, c *Config) (shared.Integration, error) {
	// Get a list of collector configs and map our local config to it.
	availableConfigs := collector.AllConfigs()
	c.toExporterConfig(availableConfigs)

	enabledCollectorNames := enabledCollectors(c.EnabledCollectors)
	collectors, err := buildCollectors(enabledCollectorNames, availableConfigs)
	if err != nil {
		return nil, err
	}

	collectorNames := make([]string, 0, len(collectors))
	for key := range collectors {
		collectorNames = append(collectorNames, key)
	}
	sort.Strings(collectorNames)
	level.Info(log).Log("msg", "enabled windows_exporter collectors", "collectors", collectorNames)

	return shared.NewCollectorIntegration(c.Name(), shared.WithCollectors(
		// Hard-coded 4m timeout to represent the time a series goes stale.
		// TODO: Make configurable if useful.
		collector.NewPrometheus(4*time.Minute, collectors),
	)), nil
}

func enabledCollectors(input string) []string {
	separated := strings.Split(input, ",")
	unique := map[string]struct{}{}
	for _, s := range separated {
		s = strings.TrimSpace(s)
		if s != "" {
			unique[s] = struct{}{}
		}
	}
	result := make([]string, 0, len(unique))
	for s := range unique {
		result = append(result, s)
	}
	return result
}

func buildCollectors(enabled []string, available []collector.Config) (map[string]collector.Collector, error) {
	collectors := map[string]collector.Collector{}

	for _, name := range enabled {
		var found collector.Config
		for _, c := range available {
			if c.Name() == name {
				found = c
				break
			}
		}
		if found == nil {
			return nil, fmt.Errorf("unknown collector %q", name)
		}

		c, err := found.Build()
		if err != nil {
			return nil, err
		}
		collectors[name] = c
	}

	return collectors, nil
}
