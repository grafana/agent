package gcp_exporter_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafana/agent/pkg/integrations/gcp_exporter"
)

func TestConfig_Validate(t *testing.T) {
	baseConfig := gcp_exporter.Config{
		ProjectID:             "project1",
		MetricPrefixes:        []string{"prefix1"},
		ExtraFilters:          nil,
		ClientTimeout:         0,
		RequestInterval:       0,
		RequestOffset:         0,
		IngestDelay:           false,
		DropDelegatedProjects: false,
	}

	t.Run("Base Config is Valid", func(t *testing.T) {
		err := baseConfig.Validate()
		require.NoError(t, err, "Base config was not valid but needs to be for these tests")
	})

	tests := []struct {
		name           string
		configModifier func(config gcp_exporter.Config) gcp_exporter.Config
	}{
		{
			name: "empty ProjectID",
			configModifier: func(config gcp_exporter.Config) gcp_exporter.Config {
				config.ProjectID = ""
				return config
			},
		},
		{
			name: "nil MetricPrefixes",
			configModifier: func(config gcp_exporter.Config) gcp_exporter.Config {
				config.MetricPrefixes = nil
				return config
			},
		},
		{
			name: "empty MetricPrefixes",
			configModifier: func(config gcp_exporter.Config) gcp_exporter.Config {
				config.MetricPrefixes = []string{}
				return config
			},
		},
		{
			name: "extraFilter which does not match a MetricPrefix",
			configModifier: func(config gcp_exporter.Config) gcp_exporter.Config {
				config.MetricPrefixes = []string{"gcp.service/logging"}
				config.ExtraFilters = []string{`logging:resource.name==\"my_resource"`}
				return config
			},
		},
		{
			name: "1 extraFilter which matches a MetricPrefix and 1 which does not",
			configModifier: func(config gcp_exporter.Config) gcp_exporter.Config {
				config.MetricPrefixes = []string{
					"gcp.service/logging",
					"gcp.service/compute",
				}
				config.ExtraFilters = []string{
					`gcp.service/logging:resource.name=="my_resource"`,
					`gcp.service/notcompute:compute_instance.name=="instance_a"`,
				}
				return config
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.configModifier(baseConfig)
			err := config.Validate()
			require.Error(t, err)
		})
	}
}
