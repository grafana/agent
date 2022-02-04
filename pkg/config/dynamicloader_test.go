//nolint:golint,goconst
package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/grafana/agent/pkg/util/subset"
	"gopkg.in/yaml.v2"

	_ "github.com/grafana/agent/pkg/integrations/install"
	"github.com/grafana/agent/pkg/integrations/node_exporter"
	v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/windows_exporter"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/assert"
)

func TestConfigMaker(t *testing.T) {
	configStr := `wal_directory: /tmp/wal`
	tDir := generatePath(t)
	writeFile(t, tDir, "metrics-1.yml", configStr)
	fileFS := generateFilePath(tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	configs, err := cmf.processMetric()
	assert.Nil(t, err)
	assert.NotNil(t, configs)
	assert.Equal(t, configs.WALDir, "/tmp/wal")
}

func TestConfigMakerWithFakeFiles(t *testing.T) {
	configStr := `wal_directory: /tmp/wal`
	tDir := generatePath(t)
	writeFile(t, tDir, "metrics-1.yml", configStr)
	writeFile(t, tDir, "fake.yml", configStr)
	fileFS := generateFilePath(tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	configs, err := cmf.processMetric()
	assert.Nil(t, err)
	assert.NotNil(t, configs)
	assert.Equal(t, configs.WALDir, "/tmp/wal")
}

func TestConfigMakerWithMultipleMetrics(t *testing.T) {
	configStr := `wal_directory: /tmp/wal`
	tDir := generatePath(t)
	writeFile(t, tDir, "metrics-1.yml", configStr)
	writeFile(t, tDir, "metrics-2.yml", configStr)

	fileFS := generateFilePath(tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	_, err := cmf.processMetric()
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "multiple metrics configurations found")
}

func TestConfigMakerWithMetricsAndInstances(t *testing.T) {
	configStr := `wal_directory: /tmp/wal`
	tDir := generatePath(t)
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
	err := cmf.ProcessConfigs(cfg, nil)
	assert.Nil(t, err)
	assert.Len(t, cfg.Metrics.Configs, 2)
}

func TestConfigMakerWithExporter(t *testing.T) {
	configStr := `
windows_exporter:
  enabled_collectors: one,two,three
`
	tDir := generatePath(t)
	writeFile(t, tDir, "integrations-1.yml", configStr)
	fileFS := generateFilePath(tDir)

	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	configs, err := cmf.processIntegrations()
	assert.NoError(t, err)
	assert.Len(t, configs, 1)
	wincfg, _ := configs[0].(v2.UpgradedConfig).LegacyConfig()
	assert.True(t, wincfg.(*windows_exporter.Config).EnabledCollectors == "one,two,three")
}

func TestConfigMakerSingletonWithExporter(t *testing.T) {
	configStr := `
windows_exporter:
  enabled_collectors: one,two,three
  instance: testinstance
`
	tDir := generatePath(t)
	writeFile(t, tDir, "integrations-1.yml", configStr)
	writeFile(t, tDir, "integrations-2.yml", configStr)

	fileFS := generateFilePath(tDir)

	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	_, err := cmf.processIntegrations()
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "found multiple instances of singleton"))
}

func TestConfigMakerWithExporterWithTemplate(t *testing.T) {
	configStr := `
windows_exporter:
  enabled_collectors: {{ (datasource "vars").value }}
  instance: testinstance
`
	tDir := generatePath(t)
	writeFile(t, tDir, "vars.yaml", "value: banana")
	writeFile(t, tDir, "integrations-1.yml", configStr)
	fileFS := generateFilePath(tDir)

	loaderCfg := LoaderConfig{
		Sources: []Datasource{{
			Name: "vars",
			URL:  generateFilePath(filepath.Join(tDir, "vars.yaml")),
		}},
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	configs, err := cmf.processIntegrations()
	assert.NoError(t, err)
	assert.Len(t, configs, 1)
	wincfg, _ := configs[0].(v2.UpgradedConfig).LegacyConfig()
	assert.True(t, wincfg.(*windows_exporter.Config).EnabledCollectors == "banana")
}

func TestConfigMakerWithMultipleExporter(t *testing.T) {
	configStr := `
windows_exporter:
  enabled_collectors: one,two,three
  instance: testinstance
node_exporter:
  autoscrape:
    enable: false
`
	tDir := generatePath(t)
	writeFile(t, tDir, "integrations-1.yml", configStr)
	fileFS := generateFilePath(tDir)

	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	configs, err := cmf.processIntegrations()
	assert.NoError(t, err)
	assert.Len(t, configs, 2)
	for _, cfg := range configs {
		switch v := cfg.(type) {
		default:
			fmt.Printf("unexpected type %T", v)
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
windows_exporter:
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
	assert.NoError(t, err)
	assert.Len(t, cfg, 1)
	oc, _ := cfg[0].(v2.UpgradedConfig).LegacyConfig()
	winCfg := oc.(*windows_exporter.Config)
	assert.True(t, winCfg.EnabledCollectors == "one,two,three")
}

func TestLoadingFromS3LoadingVarsLocally(t *testing.T) {
	configStr := `
windows_exporter:
  enabled_collectors: {{ (datasource "vars").value }}
  instance: testinstance
`
	tDir := generatePath(t)
	writeFile(t, tDir, "vars.yaml", "value: banana")
	u := pushFilesToFakeS3(t, "integrations-1.yml", configStr)
	s3Url := "s3://mybucket/?region=us-east-1&disableSSL=true&s3ForcePathStyle=true&endpoint=" + u.Host
	loaderCfg := LoaderConfig{
		Sources: []Datasource{{
			Name: "vars",
			URL:  fmt.Sprintf("file:///%s", filepath.Join(tDir, "vars.yaml")),
		}},
		TemplatePaths: []string{s3Url},
	}
	cmf := generateLoader(t, loaderCfg)
	cfg, err := cmf.processIntegrations()
	assert.NoError(t, err)
	assert.Len(t, cfg, 1)
	oc, _ := cfg[0].(v2.UpgradedConfig).LegacyConfig()
	winCfg := oc.(*windows_exporter.Config)
	assert.True(t, winCfg.EnabledCollectors == "banana")
}

func TestLoadingFromS3LoadingVarsLocallyWithRange(t *testing.T) {
	configStr := `
windows_exporter:
  enabled_collectors: banana
  instance: testinstance
  autoscrape:
    metric_relabel_configs: {{ range (datasource "vars").value }}
    - source_labels: [__address__]
      target_label: {{ . }}
      replacement: "{{ . }}-value"
    {{ end }}
`
	tDir := generatePath(t)
	writeFile(t, tDir, "vars.yaml", "value: [banana,apple,pear]")
	u := pushFilesToFakeS3(t, "integrations-1.yml", configStr)

	s3Url := "s3://mybucket/?region=us-east-1&disableSSL=true&s3ForcePathStyle=true&endpoint=" + u.Host
	loaderCfg := LoaderConfig{
		Sources: []Datasource{{
			Name: "vars",
			URL:  fmt.Sprintf("file:///%s", filepath.Join(tDir, "vars.yaml")),
		}},
		TemplatePaths: []string{s3Url},
	}
	cmf := generateLoader(t, loaderCfg)

	cfg, err := cmf.processIntegrations()
	assert.NoError(t, err)
	assert.Len(t, cfg, 1)
	expectBase := `
common:
  autoscrape:
    metric_relabel_configs:
    - target_label: banana
    - target_label: apple
    - target_label: pear
`
	outBytes, err := yaml.Marshal(cfg[0])
	assert.NoError(t, err)
	assert.NoError(t, subset.YAMLAssert([]byte(expectBase), outBytes))
}

func TestMultiplex(t *testing.T) {
	configStr := `
redis_exporter_configs:
- redis_addr: localhost:6379
  autoscrape:
    metric_relabel_configs: 
    - source_labels: [__address__]
      target_label: "banana"
      replacement: "apple"
- redis_addr: localhost:6380
`
	tDir := generatePath(t)
	writeFile(t, tDir, "integrations-1.yml", configStr)
	fileFS := generateFilePath(tDir)

	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	configs, err := cmf.processIntegrations()
	assert.Nil(t, err)
	assert.Len(t, configs, 2)
	outBytes, err := yaml.Marshal(configs)
	assert.NoError(t, err)
	outStr := string(outBytes)
	assert.True(t, strings.Contains(outStr, "localhost:6379"))
	assert.True(t, strings.Contains(outStr, "localhost:6380"))
	assert.True(t, strings.Contains(outStr, "apple"))
}

func TestTraces(t *testing.T) {
	configStr := `
configs:
- name: test_traces
  automatic_logging:
    backend: stdout
    loki_name: default
    spans: true
`
	tDir := generatePath(t)
	writeFile(t, tDir, "traces-1.yml", configStr)
	fileFS := generateFilePath(tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	cfg := &Config{}
	err := cmf.ProcessConfigs(cfg, nil)
	assert.NoError(t, err)
	assert.NotNil(t, cfg.Traces)
	assert.Len(t, cfg.Traces.Configs, 1)
	assert.True(t, cfg.Traces.Configs[0].Name == "test_traces")
}

func TestLogs(t *testing.T) {
	configStr := `
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
	tDir := generatePath(t)
	writeFile(t, tDir, "logs-1.yml", configStr)
	fileFS := generateFilePath(tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)

	cfg := &Config{}
	err := cmf.ProcessConfigs(cfg, nil)
	assert.NoError(t, err)
	assert.NotNil(t, cfg.Logs)
	assert.Len(t, cfg.Logs.Configs, 1)
	assert.True(t, cfg.Logs.Configs[0].Name == "test_logs")
}

func TestServer(t *testing.T) {
	configStr := `
http_listen_port: 8080
log_level: debug
`
	tDir := generatePath(t)
	writeFile(t, tDir, "server-1.yml", configStr)
	fileFS := generateFilePath(tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	cfg := &Config{}
	err := cmf.ProcessConfigs(cfg, nil)
	assert.NoError(t, err)
	assert.NotNil(t, cfg.Server)
	assert.True(t, cfg.Server.HTTPListenPort == 8080)
	assert.True(t, cfg.Server.LogLevel.String() == "debug")
}

func TestAgent(t *testing.T) {
	configStr := `
server:
  http_listen_port: 8080
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
	tDir := generatePath(t)
	writeFile(t, tDir, "agent-1.yml", configStr)
	fileFS := generateFilePath(tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	cfg := &Config{}
	err := cmf.ProcessConfigs(cfg, nil)
	assert.NoError(t, err)
	assert.NotNil(t, cfg.Server)
	assert.True(t, cfg.Server.HTTPListenPort == 8080)
	assert.True(t, cfg.Server.LogLevel.String() == "debug")
	assert.NotNil(t, cfg.Metrics)
	assert.True(t, cfg.Metrics.WALDir == "/tmp/grafana-agent-normal")
	assert.True(t, cfg.Metrics.Global.RemoteWrite[0].URL.String() == "https://www.example.com")
	assert.False(t, cfg.Integrations.IsZero())
	assert.True(t, cfg.Integrations.configV2.Configs[0].Name() == "node_exporter")
}

func TestAgentAddIntegrations(t *testing.T) {
	configStr := `
server:
  http_listen_port: 8080
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
	override := `
windows_exporter: {}
`
	tDir := generatePath(t)
	writeFile(t, tDir, "agent-1.yml", configStr)
	writeFile(t, tDir, "integrations-windows.yml", override)
	fileFS := generateFilePath(tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	cfg := &Config{}
	err := cmf.ProcessConfigs(cfg, nil)
	assert.NoError(t, err)
	assert.NotNil(t, cfg.Server)
	assert.True(t, cfg.Server.HTTPListenPort == 8080)
	assert.True(t, cfg.Server.LogLevel.String() == "debug")
	assert.NotNil(t, cfg.Metrics)
	assert.True(t, cfg.Metrics.WALDir == "/tmp/grafana-agent-normal")
	assert.True(t, cfg.Metrics.Global.RemoteWrite[0].URL.String() == "https://www.example.com")
	assert.False(t, cfg.Integrations.IsZero())
	assert.Len(t, cfg.Integrations.configV2.Configs, 2)
}

func TestFilterOverrides(t *testing.T) {
	agentStr := `
server:
  http_listen_port: 8080
  log_level: debug
metrics:
  wal_directory: /tmp/grafana-agent-normal
  global:
    scrape_interval: 60s
    remote_write:
    - url: https://www.example.com
integrations:
  windows_exporter: {}
`
	serverStr := `
http_listen_port: 1111
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
	tDir := generatePath(t)
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
	err := cmf.ProcessConfigs(cfg, nil)
	assert.NoError(t, err)
	// Test Agent by checking integration configs is 2

	// Test server override
	assert.True(t, cfg.Server.HTTPListenPort == 1111)
	// Test metric
	assert.True(t, cfg.Metrics.WALDir == "/tmp/grafana-agent-normal")
	// Test Metric Instances
	assert.True(t, cfg.Metrics.Configs[0].Name == "t1")
	// Test Integrations
	assert.Len(t, cfg.Integrations.configV2.Configs, 2)
	// Test Traces
	assert.True(t, cfg.Traces.Configs[0].Name == "test_traces")
	// Test Logs
	assert.True(t, cfg.Logs.Configs[0].Name == "test_logs")

}

func writeFile(t *testing.T, directory string, path string, contents string) {
	fullpath := filepath.Join(directory, path)
	err := ioutil.WriteFile(fullpath, []byte(contents), 0666)
	assert.Nil(t, err)
}

func generateLoader(t *testing.T, lc LoaderConfig) *DynamicLoader {
	cmf, err := NewDynamicLoader()
	assert.NoError(t, err)
	err = cmf.LoadConfig(lc)
	assert.NoError(t, err)
	return cmf
}

func generateFilePath(directory string) string {
	if runtime.GOOS == "windows" {
		// The URL scheme needs an additional / on windows
		return fmt.Sprintf("file:///%s", directory)
	}
	return fmt.Sprintf("file://%s", directory)
}

func generatePath(t *testing.T) string {
	tDir, err := os.MkdirTemp("", "*-test")
	assert.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(tDir) })
	return tDir
}

func pushFilesToFakeS3(t *testing.T, filename string, filecontents string) *url.URL {
	_ = os.Setenv("AWS_ANON", "true")
	t.Cleanup(func() { _ = os.Unsetenv("AWS_ANON") })

	backend := s3mem.New()
	faker := gofakes3.New(backend)

	srv := httptest.NewServer(faker.Server())
	_ = backend.CreateBucket("mybucket")
	t.Cleanup(srv.Close)
	_, err := backend.PutObject(
		"mybucket",
		filename,
		map[string]string{"Content-Type": "application/yaml"},
		bytes.NewBufferString(filecontents),
		int64(len(filecontents)),
	)
	assert.NoError(t, err)
	u, err := url.Parse(srv.URL)
	assert.NoError(t, err)
	return u
}
