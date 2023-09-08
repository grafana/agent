package config

import (
	"flag"
	"net/url"
	"strings"
	"testing"
	"time"

	commonCfg "github.com/prometheus/common/config"

	"github.com/stretchr/testify/assert"

	"github.com/grafana/agent/pkg/metrics"
	"github.com/grafana/agent/pkg/metrics/instance"
	"github.com/grafana/agent/pkg/server"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/common/model"
	promCfg "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// TestConfig_FlagDefaults makes sure that default values of flags are kept
// when parsing the config.
func TestConfig_FlagDefaults(t *testing.T) {
	cfg := `
metrics:
  wal_directory: /tmp/wal
  global:
    scrape_timeout: 33s`

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	c, err := LoadFromFunc(fs, []string{"-config.file", "test"}, func(_, _ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), false, c)
	})
	require.NoError(t, err)
	require.NotEmpty(t, c.Metrics.ServiceConfig.Lifecycler.InfNames)
	require.NotZero(t, c.Metrics.ServiceConfig.Lifecycler.NumTokens)
	require.NotZero(t, c.Metrics.ServiceConfig.Lifecycler.HeartbeatPeriod)
	require.True(t, c.ServerFlags.RegisterInstrumentation)
}

// TestConfig_ConfigAPIFlag makes sure that the read API flag is passed
// when parsing the config.
func TestConfig_ConfigAPIFlag(t *testing.T) {
	t.Run("Disabled", func(t *testing.T) {
		cfg := `{}`
		fs := flag.NewFlagSet("test", flag.ExitOnError)
		c, err := LoadFromFunc(fs, []string{"-config.file", "test"}, func(_, _ string, _ bool, c *Config) error {
			return LoadBytes([]byte(cfg), false, c)
		})
		require.NoError(t, err)
		require.False(t, c.EnableConfigEndpoints)
		require.False(t, c.Metrics.ServiceConfig.APIEnableGetConfiguration)
	})
	t.Run("Enabled", func(t *testing.T) {
		cfg := `{}`
		fs := flag.NewFlagSet("test", flag.ExitOnError)
		c, err := LoadFromFunc(fs, []string{"-config.file", "test", "-config.enable-read-api"}, func(_, _ string, _ bool, c *Config) error {
			return LoadBytes([]byte(cfg), false, c)
		})
		require.NoError(t, err)
		require.True(t, c.EnableConfigEndpoints)
		require.True(t, c.Metrics.ServiceConfig.APIEnableGetConfiguration)
	})
}

func TestConfig_OverrideDefaultsOnLoad(t *testing.T) {
	cfg := `
metrics:
  wal_directory: /tmp/wal
  global:
    scrape_timeout: 33s`
	expect := instance.GlobalConfig{
		Prometheus: promCfg.GlobalConfig{
			ScrapeInterval:     model.Duration(1 * time.Minute),
			ScrapeTimeout:      model.Duration(33 * time.Second),
			EvaluationInterval: model.Duration(1 * time.Minute),
		},
	}

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	c, err := LoadFromFunc(fs, []string{"-config.file", "test"}, func(_, _ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), false, c)
	})
	require.NoError(t, err)
	require.Equal(t, expect, c.Metrics.Global)
}

func TestConfig_OverrideByEnvironmentOnLoad(t *testing.T) {
	cfg := `
metrics:
  wal_directory: /tmp/wal
  global:
    scrape_timeout: ${SCRAPE_TIMEOUT}`
	expect := instance.GlobalConfig{
		Prometheus: promCfg.GlobalConfig{
			ScrapeInterval:     model.Duration(1 * time.Minute),
			ScrapeTimeout:      model.Duration(33 * time.Second),
			EvaluationInterval: model.Duration(1 * time.Minute),
		},
	}
	t.Setenv("SCRAPE_TIMEOUT", "33s")

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	c, err := LoadFromFunc(fs, []string{"-config.file", "test"}, func(_, _ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), true, c)
	})
	require.NoError(t, err)
	require.Equal(t, expect, c.Metrics.Global)
}

func TestConfig_OverrideByEnvironmentOnLoad_NoDigits(t *testing.T) {
	cfg := `
metrics:
  wal_directory: /tmp/wal
  global:
    external_labels:
      foo: ${1}`
	expect := labels.Labels{{Name: "foo", Value: "${1}"}}

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	c, err := LoadFromFunc(fs, []string{"-config.file", "test"}, func(_, _ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), true, c)
	})
	require.NoError(t, err)
	require.Equal(t, expect, c.Metrics.Global.Prometheus.ExternalLabels)
}

func TestConfig_FlagsAreAccepted(t *testing.T) {
	cfg := `
metrics:
  global:
    scrape_timeout: 33s`

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	args := []string{
		"-config.file", "test",
		"-metrics.wal-directory", "/tmp/wal",
		"-config.expand-env",
	}

	c, err := LoadFromFunc(fs, args, func(_, _ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), false, c)
	})
	require.NoError(t, err)
	require.Equal(t, "/tmp/wal", c.Metrics.WALDir)
}

func TestConfig_StrictYamlParsing(t *testing.T) {
	t.Run("duplicate key", func(t *testing.T) {
		cfg := `
metrics:
  wal_directory: /tmp/wal
  global:
    scrape_timeout: 10s
    scrape_timeout: 15s`
		var c Config
		err := LoadBytes([]byte(cfg), false, &c)
		require.Error(t, err)
	})

	t.Run("non existing key", func(t *testing.T) {
		cfg := `
metrics:
  wal_directory: /tmp/wal
  global:
  scrape_timeout: 10s`
		var c Config
		err := LoadBytes([]byte(cfg), false, &c)
		require.Error(t, err)
	})
}

func TestConfig_Defaults(t *testing.T) {
	var c Config
	err := LoadBytes([]byte(`{}`), false, &c)
	require.NoError(t, err)

	require.Equal(t, metrics.DefaultConfig, c.Metrics)
	require.Equal(t, DefaultVersionedIntegrations(), c.Integrations)
}

func TestConfig_TracesLokiValidates(t *testing.T) {
	tests := []struct {
		cfg string
	}{
		{
			cfg: `
loki:
  configs:
  - name: default
    positions:
      filename: /tmp/positions.yaml
    clients:
    - url: http://loki:3100/loki/api/v1/push
traces:
  configs:
  - name: default
    automatic_logging:
      backend: loki
      loki_name: default
      spans: true`,
		},
		{
			cfg: `
loki:
  configs:
  - name: default
    positions:
      filename: /tmp/positions.yaml
    clients:
    - url: http://loki:3100/loki/api/v1/push
traces:
  configs:
  - name: default
    automatic_logging:
      backend: stdout
      loki_name: doesnt_exist
      spans: true`,
		},
	}

	for _, tc := range tests {
		fs := flag.NewFlagSet("test", flag.ExitOnError)
		_, err := LoadFromFunc(fs, []string{"-config.file", "test"}, func(_, _ string, _ bool, c *Config) error {
			return LoadBytes([]byte(tc.cfg), false, c)
		})

		require.NoError(t, err)
	}
}

func TestConfig_LokiNameMigration(t *testing.T) {
	input := util.Untab(`
loki:
  configs:
  - name: foo
    positions:
      filename: /tmp/positions.yaml
    clients:
    - url: http://loki:3100/loki/api/v1/push
`)
	var cfg Config
	require.NoError(t, LoadBytes([]byte(input), false, &cfg))
	require.NoError(t, cfg.Validate(nil))

	require.NotNil(t, cfg.Logs)
	require.Equal(t, "foo", cfg.Logs.Configs[0].Name)
	require.Equal(t, []string{"`loki` has been deprecated in favor of `logs`"}, cfg.Deprecations)
}

func TestConfig_PrometheusNonNil(t *testing.T) {
	tt := []struct {
		name  string
		input string
	}{
		{
			name:  "missing",
			input: `{}`,
		},
		{
			name:  "null",
			input: `metrics: null`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var cfg Config
			require.NoError(t, LoadBytes([]byte(tc.input), false, &cfg))
			require.NoError(t, cfg.Validate(nil))

			require.NotNil(t, cfg.Metrics)
		})
	}
}

func TestConfig_PrometheusNameMigration(t *testing.T) {
	input := util.Untab(`
prometheus:
	wal_directory: /tmp
  configs:
  - name: default
`)
	var cfg Config
	require.NoError(t, LoadBytes([]byte(input), false, &cfg))
	require.NoError(t, cfg.Validate(nil))

	require.Equal(t, "default", cfg.Metrics.Configs[0].Name)
	require.Equal(t, "/tmp", cfg.Metrics.WALDir)
	require.Equal(t, []string{"`prometheus` has been deprecated in favor of `metrics`"}, cfg.Deprecations)
}

func TestConfig_TracesLokiFailsValidation(t *testing.T) {
	tests := []struct {
		cfg           string
		expectedError string
	}{
		{
			cfg: `
loki:
  configs:
  - name: foo
    positions:
      filename: /tmp/positions.yaml
    clients:
    - url: http://loki:3100/loki/api/v1/push
traces:
  configs:
  - name: default
    automatic_logging:
      backend: logs_instance
      logs_instance_name: default
      spans: true`,
			expectedError: "error in config file: failed to validate automatic_logging for traces config default: specified logs config default not found in agent config",
		},
	}

	for _, tc := range tests {
		fs := flag.NewFlagSet("test", flag.ExitOnError)
		_, err := LoadFromFunc(fs, []string{"-config.file", "test"}, func(_, _ string, _ bool, c *Config) error {
			return LoadBytes([]byte(tc.cfg), false, c)
		})

		require.EqualError(t, err, tc.expectedError)
	}
}

func TestConfig_TempoNameMigration(t *testing.T) {
	input := util.Untab(`
tempo:
  configs:
  - name: default
    automatic_logging:
      backend: stdout
      loki_name: doesnt_exist
      spans: true`)
	var cfg Config
	require.NoError(t, LoadBytes([]byte(input), false, &cfg))
	require.NoError(t, cfg.Validate(nil))

	require.NotNil(t, cfg.Traces)

	require.Equal(t, "default", cfg.Traces.Configs[0].Name)
	require.Equal(t, []string{"`tempo` has been deprecated in favor of `traces`"}, cfg.Deprecations)
}

func TestConfig_TempoTracesDuplicateMigration(t *testing.T) {
	input := util.Untab(`
traces:
  configs:
  - name: default
    automatic_logging:
      backend: stdout
      loki_name: doesnt_exist
      spans: true
tempo:
  configs:
  - name: default
    automatic_logging:
      backend: stdout
      loki_name: doesnt_exist
      spans: true`)
	var cfg Config
	require.EqualError(t, LoadBytes([]byte(input), false, &cfg), "at most one of tempo and traces should be specified")
}

func TestConfig_ExpandEnvRegex(t *testing.T) {
	cfg := `
logs:
  configs:
  - name: default
    positions:
      filename: /tmp/positions.yaml
    scrape_configs:
      - job_name: test
        pipeline_stages:
        - regex:
          source: filename
          expression: '\\temp\\Logs\\(?P<log_app>.+?)\\'`
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	myCfg, err := LoadFromFunc(fs, []string{"-config.file", "test"}, func(_, _ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), true, c)
	})
	require.NoError(t, err)
	pipelineStages := myCfg.Logs.Configs[0].ScrapeConfig[0].PipelineStages[0].(map[interface{}]interface{})
	expected := `\\temp\\Logs\\(?P<log_app>.+?)\\`
	require.Equal(t, expected, pipelineStages["expression"].(string))
}

func TestConfig_ObscureSecrets(t *testing.T) {
	cfgText := `
metrics:
  wal_directory: /tmp
  scraping_service:
    enabled: true
    kvstore:
      store: consul
      consul:
        acl_token: verysecret
    lifecycler:
      ring:
        kvstore:
          store: consul
          consul:
            acl_token: verysecret
`

	var cfg Config
	require.NoError(t, LoadBytes([]byte(cfgText), false, &cfg))

	require.Equal(t, "verysecret", cfg.Metrics.ServiceConfig.KVStore.Consul.ACLToken.String())
	require.Equal(t, "verysecret", cfg.Metrics.ServiceConfig.Lifecycler.RingConfig.KVStore.Consul.ACLToken.String())

	bb, err := yaml.Marshal(&cfg)
	require.NoError(t, err)

	require.False(t, strings.Contains(string(bb), "verysecret"), "secrets did not get obscured")
	require.True(t, strings.Contains(string(bb), "********"), "secrets did not get obscured properly")

	// Re-validate that the config object has not changed
	require.Equal(t, "verysecret", cfg.Metrics.ServiceConfig.KVStore.Consul.ACLToken.String())
	require.Equal(t, "verysecret", cfg.Metrics.ServiceConfig.Lifecycler.RingConfig.KVStore.Consul.ACLToken.String())
}

func TestConfig_RemoteWriteDefaults(t *testing.T) {
	cfg := `
metrics:
  global:
    remote_write:
      - name: "foo"
        url: "https://test/url"`

	var c Config
	err := LoadBytes([]byte(cfg), false, &c)
	require.NoError(t, err)

	expected := &promCfg.DefaultRemoteWriteConfig
	expected.Name = "foo"
	testURL, _ := url.Parse("https://test/url")
	expected.URL = &commonCfg.URL{
		URL: testURL,
	}
	require.Equal(t, expected, c.Metrics.Global.RemoteWrite[0])
	require.True(t, c.Metrics.Global.RemoteWrite[0].SendExemplars)
}

func TestAgent_OmitEmptyFields(t *testing.T) {
	var cfg Config
	yml, err := yaml.Marshal(&cfg)
	require.NoError(t, err)
	require.Equal(t, "{}\n", string(yml))
}

func TestAgentManagement_MergeEffectiveConfig(t *testing.T) {
	initialCfg := `
server:
  log_level: info
logs:
  positions_directory: /tmp
agent_management:
  host: "localhost"
  basic_auth:
    username: "initial_user"
  protocol: "http"
  polling_interval: "1m"
  remote_configuration:
    namespace: "new_namespace"
    cache_location: "/etc"`

	remoteCfg := `
server:
  log_level: debug
metrics:
  wal_directory: /tmp
  global:
    scrape_interval: 5m
integrations:
  scrape_integrations: true

agent_management:
  host: "localhost:80"
  basic_auth:
    username: "new_user"
  protocol: "http"
  polling_interval: "10s"
  remote_configuration:
    namespace: "new_namespace"
    cache_location: "/etc"`

	var ic, rc Config
	err := LoadBytes([]byte(initialCfg), false, &ic)
	assert.NoError(t, err)
	err = LoadBytes([]byte(remoteCfg), false, &rc)
	assert.NoError(t, err)

	// keep a copy of the initial config's agent management block to ensure it isn't
	// overwritten by the remote config's
	initialAgentManagement := ic.AgentManagement
	mergeEffectiveConfig(&ic, &rc)

	// agent_management configuration should not be overwritten by the remote config
	assert.Equal(t, initialAgentManagement, ic.AgentManagement)

	// since these elements are purposefully different for the previous portion of the test,
	// unset them before comparing the rest of the config
	ic.AgentManagement = AgentManagementConfig{}
	rc.AgentManagement = AgentManagementConfig{}

	assert.True(t, util.CompareYAML(ic, rc))
}

func TestConfig_EmptyServerConfigFails(t *testing.T) {
	// Since we are testing defaults via config.Load, we need a file instead of a string.
	// This test file has an empty server stanza, we expect default values out.
	defaultServerCfg := server.DefaultConfig()
	logger := server.NewLogger(&defaultServerCfg)
	fs := flag.NewFlagSet("", flag.ExitOnError)
	_, err := Load(fs, []string{"--config.file", "./testdata/server_empty.yml"}, logger)
	require.Error(t, err)
}

func TestConfig_ValidateConfigWithIntegrationV1(t *testing.T) {
	input := util.Untab(`
integrations:
	agent:
		enabled: true
`)
	var cfg Config
	require.NoError(t, LoadBytes([]byte(input), false, &cfg))
	require.NoError(t, cfg.Validate(nil))
}

func TestConfig_ValidateConfigWithIntegrationV2(t *testing.T) {
	input := util.Untab(`
integrations:
	agent:
		autoscrape:
			enabled: true
`)
	var cfg Config
	require.NoError(t, LoadBytes([]byte(input), false, &cfg))
	require.NoError(t, cfg.Validate(nil))
}
