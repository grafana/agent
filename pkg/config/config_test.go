package config

import (
	"testing"
	"time"

	"github.com/prometheus/common/model"
	promCfg "github.com/prometheus/prometheus/config"
	"github.com/stretchr/testify/require"
)

func TestConfig_OverrideDefaultsOnLoad(t *testing.T) {
	cfg := `
prometheus:
  global:
    scrape_timeout: 33s`
	expect := promCfg.GlobalConfig{
		ScrapeInterval:     model.Duration(1 * time.Minute),
		ScrapeTimeout:      model.Duration(33 * time.Second),
		EvaluationInterval: model.Duration(1 * time.Minute),
	}

	var c Config
	err := Load([]byte(cfg), &c)
	require.NoError(t, err)
	require.Equal(t, expect, c.Prometheus.Global)
}

func TestConfig_StrictYamlParsing(t *testing.T) {
	t.Run("duplicate key", func(t *testing.T) {
		cfg := `
prometheus:
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
  global:
  scrape_timeout: 10s`
		var c Config
		err := Load([]byte(cfg), &c)
		require.Error(t, err)
	})

}
