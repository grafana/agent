package gcp_metrics_exporter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/rehttp"
	"github.com/go-kit/log"
	"github.com/prometheus-community/stackdriver_exporter/collectors"
	"github.com/prometheus-community/stackdriver_exporter/utils"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/monitoring/v3"
	"google.golang.org/api/option"

	"github.com/grafana/agent/pkg/integrations"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
)

func init() {
	integrations.RegisterIntegration(&Config{})
	integrations_v2.RegisterLegacy(&Config{}, integrations_v2.TypeSingleton, metricsutils.NewNamedShim("gcp_metrics_exporter"))
}

type Config struct {
	// Google Cloud project ID from where we want to scrape metrics from
	ProjectID string `yaml:"project_id"`
	// Comma separated Google Monitoring Metric Type prefixes.
	MetricPrefixes []string `yaml:"metrics_prefixes"`
	// Filters. i.e: pubsub.googleapis.com/subscription:resource.labels.subscription_id=monitoring.regex.full_match("my-subs-prefix.*")
	ExtraFilters []string `yaml:"extra_filters"`
	// How long should the collector wait for a result from the API.
	ClientTimeout time.Duration `yaml:"client_timeout"`
	// Interval to request the Google Monitoring Metrics for. Only the most recent data point is used.
	RequestInterval time.Duration `yaml:"request_interval"`
	// Offset for the Google Stackdriver Monitoring Metrics interval into the past.
	RequestOffset time.Duration `yaml:"request_offset"`
	// Offset for the Google Stackdriver Monitoring Metrics interval into the past by the ingest delay from the metric's metadata.
	IngestDelay bool `yaml:"ingest_delay"`
	// Fill missing metrics labels with empty string to avoid label dimensions inconsistent failure.
	FillMissingLabels bool `yaml:"fill_missing_labels"`
	// Drop metrics from attached projects and fetch `project_id` only.
	DropDelegatedProjects bool `yaml:"drop_delegated_projects"`
	// If enabled will treat all DELTA metrics as an in-memory counter instead of a gauge.
	AggregateDeltas bool `yaml:"aggregate_deltas"`
}

var DefaultConfig = Config{
	ClientTimeout:         15 * time.Second,
	RequestInterval:       5 * time.Minute,
	RequestOffset:         0,
	IngestDelay:           false,
	FillMissingLabels:     true,
	DropDelegatedProjects: false,
	AggregateDeltas:       false,
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

func (c *Config) InstanceKey(_ string) (string, error) {
	return c.Name(), nil
}

func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	svc, err := createMonitoringService(context.Background(), c.ClientTimeout)
	if err != nil {
		return nil, err
	}

	monitoringCollector, err := collectors.NewMonitoringCollector(
		c.ProjectID,
		svc,
		collectors.MonitoringCollectorOptions{
			MetricTypePrefixes:    c.MetricPrefixes,
			ExtraFilters:          parseMetricExtraFilters(c.ExtraFilters),
			RequestInterval:       c.RequestInterval,
			RequestOffset:         c.RequestOffset,
			IngestDelay:           c.IngestDelay,
			FillMissingLabels:     c.FillMissingLabels,
			DropDelegatedProjects: c.DropDelegatedProjects,
			AggregateDeltas:       c.AggregateDeltas,
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

	monitoringService, err := monitoring.NewService(ctx,
		option.WithHTTPClient(googleClient),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating Google Stackdriver Monitoring service: %v", err)
	}

	return monitoringService, nil
}

func parseMetricExtraFilters(filters []string) []collectors.MetricFilter {
	var extraFilters []collectors.MetricFilter
	for _, ef := range filters {
		efPrefix, efModifier := utils.GetExtraFilterModifiers(ef, ":")
		if efPrefix != "" {
			extraFilter := collectors.MetricFilter{
				Prefix:   efPrefix,
				Modifier: efModifier,
			}
			extraFilters = append(extraFilters, extraFilter)
		}
	}
	return extraFilters
}
