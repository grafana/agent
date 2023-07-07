package flowmode

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/converter"
	convert_diag "github.com/grafana/agent/converter/diag"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/maps"

	"github.com/fatih/color"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/boringcrypto"
	"github.com/grafana/agent/pkg/cluster"
	"github.com/grafana/agent/pkg/config/instrumentation"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/usagestats"
	httpservice "github.com/grafana/agent/service/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"

	// Install Components
	_ "github.com/grafana/agent/component/all"
)

func runCommand() *cobra.Command {
	r := &flowRun{
		inMemoryAddr:     "agent.internal:12345",
		httpListenAddr:   "127.0.0.1:12345",
		storagePath:      "data-agent/",
		uiPrefix:         "/",
		disableReporting: false,
		enablePprof:      true,
		configFormat:     "flow",
	}

	cmd := &cobra.Command{
		Use:   "run [flags] file",
		Short: "Run Grafana Agent Flow",
		Long: `The run subcommand runs Grafana Agent Flow in the foreground until an interrupt
is received.

run must be provided an argument pointing at the River file to use. If the
River file wasn't specified, can't be loaded, or contains errors, run will exit
immediately.

run starts an HTTP server which can be used to debug Grafana Agent Flow or
force it to reload (by sending a GET or POST request to /-/reload). The listen
address can be changed through the --server.http.listen-addr flag.

By default, the HTTP server exposes a debugging UI at /. The path of the
debugging UI can be changed by providing a different value to
--server.http.ui-path-prefix.

Additionally, the HTTP server exposes the following debug endpoints:

  /debug/pprof   Go performance profiling tools

If reloading the config file fails, Grafana Agent Flow will continue running in
its last valid state. Components which failed may be be listed as unhealthy,
depending on the nature of the reload error.
`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			return r.Run(args[0])
		},
	}

	cmd.Flags().
		StringVar(&r.httpListenAddr, "server.http.listen-addr", r.httpListenAddr, "Address to listen for HTTP traffic on")
	cmd.Flags().StringVar(&r.inMemoryAddr, "server.http.memory-addr", r.inMemoryAddr, "Address to listen for in-memory HTTP traffic on. Change if it collides with a real address")
	cmd.Flags().StringVar(&r.storagePath, "storage.path", r.storagePath, "Base directory where components can store data")
	cmd.Flags().StringVar(&r.uiPrefix, "server.http.ui-path-prefix", r.uiPrefix, "Prefix to serve the HTTP UI at")
	cmd.Flags().
		BoolVar(&r.enablePprof, "server.http.enable-pprof", r.enablePprof, "Enable /debug/pprof profiling endpoints.")
	cmd.Flags().
		BoolVar(&r.clusterEnabled, "cluster.enabled", r.clusterEnabled, "Start in clustered mode")
	cmd.Flags().
		StringVar(&r.clusterNodeName, "cluster.node-name", r.clusterNodeName, "The name to use for this node")
	cmd.Flags().
		StringVar(&r.clusterAdvAddr, "cluster.advertise-address", r.clusterAdvAddr, "Address to advertise to the cluster")
	cmd.Flags().
		StringVar(&r.clusterJoinAddr, "cluster.join-addresses", r.clusterJoinAddr, "Comma-separated list of addresses to join the cluster at")
	cmd.Flags().
		BoolVar(&r.disableReporting, "disable-reporting", r.disableReporting, "Disable reporting of enabled components to Grafana.")
	cmd.Flags().StringVar(&r.configFormat, "config.format", r.configFormat, "The format of the source file. Supported formats: 'flow', 'prometheus'.")
	cmd.Flags().BoolVar(&r.configBypassConversionErrors, "config.bypass-conversion-errors", r.configBypassConversionErrors, "Enable bypassing errors when converting")
	return cmd
}

type flowRun struct {
	inMemoryAddr                 string
	httpListenAddr               string
	storagePath                  string
	uiPrefix                     string
	enablePprof                  bool
	disableReporting             bool
	clusterEnabled               bool
	clusterNodeName              string
	clusterAdvAddr               string
	clusterJoinAddr              string
	configFormat                 string
	configBypassConversionErrors bool
}

func (fr *flowRun) Run(configFile string) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := interruptContext()
	defer cancel()

	if configFile == "" {
		return fmt.Errorf("file argument not provided")
	}

	l, err := logging.New(os.Stderr, logging.DefaultOptions)
	if err != nil {
		return fmt.Errorf("building logger: %w", err)
	}

	t, err := tracing.New(tracing.DefaultOptions)
	if err != nil {
		return fmt.Errorf("building tracer: %w", err)
	}

	// Set the global tracer provider to catch global traces, but ideally things
	// use the tracer provider given to them so the appropriate attributes get
	// injected.
	otel.SetTracerProvider(t)

	level.Info(l).Log("boringcrypto enabled", boringcrypto.Enabled)

	// Immediately start the tracer.
	go func() {
		err := t.Run(ctx)
		if err != nil {
			level.Error(l).Log("msg", "running tracer returned an error", "err", err)
		}
	}()

	// TODO(rfratto): many of the dependencies we import register global metrics,
	// even when their code isn't being used. To reduce the number of series
	// generated by the agent, we should switch to a custom registry.
	//
	// Before doing this, we need to ensure that anything using the default
	// registry that we want to keep can be given a custom registry so desired
	// metrics are still exposed.
	reg := prometheus.DefaultRegisterer
	reg.MustRegister(newResourcesCollector(l))

	clusterer, err := cluster.New(l, reg, fr.clusterEnabled, fr.clusterNodeName, fr.httpListenAddr, fr.clusterAdvAddr, fr.clusterJoinAddr)
	if err != nil {
		return fmt.Errorf("building clusterer: %w", err)
	}
	defer func() {
		err := clusterer.Stop(ctx)
		if err != nil {
			level.Error(l).Log("msg", "failed to terminate clusterer", "err", err)
		}
	}()

	var httpData httpservice.Data

	f := flow.New(flow.Options{
		Logger:         l,
		Tracer:         t,
		Clusterer:      clusterer,
		DataPath:       fr.storagePath,
		Reg:            reg,
		HTTPPathPrefix: "/api/v0/component/",
		HTTPListenAddr: fr.inMemoryAddr,

		// Send requests to fr.inMemoryAddr directly to our in-memory listener.
		DialFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
			// NOTE(rfratto): this references a variable because httpData isn't
			// non-zero by the time we are constructing the Flow controller.
			//
			// This is a temporary hack while services are being built out.
			//
			// It is not possibl for this to panic, since we will always have the service
			// constructed by the time anything in Flow invokes this function.
			return httpData.DialFunc(ctx, network, address)
		},
	})

	reload := func() error {
		flowCfg, err := loadFlowFile(configFile, fr.configFormat, fr.configBypassConversionErrors)
		defer instrumentation.InstrumentLoad(err == nil)

		if err != nil {
			return fmt.Errorf("reading config file %q: %w", configFile, err)
		}
		if err := f.LoadFile(flowCfg, nil); err != nil {
			return fmt.Errorf("error during the initial gragent load: %w", err)
		}

		return nil
	}

	httpService := httpservice.New(httpservice.Options{
		Logger:   log.With(l, "service", "http"),
		Tracer:   t,
		Gatherer: prometheus.DefaultGatherer,

		Clusterer:  clusterer,
		ReadyFunc:  func() bool { return f.Ready() },
		ReloadFunc: reload,

		HTTPListenAddr:   fr.httpListenAddr,
		MemoryListenAddr: fr.inMemoryAddr,
		UIPrefix:         fr.uiPrefix,
		EnablePProf:      fr.enablePprof,
	})
	httpData = httpService.Data().(httpservice.Data)

	// Flow controller
	{
		wg.Add(1)
		go func() {
			defer wg.Done()
			f.Run(ctx)
		}()
	}

	// HTTP service
	{
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = httpService.Run(ctx, f)
		}()
	}

	// Report usage of enabled components
	if !fr.disableReporting {
		reporter, err := usagestats.NewReporter(l)
		if err != nil {
			return fmt.Errorf("failed to create reporter: %w", err)
		}
		go func() {
			err := reporter.Start(ctx, getEnabledComponentsFunc(f))
			if err != nil {
				level.Error(l).Log("msg", "failed to start reporter", "err", err)
			}
		}()
	}

	// Start the Clusterer's Node implementation.
	err = clusterer.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start the clusterer: %w", err)
	}

	// Perform the initial reload. This is done after starting the HTTP server so
	// that /metric and pprof endpoints are available while the Flow controller
	// is loading.
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

	reloadSignal := make(chan os.Signal, 1)
	signal.Notify(reloadSignal, syscall.SIGHUP)
	defer signal.Stop(reloadSignal)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-reloadSignal:
			if err := reload(); err != nil {
				level.Error(l).Log("msg", "failed to reload config", "err", err)
			} else {
				level.Info(l).Log("msg", "config reloaded")
			}
		}
	}
}

// getEnabledComponentsFunc returns a function that gets the current enabled components
func getEnabledComponentsFunc(f *flow.Flow) func() map[string]interface{} {
	return func() map[string]interface{} {
		components := component.GetAllComponents(f, component.InfoOptions{})
		componentNames := map[string]struct{}{}
		for _, c := range components {
			componentNames[c.Registration.Name] = struct{}{}
		}
		return map[string]interface{}{"enabled-components": maps.Keys(componentNames)}
	}
}

func loadFlowFile(filename string, converterSourceFormat string, converterBypassErrors bool) (*flow.File, error) {
	bb, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if converterSourceFormat != "flow" {
		var diags convert_diag.Diagnostics
		bb, diags = converter.Convert(bb, converter.Input(converterSourceFormat))
		hasError := hasErrorLevel(diags, convert_diag.SeverityLevelError)
		hasCritical := hasErrorLevel(diags, convert_diag.SeverityLevelCritical)
		if hasCritical || (!converterBypassErrors && hasError) {
			return nil, diags
		}
	}

	instrumentation.InstrumentConfig(bb)

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
