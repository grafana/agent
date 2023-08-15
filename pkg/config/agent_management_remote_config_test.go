package config

import (
	"testing"
	"time"

	"github.com/grafana/agent/pkg/metrics/instance"
	"github.com/stretchr/testify/require"
)

func TestBuildRemoteConfig(t *testing.T) {
	baseConfig := `
server:
    log_level: debug
`
	metricsSnippets := []Snippet{{
		Config: `
metrics_scrape_configs:
  - job_name: 'prometheus'
    scrape_interval: 15s
    static_configs:
    - targets: ['localhost:9090']
`,
	}}
	logsSnippets := []Snippet{{
		Config: `
logs_scrape_configs:
  - job_name: 'loki'
    static_configs:
    - targets: ['localhost:3100']
`,
	}}
	integrationSnippets := []Snippet{{
		Config: `
integration_configs:
  agent:
    enabled: true
    relabel_configs:
      - action: replace
        source_labels:
          - agent_hostname
        target_label: instance
`,
	}}

	allSnippets := []Snippet{}
	allSnippets = append(allSnippets, metricsSnippets...)
	allSnippets = append(allSnippets, logsSnippets...)
	allSnippets = append(allSnippets, integrationSnippets...)

	t.Run("only metrics snippets provided", func(t *testing.T) {
		rc := RemoteConfig{
			BaseConfig: BaseConfigContent(baseConfig),
			Snippets:   metricsSnippets,
		}
		c, err := rc.BuildAgentConfig()
		require.NoError(t, err)
		require.Equal(t, len(c.Metrics.Configs), 1)
		require.Empty(t, c.Logs)
		require.Empty(t, c.Integrations.ConfigV1.Integrations)
	})

	t.Run("only log snippets provided", func(t *testing.T) {
		rc := RemoteConfig{
			BaseConfig: BaseConfigContent(baseConfig),
			Snippets:   logsSnippets,
		}
		c, err := rc.BuildAgentConfig()
		require.NoError(t, err)
		require.Equal(t, len(c.Logs.Configs), 1)
		require.Empty(t, c.Metrics.Configs)
		require.Empty(t, c.Integrations.ConfigV1.Integrations)
	})

	t.Run("only integration snippets provided", func(t *testing.T) {
		rc := RemoteConfig{
			BaseConfig: BaseConfigContent(baseConfig),
			Snippets:   integrationSnippets,
		}
		c, err := rc.BuildAgentConfig()
		require.NoError(t, err)
		require.Empty(t, c.Metrics.Configs)
		require.Empty(t, c.Logs)
		require.Equal(t, 1, len(c.Integrations.ConfigV1.Integrations))
	})

	t.Run("base with already logs, metrics and integrations provided", func(t *testing.T) {
		fullConfig := `
metrics:
  configs:
  - name: default
    scrape_configs:
    - job_name: default-prom
      static_configs:
      - targets:
        - localhost:9090
logs:
  positions_directory: /tmp/grafana-agent-positions
  configs:
  - name: default
    scrape_configs:
    - job_name: default-loki
      static_configs:
      - targets:
        - localhost:3100
integrations:
  node_exporter:
    enabled: true
`
		rc := RemoteConfig{
			BaseConfig: BaseConfigContent(fullConfig),
			Snippets:   allSnippets,
		}
		c, err := rc.BuildAgentConfig()
		require.NoError(t, err)
		require.Equal(t, len(c.Logs.Configs), 2)
		require.Equal(t, len(c.Metrics.Configs), 2)
		require.Equal(t, 2, len(c.Integrations.ConfigV1.Integrations))
	})

	t.Run("all snippets provided", func(t *testing.T) {
		rc := RemoteConfig{
			BaseConfig: BaseConfigContent(baseConfig),
			Snippets:   allSnippets,
		}
		c, err := rc.BuildAgentConfig()
		require.NoError(t, err)
		require.Equal(t, 1, len(c.Logs.Configs))
		require.Equal(t, 1, len(c.Metrics.Configs))
		require.Equal(t, 1, len(c.Integrations.ConfigV1.Integrations))

		// check some fields to make sure the config was parsed correctly
		require.Equal(t, "prometheus", c.Metrics.Configs[0].ScrapeConfigs[0].JobName)
		require.Equal(t, "loki", c.Logs.Configs[0].ScrapeConfig[0].JobName)
		require.Equal(t, "agent", c.Integrations.ConfigV1.Integrations[0].Name())

		// make sure defaults for metric snippets are applied
		require.Equal(t, instance.DefaultConfig.WALTruncateFrequency, c.Metrics.Configs[0].WALTruncateFrequency)
		require.Equal(t, instance.DefaultConfig.HostFilter, c.Metrics.Configs[0].HostFilter)
		require.Equal(t, instance.DefaultConfig.MinWALTime, c.Metrics.Configs[0].MinWALTime)
		require.Equal(t, instance.DefaultConfig.MaxWALTime, c.Metrics.Configs[0].MaxWALTime)
		require.Equal(t, instance.DefaultConfig.RemoteFlushDeadline, c.Metrics.Configs[0].RemoteFlushDeadline)
		require.Equal(t, instance.DefaultConfig.WriteStaleOnShutdown, c.Metrics.Configs[0].WriteStaleOnShutdown)
		require.Equal(t, instance.DefaultGlobalConfig, c.Metrics.Global)

		// make sure defaults for log snippets are applied
		require.Equal(t, 10*time.Second, c.Logs.Configs[0].PositionsConfig.SyncPeriod)
		require.Equal(t, "", c.Logs.Configs[0].PositionsConfig.PositionsFile)
		require.Equal(t, false, c.Logs.Configs[0].PositionsConfig.IgnoreInvalidYaml)
		require.Equal(t, false, c.Logs.Configs[0].TargetConfig.Stdin)

		// make sure defaults for integration snippets are applied
		require.Equal(t, true, c.Integrations.ConfigV1.ScrapeIntegrations)
		require.Equal(t, true, c.Integrations.ConfigV1.UseHostnameLabel)
		require.Equal(t, true, c.Integrations.ConfigV1.ReplaceInstanceLabel)
		require.Equal(t, 5*time.Second, c.Integrations.ConfigV1.IntegrationRestartBackoff)
	})
}
