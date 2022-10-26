package config

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"

	_ "github.com/grafana/agent/pkg/integrations/install" // Install integrations for tests
)

func TestIntegrations_v1(t *testing.T) {
	cfg := `
metrics:
  wal_directory: /tmp/wal

integrations:
  agent:
    enabled: true`

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	c, err := load(fs, []string{"-config.file", "test"}, func(_, _ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), false, c)
	})
	require.NoError(t, err)
	require.NotNil(t, c.Integrations.configV1)
}

func TestIntegrations_v2(t *testing.T) {
	cfg := `
metrics:
  wal_directory: /tmp/wal

integrations:
  agent:
    autoscrape:
      enable: false`

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	c, err := load(fs, []string{"-config.file", "test", "-enable-features=integrations-next"}, func(_, _ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), false, c)
	})
	require.NoError(t, err)
	require.NotNil(t, c.Integrations.configV2)
}

func TestEnabledIntegrations_v1(t *testing.T) {
	cfg := `
metrics:
  wal_directory: /tmp/wal

integrations:
  agent:
    enabled: true
  node_exporter:
    enabled: true`

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	c, err := load(fs, []string{"-config.file", "test"}, func(_, _ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), false, c)
	})
	require.NoError(t, err)
	require.Equal(t, c.Integrations.EnabledIntegrations(), []string{"agent", "node_exporter"})
}

func TestEnabledIntegrations_v2(t *testing.T) {
	cfg := `
metrics:
  wal_directory: /tmp/wal

integrations:
  agent:
    autoscrape:
      enable: false
  node_exporter:
    autoscrape:
      enable: false`

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	c, err := load(fs, []string{"-config.file", "test", "-enable-features=integrations-next"}, func(_, _ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), false, c)
	})
	require.NoError(t, err)
	require.Equal(t, c.Integrations.EnabledIntegrations(), []string{"node_exporter", "agent"})
}

func TestEnabledIntegrations_v2MultipleInstances(t *testing.T) {
	cfg := `
metrics:
  wal_directory: /tmp/wal

integrations: 
  redis_configs:
  - redis_addr: "redis-0:6379"
  - redis_addr: "redis-1:6379"`

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	c, err := load(fs, []string{"-config.file", "test", "-enable-features=integrations-next"}, func(_, _ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), false, c)
	})
	require.NoError(t, err)
	require.Equal(t, c.Integrations.EnabledIntegrations(), []string{"redis"})
}
