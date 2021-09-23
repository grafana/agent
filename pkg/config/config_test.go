package config

import (
	"flag"
	"os"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/metrics"
	"github.com/grafana/agent/pkg/metrics/instance"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/common/model"
	promCfg "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/stretchr/testify/require"

	_ "github.com/grafana/agent/pkg/integrations/github_exporter"   // register github_exporter
	_ "github.com/grafana/agent/pkg/integrations/kafka_exporter"    // register kafka_exporter
	_ "github.com/grafana/agent/pkg/integrations/mongodb_exporter"  // register mongodb_exporter
	_ "github.com/grafana/agent/pkg/integrations/mysqld_exporter"   // register mysqld_exporter
	_ "github.com/grafana/agent/pkg/integrations/postgres_exporter" // register postgres_exporter
	_ "github.com/grafana/agent/pkg/integrations/redis_exporter"    // register redis_exporter
)

func TestConfig_SecretMysqlD(t *testing.T) {
	stringCfg := `
prometheus:
  wal_directory: /tmp/agent
integrations:
  mysqld_exporter:
    enabled: true
    data_source_name: root:secret_password@myserver:3306`
	checkSecret(t, stringCfg, "secret_password")

}

func TestConfig_SecretGithub(t *testing.T) {
	stringCfg := `
prometheus:
  wal_directory: /tmp/agent
integrations:
  github_exporter:
    enabled: true
    api_token: secret_api`
	checkSecret(t, stringCfg, "secret_api")
}

func TestConfig_SecretKafkaUserName(t *testing.T) {
	stringCfg := `
prometheus:
  wal_directory: /tmp/agent
integrations:
  kafka_exporter:
    enabled: true
    sasl_password: secret_password
`
	checkSecret(t, stringCfg, "secret_password")
}

func TestConfig_SecretKafkaPassword(t *testing.T) {
	stringCfg := `
prometheus:
  wal_directory: /tmp/agent
integrations:
  kafka_exporter:
    enabled: true
    sasl_username: secret_username
`
	checkSecret(t, stringCfg, "secret_username")
}

func TestConfig_SecretMongoDB(t *testing.T) {
	stringCfg := `
prometheus:
  wal_directory: /tmp/agent
integrations:
  mongodb_exporter:
    enabled: true
    mongodb_uri: secret_password_in_uri
`
	checkSecret(t, stringCfg, "secret_password_in_uri")
}

func TestConfig_SecretPostgres(t *testing.T) {
	stringCfg := `
prometheus:
  wal_directory: /tmp/agent
integrations:
  postgres_exporter:
    enabled: true
    data_source_names: ["secret_password_in_uri","secret_password_in_uri_2"]
`
	checkSecret(t, stringCfg, "secret_password_in_uri")
	checkSecret(t, stringCfg, "secret_password_in_uri_2")

}

func TestConfig_SecretRedisPassword(t *testing.T) {
	stringCfg := `
prometheus:
  wal_directory: /tmp/agent
integrations:
  redis_exporter:
    enabled: true
    redis_password: secret_password
`
	checkSecret(t, stringCfg, "secret_password")
}

func TestConfig_SecretRedisUsername(t *testing.T) {
	stringCfg := `
prometheus:
  wal_directory: /tmp/agent
integrations:
  redis_exporter:
    enabled: true
    redis_user: secret_user
`
	checkSecret(t, stringCfg, "secret_user")
}

func checkSecret(t *testing.T, rawCfg string, originalValue string) {
	var cfg = &Config{}
	err := LoadBytes([]byte(rawCfg), false, cfg)
	require.NoError(t, err)
	bb, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	scrubbedCfg := string(bb)
	require.True(t, strings.Contains(scrubbedCfg, "<secret>"))
	require.False(t, strings.Contains(scrubbedCfg, originalValue))
}

// TestConfig_FlagDefaults makes sure that default values of flags are kept
// when parsing the config.
func TestConfig_FlagDefaults(t *testing.T) {
	cfg := `
prometheus:
  wal_directory: /tmp/wal
  global:
    scrape_timeout: 33s`

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	c, err := load(fs, []string{"-config.file", "test"}, func(_ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), false, c)
	})
	require.NoError(t, err)
	require.NotEmpty(t, c.Metrics.ServiceConfig.Lifecycler.InfNames)
	require.NotZero(t, c.Metrics.ServiceConfig.Lifecycler.NumTokens)
	require.NotZero(t, c.Metrics.ServiceConfig.Lifecycler.HeartbeatPeriod)
	require.True(t, c.Server.RegisterInstrumentation)
}

func TestConfig_OverrideDefaultsOnLoad(t *testing.T) {
	cfg := `
prometheus:
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
	c, err := load(fs, []string{"-config.file", "test"}, func(_ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), false, c)
	})
	require.NoError(t, err)
	require.Equal(t, expect, c.Metrics.Global)
}

func TestConfig_OverrideByEnvironmentOnLoad(t *testing.T) {
	cfg := `
prometheus:
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
	_ = os.Setenv("SCRAPE_TIMEOUT", "33s")

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	c, err := load(fs, []string{"-config.file", "test"}, func(_ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), true, c)
	})
	require.NoError(t, err)
	require.Equal(t, expect, c.Metrics.Global)
}

func TestConfig_OverrideByEnvironmentOnLoad_NoDigits(t *testing.T) {
	cfg := `
prometheus:
  wal_directory: /tmp/wal
  global:
    external_labels:
      foo: ${1}`
	expect := labels.Labels{{Name: "foo", Value: "${1}"}}

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	c, err := load(fs, []string{"-config.file", "test"}, func(_ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), true, c)
	})
	require.NoError(t, err)
	require.Equal(t, expect, c.Metrics.Global.Prometheus.ExternalLabels)
}

func TestConfig_FlagsAreAccepted(t *testing.T) {
	cfg := `
prometheus:
  global:
    scrape_timeout: 33s`

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	args := []string{
		"-config.file", "test",
		"-prometheus.wal-directory", "/tmp/wal",
		"-config.expand-env",
	}

	c, err := load(fs, args, func(_ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), false, c)
	})
	require.NoError(t, err)
	require.Equal(t, "/tmp/wal", c.Metrics.WALDir)
}

func TestConfig_StrictYamlParsing(t *testing.T) {
	t.Run("duplicate key", func(t *testing.T) {
		cfg := `
prometheus:
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
prometheus:
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
	require.Equal(t, integrations.DefaultManagerConfig, c.Integrations)
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
		_, err := load(fs, []string{"-config.file", "test"}, func(_ string, _ bool, c *Config) error {
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
	require.NoError(t, cfg.ApplyDefaults())

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
			input: `prometheus: null`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var cfg Config
			require.NoError(t, LoadBytes([]byte(tc.input), false, &cfg))
			require.NoError(t, cfg.ApplyDefaults())

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
	require.NoError(t, cfg.ApplyDefaults())

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
		_, err := load(fs, []string{"-config.file", "test"}, func(_ string, _ bool, c *Config) error {
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
	require.NoError(t, cfg.ApplyDefaults())

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
