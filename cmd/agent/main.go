package main

import (
	"fmt"
	"github.com/grafana/agent/pkg/flow"
	"os"
	"time"

	// Adds version information
	_ "github.com/grafana/agent/pkg/build"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"

	// Register Prometheus SD components
	_ "github.com/grafana/loki/clients/pkg/promtail/discovery/consulagent"
	_ "github.com/prometheus/prometheus/discovery/install"

	// Register integrations
	_ "github.com/grafana/agent/pkg/integrations/install"

	glog "github.com/go-kit/kit/log"

	_ "github.com/grafana/agent/component/all"
)

func init() {
	prometheus.MustRegister(version.NewCollector("agent"))
}

func main() {
	content := `
		cluster {
			self = "localhost"
		}
		
		targetout "t1" {
			name = "t1"
			input = cluster.output
		}

		targetout "t2" {
			name = "t2"
			input = cluster.output
		}
	`
	logger := glog.NewLogfmtLogger(os.Stdout)
	f, err := flow.ReadFile("cluster", []byte(content))
	if err != nil {
		fmt.Println("exitting")
		os.Exit(1)
	}
	ctrl := flow.New(flow.Options{
		Logger:   logger,
		DataPath: os.TempDir(),
	})
	ctrl.LoadFile(f)
	time.Sleep(10 * time.Hour)
	return
	/*
		// If Windows is trying to run us as a service, go through that
		// path instead.
		if IsWindowsService() {
			err := RunService()
			if err != nil {
				log.Fatalln(err)
			}
			return
		}

		reloader := func() (*config.Config, error) {
			fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			return config.Load(fs, os.Args[1:])
		}
		cfg, err := reloader()
		if err != nil {
			log.Fatalln(err)
		}

		// After this point we can start using go-kit logging.
		logger := server.NewLogger(&cfg.Server)
		util_log.Logger = logger

		ep, err := NewEntrypoint(logger, cfg, reloader)
		if err != nil {
			level.Error(logger).Log("msg", "error creating the agent server entrypoint", "err", err)
			os.Exit(1)
		}

		if err = ep.Start(); err != nil {
			level.Error(logger).Log("msg", "error running agent", "err", err)
			// Don't os.Exit here; we want to do cleanup by stopping promMetrics
		}

		ep.Stop()
		level.Info(logger).Log("msg", "agent exiting")*/
}
