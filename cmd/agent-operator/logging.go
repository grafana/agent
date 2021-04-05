package main

import (
	"os"

	cortex_log "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/cmd/agent-operator/internal/logutil"
	controller "sigs.k8s.io/controller-runtime"
)

// setupLogger sets up our logger. If this function fails, the program will
// exit.
func setupLogger(l log.Logger, cfg *Config) log.Logger {
	newLogger, err := cortex_log.NewPrometheusLogger(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		level.Error(l).Log("msg", "failed to create logger", "err", err)
		os.Exit(1)
	}
	l = newLogger

	adapterLogger := logutil.Wrap(l)

	// NOTE: we don't set up a caller field here, unlike the normal agent.
	// There's too many multiple nestings of the logger that prevent getting the
	// caller from working properly.

	// Set up the global logger and the controller-local logger.
	controller.SetLogger(adapterLogger)
	cfg.Controller.Logger = adapterLogger
	return l
}
