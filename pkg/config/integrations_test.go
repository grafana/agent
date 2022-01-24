package config

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegrations_v1(t *testing.T) {
	cfg := `
metrics:
  wal_directory: /tmp/wal

integrations:
  agent:
    enabled: true`

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	c, err := LoadTest(fs, []string{"-config.file", "test"}, func(_ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), false, c)
	})
	require.NoError(t, err)
	require.False(t, c.Integrations.IsZero())
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
	c, err := LoadTest(fs, []string{"-config.file", "test", "-enable-features=integrations-next"}, func(_ string, _ bool, c *Config) error {
		return LoadBytes([]byte(cfg), false, c)
	})
	require.NoError(t, err)
	require.False(t, c.Integrations.IsZero())
}
