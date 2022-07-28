package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/fatih/color"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// Install Components
	_ "github.com/grafana/agent/component/all"
)

func isFlowEnabled() bool {
	key, found := os.LookupEnv("EXPERIMENTAL_ENABLE_FLOW")
	if !found {
		return false
	}
	return key == "true"
}

func runFlow() error {
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := interruptContext()
	defer cancel()

	var (
		httpListenAddr = "127.0.0.1:12345"
		configFile     string
		storagePath    = "data-agent/"
	)

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&httpListenAddr, "server.http-listen-addr", httpListenAddr, "address to listen for http traffic on")
	fs.StringVar(&configFile, "config.file", configFile, "path to config file to load")
	fs.StringVar(&storagePath, "storage.path", storagePath, "Base directory where Flow components can store data")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("error parsing flags: %w", err)
	}

	// Validate flags
	if configFile == "" {
		return fmt.Errorf("the -config.file flag is required")
	}

	l, err := logging.New(os.Stderr, logging.DefaultOptions)
	if err != nil {
		return fmt.Errorf("building logger: %w", err)
	}

	f := flow.New(flow.Options{
		Logger:   l,
		DataPath: storagePath,
		Reg:      prometheus.DefaultRegisterer,
	})

	reload := func() error {
		flowCfg, err := loadFlowFile(configFile)
		if err != nil {
			return fmt.Errorf("reading config file %q: %w", configFile, err)
		}
		if err := f.LoadFile(flowCfg); err != nil {
			return fmt.Errorf("error during the initial gragent load: %w", err)
		}
		return nil
	}

	if err := reload(); err != nil {
		var diags diag.Diagnostics
		if errors.As(err, &diags) {
			bb, _ := os.ReadFile(configFile)

			p := diag.NewPrinter(diag.PrinterConfig{
				Color:              !color.NoColor,
				ContextLinesBefore: 1,
				ContextLinesAfter:  1,
			})
			_ = p.Fprint(os.Stderr, map[string][]byte{configFile: bb}, diags)

			// Print newline after the diagnostics.
			fmt.Println()

			return fmt.Errorf("could not perform the initial load successfully")
		}

		// Exit if the initial load files
		return err
	}

	// HTTP server
	{
		lis, err := net.Listen("tcp", httpListenAddr)
		if err != nil {
			return fmt.Errorf("failed to listen on %s: %w", httpListenAddr, err)
		}

		r := mux.NewRouter()
		r.Handle("/metrics", promhttp.Handler())
		r.Handle("/debug/config", f.ConfigHandler())
		r.Handle("/debug/graph", f.GraphHandler())
		r.PathPrefix("/debug/pprof").Handler(http.DefaultServeMux)

		r.HandleFunc("/-/reload", func(w http.ResponseWriter, _ *http.Request) {
			err := reload()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, "config reloaded")
		})

		srv := &http.Server{Handler: r}

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer cancel()

			level.Info(l).Log("msg", "now listening for http traffic", "addr", httpListenAddr)
			if err := srv.Serve(lis); err != nil {
				level.Info(l).Log("msg", "http server closed", "err", err)
			}
		}()

		defer func() { _ = srv.Shutdown(ctx) }()
	}

	<-ctx.Done()
	return f.Close()
}

func loadFlowFile(filename string) (*flow.File, error) {
	bb, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return flow.ReadFile(filename, bb)
}

func interruptContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		select {
		case <-sig:
		case <-ctx.Done():
		}
		signal.Stop(sig)

		fmt.Fprintln(os.Stderr, "interrupt received")
	}()

	return ctx, cancel
}
