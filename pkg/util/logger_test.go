package util

import (
	"bytes"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/common/server"
	"gopkg.in/yaml.v2"
)

func TestLogger_ApplyConfig(t *testing.T) {
	var buf bytes.Buffer
	makeLogger := func(cfg *server.Config) (log.Logger, error) {
		l := log.NewLogfmtLogger(log.NewSyncWriter(&buf))
		if cfg.LogFormat.String() == "json" {
			l = log.NewJSONLogger(log.NewSyncWriter(&buf))
		}
		l = level.NewFilter(l, cfg.LogLevel.Gokit)
		return l, nil
	}

	var cfg server.Config
	cfgText := `log_level: error`

	err := yaml.Unmarshal([]byte(cfgText), &cfg)
	require.NoError(t, err)

	l := newLogger(&cfg, makeLogger)
	level.Debug(l).Log("msg", "this should not appear")

	cfgText = `
log_level: debug
log_format: json`
	err = yaml.Unmarshal([]byte(cfgText), &cfg)
	require.NoError(t, err)

	err = l.ApplyConfig(&cfg)
	require.NoError(t, err)

	level.Debug(l).Log("msg", "this should appear")
	require.JSONEq(t, `{
		"level":"debug",
		"msg":"this should appear",
		"caller": "logger.go:70"
	}`, buf.String())
}
