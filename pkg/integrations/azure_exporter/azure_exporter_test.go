package azure_exporter_test

import (
	"math"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/require"
	"github.com/webdevops/azure-metrics-exporter/metrics"

	"github.com/grafana/agent/pkg/integrations/azure_exporter"
)

func TestConfig_NewIntegration_ConfigToSettings(t *testing.T) {
	logger := log.NewJSONLogger(os.Stdout)
	baseConfig := azure_exporter.Config{
		Subscriptions:            []string{"subscriptionA"},
		ResourceType:             "resourceType",
		ResourceGraphQueryFilter: "filter_me",
		Metrics:                  []string{"MetricA"},
		MetricAggregations:       []string{"MiNimUm"},
		Timespan:                 "timespan_me",
		IncludedResourceTags:     []string{"tag_me"},
		MetricNamespace:          "namespace_me",
		MetricNameTemplate:       "name_template_me",
		MetricHelpTemplate:       "help_template_me",
		AzureCloudEnvironment:    "azurecloud",
	}
	baseSettings := metrics.RequestMetricSettings{
		Name:            "not_used",
		Subscriptions:   []string{"subscriptionA"},
		ResourceType:    "resourceType",
		Filter:          "filter_me",
		Timespan:        "timespan_me",
		Interval:        to.Ptr("timespan_me"),
		Metrics:         []string{"MetricA"},
		MetricNamespace: "namespace_me",
		Aggregations:    []string{"MiNimUm"},
		MetricTemplate:  "name_template_me",
		HelpTemplate:    "help_template_me",
		TagLabels:       []string{"tag_me"},
		//unused
		Target:          nil,
		MetricTop:       nil,
		MetricFilter:    "",
		MetricOrderBy:   "",
		Cache:           nil,
		ResourceSubPath: "",
	}

	baseConfigValid := t.Run("maps expected fields", func(t *testing.T) {
		integration, err := baseConfig.NewIntegration(logger)
		require.NoError(t, err)
		require.NotNil(t, integration)
		require.IsType(t, integration, azure_exporter.Exporter{})

		azureExporter := integration.(azure_exporter.Exporter)
		require.Equal(t, &baseSettings, azureExporter.Settings)
	})
	if !baseConfigValid {
		return
	}

	tests := []struct {
		name             string
		configModifier   func(azure_exporter.Config) azure_exporter.Config
		settingsModifier func(metrics.RequestMetricSettings) metrics.RequestMetricSettings
	}{
		{
			name: "maps ResourceSubPath for the storageaccounts MetricNamespace",
			configModifier: func(config azure_exporter.Config) azure_exporter.Config {
				config.MetricNamespace = "microsoft.storage/storageaccounts/blobby"
				return config
			},
			settingsModifier: func(settings metrics.RequestMetricSettings) metrics.RequestMetricSettings {
				settings.MetricNamespace = "microsoft.storage/storageaccounts/blobby"
				settings.ResourceSubPath = "/blobby/default"
				return settings
			},
		},
		{
			name: "can set a metric filter for a single dimension",
			configModifier: func(config azure_exporter.Config) azure_exporter.Config {
				config.IncludedDimensions = []string{"dimension1"}
				return config
			},
			settingsModifier: func(settings metrics.RequestMetricSettings) metrics.RequestMetricSettings {
				settings.MetricFilter = "dimension1 eq '*'"
				settings.MetricTop = to.Ptr[int32](math.MaxInt32)
				return settings
			},
		},
		{
			name: "can set a metric filter for a multiple dimensions",
			configModifier: func(config azure_exporter.Config) azure_exporter.Config {
				config.IncludedDimensions = []string{"dimension1", "dimension2", "dimension3"}
				return config
			},
			settingsModifier: func(settings metrics.RequestMetricSettings) metrics.RequestMetricSettings {
				settings.MetricFilter = "dimension1 eq '*' and dimension2 eq '*' and dimension3 eq '*'"
				settings.MetricTop = to.Ptr[int32](math.MaxInt32)
				return settings
			},
		},
		{
			name: "sets config timespan to setting interval and timespan",
			configModifier: func(config azure_exporter.Config) azure_exporter.Config {
				config.Timespan = "timespan-value"
				return config
			},
			settingsModifier: func(settings metrics.RequestMetricSettings) metrics.RequestMetricSettings {
				settings.Timespan = "timespan-value"
				settings.Interval = to.Ptr[string]("timespan-value")
				return settings
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullConfig := tt.configModifier(baseConfig)
			integration, err := fullConfig.NewIntegration(logger)
			require.NoError(t, err)
			require.NotNil(t, integration)
			require.IsType(t, integration, azure_exporter.Exporter{})

			azureExporter := integration.(azure_exporter.Exporter)
			expectedSettings := tt.settingsModifier(baseSettings)
			require.Equal(t, &expectedSettings, azureExporter.Settings)
		})
	}
}

func TestConfig_NewIntegration_Invalid_Config(t *testing.T) {
	logger := log.NewJSONLogger(os.Stdout)
	baseConfig := azure_exporter.Config{
		Subscriptions:         []string{"subscriptionA"},
		ResourceType:          "resourceType",
		Metrics:               []string{"MetricA"},
		AzureCloudEnvironment: "azurecloud",
	}

	baseConfigValid := t.Run("Base Config is Valid", func(t *testing.T) {
		_, err := baseConfig.NewIntegration(logger)
		require.NoError(t, err, "Base config was not valid but needs to be for these tests")
	})
	if !baseConfigValid {
		return
	}

	tests := []struct {
		name           string
		configModifier func(azure_exporter.Config) azure_exporter.Config
	}{
		{
			name: "nil Subscriptions",
			configModifier: func(config azure_exporter.Config) azure_exporter.Config {
				config.Subscriptions = nil
				return config
			},
		},
		{
			name: "empty Subscriptions",
			configModifier: func(config azure_exporter.Config) azure_exporter.Config {
				config.Subscriptions = []string{}
				return config
			},
		},
		{
			name: "empty ResourceType",
			configModifier: func(config azure_exporter.Config) azure_exporter.Config {
				config.ResourceType = ""
				return config
			},
		},
		{
			name: "nil metrics",
			configModifier: func(config azure_exporter.Config) azure_exporter.Config {
				config.Metrics = nil
				return config
			},
		},
		{
			name: "empty metrics",
			configModifier: func(config azure_exporter.Config) azure_exporter.Config {
				config.Metrics = []string{}
				return config
			},
		},
		{
			name: "invalid aggregation",
			configModifier: func(config azure_exporter.Config) azure_exporter.Config {
				config.MetricAggregations = []string{"I'm Invalid"}
				return config
			},
		},
		{
			name: "invalid azure_cloud_environment",
			configModifier: func(config azure_exporter.Config) azure_exporter.Config {
				config.AzureCloudEnvironment = "Not Real"
				return config
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invalidConfig := tt.configModifier(baseConfig)
			_, err := invalidConfig.NewIntegration(logger)
			require.Error(t, err)
		})
	}
}
