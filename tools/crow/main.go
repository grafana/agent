// Command grafana-agent-crow is a correctness checker tool that validates that
// scraped metrics are delivered to a remote_write endpoint. Inspired by Loki
// Canary and Cortex test-exporter.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	// Adds version information
	_ "github.com/grafana/agent/pkg/build"
	"github.com/grafana/agent/pkg/server"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/crow"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
)

func init() {
	prometheus.MustRegister(version.NewCollector("grafana_agent_crow"))
}

func main() {
	var (
		fs = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		serverCfg   = server.DefaultConfig()
		serverFlags = server.DefaultFlags

		crowCfg     = crow.DefaultConfig
		showVersion bool
	)

	serverFlags.RegisterFlags(fs)
	crowCfg.RegisterFlagsWithPrefix(fs, "crow.")
	fs.BoolVar(&showVersion, "version", false, "show version")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "failed to parse flags", err)
		os.Exit(1)
	}
	if showVersion {
		fmt.Println(version.Print(os.Args[0]))
		os.Exit(0)
	}

	l := server.NewLogger(&serverCfg)
	crowCfg.Log = l

	s, err := server.New(l, prometheus.DefaultRegisterer, prometheus.DefaultGatherer, serverCfg, serverFlags)
	if err != nil {
		level.Error(l).Log("msg", "failed to initialize server", "err", err)
		os.Exit(1)
	}

	c, err := crow.New(crowCfg)
	if err != nil {
		level.Error(l).Log("msg", "failed to initialize crow", "err", err)
		os.Exit(1)
	}
	defer c.Stop()

	// The server comes with a /metrics endpoint by default using s.Registerer.
	// Create a /validate endpoint to handle our validation metrics.
	validator := prometheus.NewRegistry()
	s.HTTP.Handle("/validate", promhttp.HandlerFor(validator, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))

	// Register crow's metrics to /metrics and /validate respectively.
	prometheus.DefaultRegisterer.MustRegister(c.StateMetrics())
	validator.MustRegister(c.TestMetrics())

	ctx, cancel := server.SignalContext(context.Background(), l)
	defer cancel()

	if err := s.Run(ctx); err != nil {
		level.Error(l).Log("msg", "server exited with error", "err", err)
		os.Exit(1)
	}
}
