package gcp

import (
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/gcp_exporter"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.gcp",
		Args:    Arguments{},
		Exports: exporter.Exports{},
		Build:   exporter.New(createExporter, "gcp"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return a.Convert().NewIntegration(opts.Logger)
}

var DefaultArguments = Arguments{
	ClientTimeout:         15 * time.Second,
	RequestInterval:       5 * time.Minute,
	RequestOffset:         0,
	IngestDelay:           false,
	DropDelegatedProjects: false,
}

type Arguments struct {
	// Google Cloud project ID from where we want to scrape metrics from
	ProjectIDs []string `river:"project_ids,attr"`
	// Comma separated Google Monitoring Metric Type prefixes.
	MetricPrefixes []string `river:"metrics_prefixes,attr"`
	// Filters. i.e: pubsub.googleapis.com/subscription:resource.labels.subscription_id=monitoring.regex.full_match("my-subs-prefix.*")
	ExtraFilters []string `river:"extra_filters,attr,optional"`
	// Interval to request the Google Monitoring Metrics for. Only the most recent data point is used.
	RequestInterval time.Duration `river:"request_interval,attr,optional"`
	// Offset for the Google Stackdriver Monitoring Metrics interval into the past.
	RequestOffset time.Duration `river:"request_offset,attr,optional"`
	// Offset for the Google Stackdriver Monitoring Metrics interval into the past by the ingest delay from the metric's metadata.
	IngestDelay bool `river:"ingest_delay,attr,optional"`
	// Drop metrics from attached projects and fetch `project_id` only.
	DropDelegatedProjects bool `river:"drop_delegated_projects,attr,optional"`
	// How long should the collector wait for a result from the API.
	ClientTimeout time.Duration `river:"gcp_client_timeout,attr,optional"`
}

// SetToDefault implements river.Defaulter.
func (a *Arguments) SetToDefault() {
	*a = DefaultArguments
}

// Validate implements river.Validator.
func (a *Arguments) Validate() error {
	err := a.Convert().Validate()
	if err != nil {
		return err
	}
	return nil
}

func (a *Arguments) Convert() *gcp_exporter.Config {
	// NOTE(tburgessdev): this works because we can set up this exporter's Arguments struct
	// to have the exact same field types as the gcp_exporter.Config struct.
	c := gcp_exporter.Config(*a)
	return &c
}
