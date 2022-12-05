package gcp_metrics_exporter

import (
	"context"
	"fmt"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
	"net/http"
	"time"

	"github.com/PuerkitoBio/rehttp"
	"github.com/go-kit/log"
	"github.com/prometheus-community/stackdriver_exporter/collectors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/monitoring/v3"
	"google.golang.org/api/option"

	"github.com/grafana/agent/pkg/integrations"
)

func init() {
	integrations.RegisterIntegration(&Config{})
	integrations_v2.RegisterLegacy(&Config{}, integrations_v2.TypeSingleton, metricsutils.NewNamedShim("gcp_metrics_exporter"))
}

type Config struct {
	ProjectID      string        `yaml:"project_id"`
	ClientTimeout  time.Duration `yaml:"client_timeout"`
	MetricPrefixes []string      `yaml:"metrics_prefixes"`
}

var DefaultConfig = Config{
	ProjectID:      "1",
	ClientTimeout:  15 * time.Second,
	MetricPrefixes: []string{},
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

func (c *Config) Name() string {
	return "gcp_metrics_exporter"
}

func (c *Config) InstanceKey(agentKey string) (string, error) {
	//TODO(daniele) find something
	return c.Name(), nil
}

func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc, err := createMonitoringService(ctx, c.ClientTimeout)
	if err != nil {
		return nil, err
	}

	// TODO(daniele) fill counterStore and distributionStore
	monitoringCollector, err := collectors.NewMonitoringCollector(
		c.ProjectID,
		svc,
		collectors.MonitoringCollectorOptions{
			// TODO(daniele) fill options - using the default flags from stackdriver CLI
			MetricTypePrefixes:    c.MetricPrefixes,
			ExtraFilters:          nil,
			RequestInterval:       5 * time.Minute,
			RequestOffset:         0,
			IngestDelay:           false,
			FillMissingLabels:     true,
			DropDelegatedProjects: false,
			AggregateDeltas:       false,
		},
		l,
		collectors.NewInMemoryDeltaCounterStore(l, 30*time.Minute),
		collectors.NewInMemoryDeltaDistributionStore(l, 30*time.Minute),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create monitoring collector: %w", err)
	}

	return integrations.NewCollectorIntegration(
		c.Name(), integrations.WithCollectors(monitoringCollector),
	), nil
}

func createMonitoringService(ctx context.Context, httpTimeout time.Duration) (*monitoring.Service, error) {
	googleClient, err := google.DefaultClient(ctx, monitoring.MonitoringReadScope)
	if err != nil {
		return nil, fmt.Errorf("error creating Google client: %v", err)
	}

	googleClient.Timeout = httpTimeout
	// TODO(daniele) - using the default flags from stackdriver CLI
	googleClient.Transport = rehttp.NewTransport(
		googleClient.Transport,
		rehttp.RetryAll(
			rehttp.RetryMaxRetries(4),
			rehttp.RetryStatuses(http.StatusServiceUnavailable)),
		rehttp.ExpJitterDelay(time.Second, 5*time.Second),
	)

	monitoringService, err := monitoring.NewService(ctx, option.WithHTTPClient(googleClient))
	if err != nil {
		return nil, fmt.Errorf("error creating Google Stackdriver Monitoring service: %v", err)
	}

	return monitoringService, nil
}
