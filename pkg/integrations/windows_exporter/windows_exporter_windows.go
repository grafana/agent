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
	// We need to create a list of all the possible collectors.
	collectors := collector.CreateInitializers()
	// We need to pass in kingpin so that the settings get created appropriately. Even though we arent going to use its output.
	windowsExporter := kingpin.New("", "")
	// We only need this to fill in the appropriate settings structs so we can override them.
	collector.RegisterCollectorsFlags(collectors, windowsExporter)
	// Override the settings structs generated from the kingping.app switch our own.
	err := c.toExporterConfig(collectors)
	if err != nil {
		return nil, err
	}
	// Register the performance monitors
	collector.RegisterCollectors(collectors)
	// Filter down to the enabled collectors
	enabledCollectorNames := enabledCollectors(c.EnabledCollectors)
	// Finally build the collectors that we need to run.
	builtCollectors, err := buildCollectors(collectors, enabledCollectorNames)
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
		collector.NewPrometheus(4*time.Minute, builtCollectors),
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

func buildCollectors(colls map[string]*collector.Initializer, enabled []string) (map[string]collector.Collector, error) {
	collectors := map[string]collector.Collector{}

	for _, name := range enabled {
		c, err := collector.Build(colls, name)
		if err != nil {
			return nil, err
		}
		collectors[name] = c
	}

	return collectors, nil
}
