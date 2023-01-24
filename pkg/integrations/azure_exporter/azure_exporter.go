package azure_exporter

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	azure_config "github.com/webdevops/azure-metrics-exporter/config"
	"github.com/webdevops/azure-metrics-exporter/metrics"
	"github.com/webdevops/go-common/azuresdk/armclient"
	"github.com/webdevops/go-common/azuresdk/cloudconfig"

	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
)

func init() {
	integrations.RegisterIntegration(&Config{})
	integrations_v2.RegisterLegacy(&Config{}, integrations_v2.TypeMultiplex, metricsutils.NewNamedShim("azure"))
}

// DefaultConfig holds the default settings for the azure_exporter integration.
var DefaultConfig = Config{
	Timespan:              "PT1M",
	MetricNameTemplate:    "azure_{type}_{metric}_{aggregation}_{unit}",
	MetricHelpTemplate:    "Azure metric {metric} for {type} with aggregation {aggregation} as {unit}",
	IncludedResourceTags:  []string{"owner"},
	AzureCloudEnvironment: "azurecloud",
}

type Config struct {
	Subscriptions            []string `yaml:"subscriptions"`               // Required
	ResourceGraphQueryFilter string   `yaml:"resource_graph_query_filter"` // Required

	// Valid values can be derived from https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/metrics-supported
	// Required: Root level names ex. Microsoft.DataShare/accounts
	ResourceType string `yaml:"resource_type"`
	// Required: Metric in the table for a ResourceType
	Metrics []string `yaml:"metrics"`
	// Optional: If not provided Aggregation Type is used for the Metric
	// Valid values are minimum, maximum, average, total, and count
	MetricAggregations []string `yaml:"metric_aggregations"`

	// All fields below are optional
	// Must be an ISO8601 Duration - defaults to PT1M if not specified
	Timespan             string   `yaml:"timespan"`
	IncludedDimensions   []string `yaml:"included_dimensions"`
	IncludedResourceTags []string `yaml:"included_resource_tags"`

	// MetricNamespace is used for ResourceTypes which have multiple levels of metrics
	// As an example the ResourceType Microsoft.Storage/storageAccounts has metrics for
	//  Microsoft.Storage/storageAccounts (generic metrics which apply to all storage accounts)
	//	Microsoft.Storage/storageAccounts/blobServices (generic metrics + metrics which only apply to blob stores)
	// 	Microsoft.Storage/storageAccounts/fileServices (generic metrics + metrics which only apply to file stores)
	//	Microsoft.Storage/storageAccounts/queueServices (generic metrics + metrics which only apply to queue stores)
	//	Microsoft.Storage/storageAccounts/tableServices (generic metrics + metrics which only apply to table stores)
	// If you want blob store metrics you will need to set
	//  ResourceType = Microsoft.Storage/storageAccounts
	//	MetricNamespace = Microsoft.Storage/storageAccounts/blobServices
	MetricNamespace string `yaml:"metric_namespace"`

	MetricNameTemplate string `yaml:"metric_name_template"`
	MetricHelpTemplate string `yaml:"metric_help_template"`

	AzureCloudEnvironment string `yaml:"azure_cloud_environment"`

	Common config.Common `yaml:",inline"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

func (c *Config) InstanceKey(_ string) (string, error) {
	if c.Common.InstanceKey != nil {
		return *c.Common.InstanceKey, nil
	}

	// This will ensure compatibility for running multiple instances in v2 without setting an explicit InstanceKey
	instanceKey := fmt.Sprintf("%s:%s:%s", c.Name(), strings.Join(c.Subscriptions, ","), c.ResourceType)

	return instanceKey, nil
}

func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	if err := validateConfig(c); err != nil {
		return nil, err
	}

	settings := &metrics.RequestMetricSettings{
		Subscriptions:   c.Subscriptions,
		Metrics:         c.Metrics,
		ResourceType:    c.ResourceType,
		TagLabels:       c.IncludedResourceTags,
		Aggregations:    c.MetricAggregations,
		Filter:          c.ResourceGraphQueryFilter,
		MetricNamespace: c.MetricNamespace,
		MetricTemplate:  c.MetricNameTemplate,
		HelpTemplate:    c.MetricHelpTemplate,

		// Interval controls data aggregation timeframe ie 1 minute or 5 minutes aggregations
		// Timespan controls query start and end time
		// Interval == Timespan to ensure we only get a single data point any more than that is useless
		Timespan: c.Timespan,
		Interval: to.Ptr[string](c.Timespan),

		// Unused settings just here to capture they are intentionally not set
		Name:  "",
		Cache: nil,
	}

	// Dimensions can only be retrieved via an obscure manner of including a "metric filter" on the query
	// This isn't documented in the Azure API only the exporter: https://github.com/webdevops/azure-metrics-exporter#virtualnetworkgateway-connections-dimension-support
	if len(c.IncludedDimensions) > 0 {
		builder := strings.Builder{}
		for index, dimension := range c.IncludedDimensions {
			if _, err := builder.WriteString(dimension); err != nil {
				return nil, err
			}
			if _, err := builder.WriteString(" eq '*'"); err != nil {
				return nil, err
			}
			// Not the last dimension add an `and`
			if index != (len(c.IncludedDimensions) - 1) {
				if _, err := builder.WriteString(" and "); err != nil {
					return nil, err
				}
			}
		}
		settings.MetricFilter = builder.String()

		// The metric filter introduces a secondary complexity where data is limited by a "top" parameter (default 10)
		// We don't get any knowledge if the result is cut off and there's no support for paging
		// set the value as high as possible to hopefully prevent issues
		settings.MetricTop = to.Ptr[int32](math.MaxInt32)
		settings.MetricOrderBy = "" // Order is only relevant if top won't return all the results our high value should prevent this
	}

	concurrencyConfig := azure_config.Opts{
		Prober: struct {
			// Limits the number of subscriptions which can concurrently be sending metric requests - value taken from OSS exporter
			ConcurrencySubscription int `long:"concurrency.subscription"          env:"CONCURRENCY_SUBSCRIPTION"           description:"Concurrent subscription fetches"                                  default:"5"`
			// Limits the number of concurrent metric requests for a single subscription  - value taken from OSS exporter
			ConcurrencySubscriptionResource int  `long:"concurrency.subscription.resource" env:"CONCURRENCY_SUBSCRIPTION_RESOURCE"  description:"Concurrent requests per resource (inside subscription requests)"  default:"10"`
			Cache                           bool `long:"enable-caching"                    env:"ENABLE_CACHING"                     description:"Enable internal caching"`
		}{5, 10, false},
	}

	logrusLogger := integrations.NewLogger(l)
	logEntry := logrusLogger.WithFields(logrus.Fields{
		"resource_type":               c.ResourceType,
		"resource_graph_query_filter": c.ResourceGraphQueryFilter,
		"subscriptions":               c.Subscriptions,
		"metric_namespace":            c.MetricNamespace,
		"metrics":                     c.Metrics,
	})
	return Exporter{
		cfg:               c,
		logger:            logrusLogger,
		logEntry:          logEntry,
		Settings:          settings,
		ConcurrencyConfig: concurrencyConfig,
	}, nil
}

func validateConfig(c *Config) error {
	var configErrors []string

	if c.Subscriptions == nil || len(c.Subscriptions) == 0 {
		configErrors = append(configErrors, "subscriptions cannot be empty")
	}

	if c.ResourceType == "" {
		configErrors = append(configErrors, "resource_type cannot be empty")
	}

	if c.Metrics == nil || len(c.Metrics) == 0 {
		configErrors = append(configErrors, "metrics cannot be empty")
	}

	validAggregations := []string{"minimum", "maximum", "average", "total", "count"}

	for _, aggregation := range c.MetricAggregations {
		lowerCaseAggregation := strings.ToLower(aggregation)
		found := false
		for _, validAggregation := range validAggregations {
			if validAggregation == lowerCaseAggregation {
				found = true
				break
			}
		}
		if !found {
			configErrors = append(configErrors, fmt.Sprintf("%s is an invalid value for metric_aggregations. Valid options are %s", aggregation, strings.Join(validAggregations, ",")))
		}
	}

	if _, err := cloudconfig.NewCloudConfig(c.AzureCloudEnvironment); err != nil {
		configErrors = append(configErrors, fmt.Errorf("failed to create an azure cloud configuration from azure cloud environment %s, %v", c.AzureCloudEnvironment, err).Error())
	}

	if len(configErrors) != 0 {
		return errors.New(strings.Join(configErrors, "\n"))
	}

	return nil
}

func (c *Config) Name() string {
	return "azure_exporter"
}

type Exporter struct {
	cfg               *Config
	logger            *logrus.Logger // used by azure client
	logEntry          *logrus.Entry  // used by oss exporter
	Settings          *metrics.RequestMetricSettings
	ConcurrencyConfig azure_config.Opts
}

func (e Exporter) MetricsHandler() (http.Handler, error) {
	reg := prometheus.NewRegistry()
	ctx := context.Background()
	prober := metrics.NewMetricProber(ctx, e.logEntry, nil, e.Settings, e.ConcurrencyConfig)

	client, err := armclient.NewArmClientWithCloudName(e.cfg.AzureCloudEnvironment, e.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure client, %v", err)
	}
	prober.SetAzureClient(client)
	prober.SetPrometheusRegistry(reg)

	err = prober.ServiceDiscovery.FindResourceGraph(ctx, e.Settings.Subscriptions, e.Settings.ResourceType, e.Settings.Filter)
	if err != nil {
		return nil, fmt.Errorf("service discovery failed, %v", err)
	}

	prober.Run()

	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{}), nil
}

func (e Exporter) ScrapeConfigs() []config.ScrapeConfig {
	return []config.ScrapeConfig{{JobName: e.cfg.Name(), MetricsPath: "/metrics"}}
}

func (e Exporter) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
