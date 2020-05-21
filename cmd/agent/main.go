package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	// Adds version information
	_ "github.com/grafana/agent/pkg/build"

	"github.com/cortexproject/cortex/pkg/util"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/agent/pkg/prom"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"github.com/weaveworks/common/server"
)

func init() {
	prometheus.MustRegister(version.NewCollector("agent"))
}

func main() {
	var (
		printVersion bool

		cfg        config.Config
		configFile string
	)

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&configFile, "config.file", "", "configuration file to load")
	fs.BoolVar(&printVersion, "version", false, "Print this build's version information")
	cfg.RegisterFlags(fs)

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("error parsing flags: %v\n", err)
	}

	if printVersion {
		fmt.Println(version.Print("agent"))
		return
	}

	if configFile == "" {
		log.Fatalln("-config.file flag required")
	} else if err := config.LoadFile(configFile, &cfg); err != nil {
		log.Fatalf("error loading config file %s: %v\n", configFile, err)
	}

	// Parse the flags again to override any yaml values with command line flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("error parsing flags: %v\n", err)
	}

	// After this point we can use util.Logger and stop using the log package
	util.InitLogger(&cfg.Server)

	promMetrics, err := prom.New(cfg.Prometheus, util.Logger)
	if err != nil {
		level.Error(util.Logger).Log("msg", "failed to create prometheus instance", "err", err)
		os.Exit(1)
	}

	srv, err := server.New(cfg.Server)
	if err != nil {
		level.Error(util.Logger).Log("msg", "failed to create server", "err", err)
		os.Exit(1)
	}

	// Hook up API paths to the router
	promMetrics.WireAPI(srv.HTTP)
	promMetrics.WireGRPC(srv.GRPC)

	srv.HTTP.HandleFunc("/-/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Agent is Healthy.\n")
	})
	srv.HTTP.HandleFunc("/-/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Agent is Ready.\n")
	})

	if err := srv.Run(); err != nil {
		level.Error(util.Logger).Log("msg", "error running agent", "err", err)
		// Don't os.Exit here; we want to do cleanup by stopping promMetrics
	}

	promMetrics.Stop()
	level.Info(util.Logger).Log("msg", "agent exiting")
}
