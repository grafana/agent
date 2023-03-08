//nolint:golint,goconst
package config

import (
	"strings"
	"testing"

	"github.com/grafana/agent/pkg/integrations/node_exporter"
	"github.com/grafana/agent/pkg/integrations/windows_exporter"

	v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigMaker(t *testing.T) {
	configStr := `wal_directory: /tmp/wal`
	tDir := t.TempDir()
	writeFile(t, tDir, "metrics-1.yml", configStr)
	fileFS := generateFilePath(tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	configs, err := cmf.processMetrics()
	require.NoError(t, err)
	assert.NotNil(t, configs)
	assert.Equal(t, configs.WALDir, "/tmp/wal")
}

func TestConfigMakerWithFakeFiles(t *testing.T) {
	configStr := `wal_directory: /tmp/wal`
	tDir := t.TempDir()
	writeFile(t, tDir, "metrics-1.yml", configStr)
	writeFile(t, tDir, "fake.yml", configStr)
	fileFS := generateFilePath(tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	configs, err := cmf.processMetrics()
	require.NoError(t, err)
	assert.NotNil(t, configs)
	assert.Equal(t, configs.WALDir, "/tmp/wal")
}

func TestConfigMakerWithMultipleMetrics(t *testing.T) {
	configStr := `wal_directory: /tmp/wal`
	tDir := t.TempDir()
	writeFile(t, tDir, "metrics-1.yml", configStr)
	writeFile(t, tDir, "metrics-2.yml", configStr)

	fileFS := generateFilePath(tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	_, err := cmf.processMetrics()
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "found 2 metrics templates; expected 0 or 1"))
}

func TestConfigMakerWithMetricsAndInstances(t *testing.T) {
	configStr := `wal_directory: /tmp/wal`
	tDir := t.TempDir()
	writeFile(t, tDir, "metrics-1.yml", configStr)
	writeFile(t, tDir, "metrics_instances-1.yml", "name: t1")
	writeFile(t, tDir, "metrics_instances-2.yml", "name: t2")
	writeFile(t, tDir, "server-1.yml", `
http_listen_port: 12345
log_level: debug
`)
	fileFS := generateFilePath(tDir)

	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	cfg := &Config{}
	err := cmf.ProcessConfigs(cfg)
	require.NoError(t, err)
	assert.Len(t, cfg.Metrics.Configs, 2)
}

func TestConfigMakerWithExporter(t *testing.T) {
	configStr := `
windows:
  enabled_collectors: one,two,three
`
	tDir := t.TempDir()
	writeFile(t, tDir, "integrations-1.yml", configStr)
	fileFS := generateFilePath(tDir)

	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	configs, err := cmf.processIntegrations()
	require.NoError(t, err)
	require.Len(t, configs, 1)
	wincfg, _ := configs[0].(v2.UpgradedConfig).LegacyConfig()
	assert.True(t, wincfg.(*windows_exporter.Config).EnabledCollectors == "one,two,three")
}

func TestConfigMakerWithMultipleExporter(t *testing.T) {
	configStr := `
windows:
  enabled_collectors: one,two,three
  instance: testinstance
node_exporter:
  autoscrape:
    enable: false
`
	tDir := t.TempDir()
	writeFile(t, tDir, "integrations-1.yml", configStr)
	fileFS := generateFilePath(tDir)

	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	configs, err := cmf.processIntegrations()
	require.NoError(t, err)
	assert.Len(t, configs, 2)
	for _, cfg := range configs {
		switch v := cfg.(type) {
		default:
			t.Errorf("unexpected type %T", v)
		case v2.UpgradedConfig:
			oldConfig, _ := v.LegacyConfig()
			switch oc := oldConfig.(type) {
			case *windows_exporter.Config:
				assert.True(t, "one,two,three" == oc.EnabledCollectors)
			case *node_exporter.Config:
				assert.NotNil(t, v)
			}
		}
	}
}

func TestLoadingFromS3(t *testing.T) {
	configStr := `
windows:
  enabled_collectors: one,two,three
  instance: testinstance
`
	u := pushFilesToFakeS3(t, "integrations-1.yml", configStr)
	s3Url := "s3://mybucket/?region=us-east-1&disableSSL=true&s3ForcePathStyle=true&endpoint=" + u.Host
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{s3Url},
	}
	cmf := generateLoader(t, loaderCfg)
	cfg, err := cmf.processIntegrations()
	require.NoError(t, err)
	assert.Len(t, cfg, 1)
	oc, _ := cfg[0].(v2.UpgradedConfig).LegacyConfig()
	winCfg := oc.(*windows_exporter.Config)
	assert.True(t, winCfg.EnabledCollectors == "one,two,three")
}

func TestMultiplex(t *testing.T) {
	configStr := `
redis_configs:
- redis_addr: localhost:6379
  autoscrape:
    metric_relabel_configs: 
    - source_labels: [__address__]
      target_label: "banana"
      replacement: "apple"
- redis_addr: localhost:6380
`
	tDir := t.TempDir()
	writeFile(t, tDir, "integrations-1.yml", configStr)
	fileFS := generateFilePath(tDir)

	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	cfg := &Config{}

	err := cmf.ProcessConfigs(cfg)
	require.NoError(t, err)
	assert.Len(t, cfg.Integrations.ExtraIntegrations, 2)
}

func TestAgentAddIntegrations(t *testing.T) {
	configStr := `
server:
  log_level: debug
metrics:
  wal_directory: /tmp/grafana-agent-normal
  global:
    scrape_interval: 60s
    remote_write:
    - url: https://www.example.com
integrations:
  node_exporter: {}
`
	addIntegration := `
windows: {}
`
	tDir := t.TempDir()
	writeFile(t, tDir, "agent-1.yml", configStr)
	writeFile(t, tDir, "integrations-windows.yml", addIntegration)
	fileFS := generateFilePath(tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	cfg := &Config{}
	err := cmf.ProcessConfigs(cfg)
	require.NoError(t, err)
	// Since the normal agent uses deferred parsing this is required to load the integration from agent-1.yml
	err = cfg.Integrations.setVersion(integrationsVersion2)
	require.NoError(t, err)
	assert.True(t, cfg.ServerFlags.HTTP.ListenAddress == "127.0.0.1:12345")
	assert.True(t, cfg.Server.LogLevel.String() == "debug")
	assert.True(t, cfg.Metrics.WALDir == "/tmp/grafana-agent-normal")
	assert.True(t, cfg.Metrics.Global.RemoteWrite[0].URL.String() == "https://www.example.com")
	assert.False(t, cfg.Integrations.IsZero())
	// ExtraIngrations should be 1 from the integrations-windows.yml
	// the node_exporter integration in the agent-1.yml is in the configV2
	assert.Len(t, cfg.Integrations.ExtraIntegrations, 1)
	assert.Len(t, cfg.Integrations.configV2.Configs, 2)
}

func TestFilterOverrides(t *testing.T) {
	agentStr := `
server:
  log_level: debug
metrics:
  wal_directory: /tmp/grafana-agent-normal
  global:
    scrape_interval: 60s
    remote_write:
    - url: https://www.example.com
integrations:
  windows: {}
`
	serverStr := `
http_tls_config:
  cert_file: /fake/file.cert
  key_file: /fake/file.key
`
	metricsStr := `
wal_directory: /tmp/grafana-agent-normal
global:
  scrape_interval: 60s
  remote_write:
  - url: https://www.example.com
`
	metricsInstanceStr := `
name: t1
`
	integrationStr := `
node_exporter: {}
`
	tracesStr := `
configs:
- name: test_traces
  automatic_logging:
    backend: stdout
    loki_name: default
    spans: true
`
	logsStr := `
configs:
- name: test_logs
  positions:
    filename: /tmp/positions.yaml
  scrape_configs:
    - job_name: test
      pipeline_stages:
      - regex:
        source: filename
        expression: '\\temp\\Logs\\(?P<log_app>.+?)\\'
`
	tDir := t.TempDir()
	writeFile(t, tDir, "a-1.yml", agentStr)
	writeFile(t, tDir, "s-1.yml", serverStr)
	writeFile(t, tDir, "m-1.yml", metricsStr)
	writeFile(t, tDir, "mi-1.yml", metricsInstanceStr)
	writeFile(t, tDir, "i-1.yml", integrationStr)
	writeFile(t, tDir, "t-1.yml", tracesStr)
	writeFile(t, tDir, "l-1.yml", logsStr)
	fileFS := generateFilePath(tDir)
	loaderCfg := LoaderConfig{
		Sources:               nil,
		TemplatePaths:         []string{fileFS},
		AgentFilter:           "a-*.yml",
		ServerFilter:          "s-*.yml",
		MetricsFilter:         "m-*.yml",
		MetricsInstanceFilter: "mi-*.yml",
		IntegrationsFilter:    "i-*.yml",
		LogsFilter:            "l-*.yml",
		TracesFilter:          "t-*.yml",
	}
	cmf := generateLoader(t, loaderCfg)
	cfg := &Config{}
	err := cmf.ProcessConfigs(cfg)
	require.NoError(t, err)
	// Since the normal agent uses deferred parsing this is required to load the integration from agent-1.yml
	err = cfg.Integrations.setVersion(integrationsVersion2)
	require.NoError(t, err)
	// Test server override
	assert.Equal(t, "/fake/file.cert", cfg.Server.HTTP.TLSConfig.TLSCertPath)
	assert.Equal(t, "/fake/file.key", cfg.Server.HTTP.TLSConfig.TLSKeyPath)
	// Test metric
	assert.True(t, cfg.Metrics.WALDir == "/tmp/grafana-agent-normal")
	// Test Metric Instances
	assert.True(t, cfg.Metrics.Configs[0].Name == "t1")
	// Test Integrations
	assert.Len(t, cfg.Integrations.ExtraIntegrations, 1)
	assert.Len(t, cfg.Integrations.configV2.Configs, 2)
	// Test Traces
	assert.True(t, cfg.Traces.Configs[0].Name == "test_traces")
	// Test Logs
	assert.True(t, cfg.Logs.Configs[0].Name == "test_logs")
}
