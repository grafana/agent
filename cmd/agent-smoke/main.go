package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	smoke "github.com/grafana/agent/cmd/agent-smoke/internal"
)

func main() {
	var (
		debugLog          bool
		namespace         string
		kubeconfig        string
		withTimeout       time.Duration
		chaosFrequency    time.Duration
		mutationFrequency time.Duration
	)

	flag.BoolVar(&debugLog, "debug", false, "enable debug logging")
	flag.StringVar(&namespace, "namespace", "agent-smoke-test", "namespace smoke test should run in")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.DurationVar(&withTimeout, "duration", time.Duration(0), "timeout after duration, example 3h")
	flag.DurationVar(&chaosFrequency, "chaos-frequency", 30*time.Minute, "chaos frequency duration")
	flag.DurationVar(&mutationFrequency, "mutation-frequency", 5*time.Minute, "mutation frequency duration")
	flag.Parse()

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	level.AllowInfo()
	level.Info(logger).Log("msg", "starting agent smoke framework")
	if debugLog {
		level.AllowDebug()
		level.Debug(logger).Log("msg", "debug logging enabled")
	}

	ctx, cancel := context.WithCancel(context.Background())
	if withTimeout > 0 {
		level.Debug(logger).Log("msg", "starting with timeout", "duration", withTimeout.String())
		ctx, cancel = context.WithTimeout(context.Background(), withTimeout)
	}
	defer cancel()

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
