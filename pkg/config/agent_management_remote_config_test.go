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

	bothSnippets := append(metricsSnippets, logsSnippets...)

	t.Run("only metrics snippets provided", func(t *testing.T) {
		rc := RemoteConfig{
			BaseConfig: BaseConfigContent(baseConfig),
			Snippets:   metricsSnippets,
		}
		c, err := rc.BuildAgentConfig()
		require.NoError(t, err)
		require.Equal(t, len(c.Metrics.Configs), 1)
		require.Empty(t, c.Logs)
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
	})

	t.Run("base with already logs and metrics provided", func(t *testing.T) {
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
`
		rc := RemoteConfig{
			BaseConfig: BaseConfigContent(fullConfig),
			Snippets:   bothSnippets,
		}
		c, err := rc.BuildAgentConfig()
		require.NoError(t, err)
		require.Equal(t, len(c.Logs.Configs), 2)
		require.Equal(t, len(c.Metrics.Configs), 2)
	})

	t.Run("both snippets provided", func(t *testing.T) {
		rc := RemoteConfig{
			BaseConfig: BaseConfigContent(baseConfig),
			Snippets:   bothSnippets,
		}
		c, err := rc.BuildAgentConfig()
		require.NoError(t, err)
		require.Equal(t, len(c.Logs.Configs), 1)
		require.Equal(t, len(c.Metrics.Configs), 1)

		// check some fields to make sure the config was parsed correctly
		require.Equal(t, "prometheus", c.Metrics.Configs[0].ScrapeConfigs[0].JobName)
		require.Equal(t, "loki", c.Logs.Configs[0].ScrapeConfig[0].JobName)

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
	})
}
