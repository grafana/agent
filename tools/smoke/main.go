package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/util"
	smoke "github.com/grafana/agent/tools/smoke/internal"
	"github.com/weaveworks/common/logging"
)

func main() {
	var (
		opts        smoke.Options
		logLevel    logging.Level
		logFormat   logging.Format
		withTimeout time.Duration
	)

	logLevel.RegisterFlags(flag.CommandLine)
	logFormat.RegisterFlags(flag.CommandLine)
	flag.StringVar(&opts.Namespace, "namespace", "agent-smoke-test", "namespace smoke test should run in")
	flag.StringVar(&opts.Kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.DurationVar(&withTimeout, "duration", time.Duration(0), "timeout after duration, example 3h")
	flag.DurationVar(&opts.ChaosFrequency, "chaos-frequency", 30*time.Minute, "chaos frequency duration")
	flag.DurationVar(&opts.MutationFrequency, "mutation-frequency", 5*time.Minute, "mutation frequency duration")
	flag.Parse()

	logger := util.NewLoggerFromLevel(logLevel, logFormat)

	ctx := context.Background()
	if withTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, withTimeout)
		defer cancel()
		level.Debug(logger).Log("msg", "running with duration", "duration", withTimeout.String())
	}

	smokeTest, err := smoke.NewSmokeTest(logger, opts)
	if err != nil {
		level.Error(logger).Log("msg", "error constructing smoke test", "err", err)
		os.Exit(1)
	}
	if err := smokeTest.Run(ctx); err != nil {
		level.Error(logger).Log("msg", "smoke test run failure", "err", err)
		os.Exit(1)
	}
}
