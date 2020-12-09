package config

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	promCfg "github.com/prometheus/prometheus/config"
	"github.com/stretchr/testify/require"
)

// TestConfig_FlagDefaults makes sure that default values of flags are kept
// when parsing the config.
func TestConfig_FlagDefaults(t *testing.T) {
	cfg := `
prometheus:
  wal_directory: /tmp/wal
  global:
    scrape_timeout: 33s`

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	c, err := load(fs, []string{"-config.file", "test"}, func(_ string, c *Config) error {
		return LoadBytes([]byte(cfg), c)
	})
	require.NoError(t, err)
	require.NotEmpty(t, c.Prometheus.ServiceConfig.Lifecycler.InfNames)
	require.NotZero(t, c.Prometheus.ServiceConfig.Lifecycler.NumTokens)
	require.NotZero(t, c.Prometheus.ServiceConfig.Lifecycler.HeartbeatPeriod)
}

func TestConfig_OverrideDefaultsOnLoad(t *testing.T) {
	cfg := `
prometheus:
  wal_directory: /tmp/wal
  global:
    scrape_timeout: 33s`
	expect := promCfg.GlobalConfig{
		ScrapeInterval:     model.Duration(1 * time.Minute),
		ScrapeTimeout:      model.Duration(33 * time.Second),
		EvaluationInterval: model.Duration(1 * time.Minute),
	}

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	c, err := load(fs, []string{"-config.file", "test"}, func(_ string, c *Config) error {
		return LoadBytes([]byte(cfg), c)
	})
	require.NoError(t, err)
	require.Equal(t, expect, c.Prometheus.Global)
}

func TestConfig_OverrideByEnvironmentOnLoad(t *testing.T) {
	cfg := `
prometheus:
  wal_directory: /tmp/wal
  global:
    scrape_timeout: ${SCRAPE_TIMEOUT}`
	expect := promCfg.GlobalConfig{
		ScrapeInterval:     model.Duration(1 * time.Minute),
		ScrapeTimeout:      model.Duration(33 * time.Second),
		EvaluationInterval: model.Duration(1 * time.Minute),
	}
	_ = os.Setenv("SCRAPE_TIMEOUT", "33s")

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	c, err := load(fs, []string{"-config.file", "test"}, func(_ string, c *Config) error {
		expandedConfig := os.ExpandEnv(cfg)
		return LoadBytes([]byte(expandedConfig), c)
	})
	require.NoError(t, err)
	require.Equal(t, expect, c.Prometheus.Global)
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
	}

	c, err := load(fs, args, func(_ string, c *Config) error {
		return LoadBytes([]byte(cfg), c)
	})
	require.NoError(t, err)
	require.Equal(t, "/tmp/wal", c.Prometheus.WALDir)
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
		err := LoadBytes([]byte(cfg), &c)
		require.Error(t, err)
	})

	t.Run("non existing key", func(t *testing.T) {
		cfg := `
prometheus:
  wal_directory: /tmp/wal
  global:
  scrape_timeout: 10s`
		var c Config
		err := LoadBytes([]byte(cfg), &c)
		require.Error(t, err)
	})
}
