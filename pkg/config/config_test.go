package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_StrictYamlParsing(t *testing.T) {
	t.Run("duplicate key", func(t *testing.T) {
		cfg := `
prometheus:
  wal_directory: /tmp/wal
  global:
    scrape_timeout: 10s
    scrape_timeout: 15s`
		var c Config
		err := Load([]byte(cfg), &c)
		require.Error(t, err)
	})

	t.Run("non existing key", func(t *testing.T) {
		cfg := `
prometheus:
  wal_directory: /tmp/wal
  global:
  scrape_timeout: 10s`
		var c Config
		err := Load([]byte(cfg), &c)
		require.Error(t, err)
	})
}
