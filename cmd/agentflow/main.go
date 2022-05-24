package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof" // anonymous import to get the pprof handler registered
	"os"
	"os/signal"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/flow"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// Install components
	_ "github.com/grafana/agent/component/all"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
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

	l := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	f := flow.New(flow.Options{
		Logger:   l,
		DataPath: storagePath,
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
		r.Handle("/-/config", f.ConfigHandler())
		r.Handle("/metrics", promhttp.Handler())
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

	f, diags := flow.ReadFile(filename, bb)
	if !diags.HasErrors() {
		return f, nil
	}
	return f, diags
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
