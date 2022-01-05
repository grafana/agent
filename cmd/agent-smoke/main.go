package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/go-kit/log/level"
	smoke "github.com/grafana/agent/cmd/agent-smoke/internal"
	"github.com/grafana/agent/pkg/util"
	"github.com/weaveworks/common/logging"
)

func main() {
	var (
		namespace         string
		kubeconfig        string
		logLevel          logging.Level
		logFormat         logging.Format
		withTimeout       time.Duration
		chaosFrequency    time.Duration
		mutationFrequency time.Duration
	)

	logLevel.RegisterFlags(flag.CommandLine)
	logFormat.RegisterFlags(flag.CommandLine)
	flag.StringVar(&namespace, "namespace", "agent-smoke-test", "namespace smoke test should run in")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.DurationVar(&withTimeout, "duration", time.Duration(0), "timeout after duration, example 3h")
	flag.DurationVar(&chaosFrequency, "chaos-frequency", 30*time.Minute, "chaos frequency duration")
	flag.DurationVar(&mutationFrequency, "mutation-frequency", 5*time.Minute, "mutation frequency duration")
	flag.Parse()

	logger := util.NewLoggerFromLevel(logLevel, logFormat)

	ctx := context.Background()
	if withTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, withTimeout)
		defer cancel()
		level.Debug(logger).Log("msg", "running with duration", "duration", withTimeout.String())
	}

	opts := []smoke.Option{
		smoke.WithLogger(logger),
		smoke.WithNamespace(namespace),
		smoke.WithChaosFrequency(chaosFrequency),
		smoke.WithMutationFrequency(mutationFrequency),
	}
	if kubeconfig != "" {
		opts = append(opts, smoke.WithKubeConfig(kubeconfig))
	}

	smokeTest, err := smoke.NewSmokeTest(opts...)
	if err != nil {
		level.Error(logger).Log("msg", "error constructing smoke test", "err", err)
		os.Exit(-1)
	}
	if err := smokeTest.Run(ctx); err != nil {
		level.Error(logger).Log("msg", "smoke test run failure", "err", err)
		os.Exit(-1)
	}
}
