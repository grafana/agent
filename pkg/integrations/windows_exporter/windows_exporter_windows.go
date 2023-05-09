package windows_exporter //nolint:golint

import (
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/prometheus-community/windows_exporter/collector"
)

// New creates a new windows_exporter integration.
func New(logger log.Logger, c *Config) (integrations.Integration, error) {
	windowsExporter := kingpin.New("", "")
	collector.RegisterCollectorsFlags(windowsExporter)
	c.toExporterConfig(windowsExporter)

	if _, err := windowsExporter.Parse([]string{}); err != nil {
		return nil, err
	}

	collector.RegisterCollectors()
	enabledCollectorNames := enabledCollectors(c.EnabledCollectors)
	collectors, err := buildCollectors(enabledCollectorNames)
	if err != nil {
		return nil, err
	}

	collectorNames := make([]string, 0, len(collectors))
	for key := range collectors {
		collectorNames = append(collectorNames, key)
	}
	sort.Strings(collectorNames)
	level.Info(logger).Log("msg", "enabled windows_exporter collectors", "collectors", strings.Join(collectorNames, ","))

	return integrations.NewCollectorIntegration(c.Name(), integrations.WithCollectors(
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

func buildCollectors(enabled []string) (map[string]collector.Collector, error) {
	collectors := map[string]collector.Collector{}

	for _, name := range enabled {
		c, err := collector.Build(name)
		if err != nil {
			return nil, err
		}
		collectors[name] = c
	}

	return collectors, nil
}
