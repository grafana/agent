package flowmode

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/converter"
	convert_diag "github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/pkg/boringcrypto"
	"github.com/grafana/agent/pkg/config/instrumentation"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/grafana/agent/pkg/usagestats"
	"github.com/grafana/agent/service"
	"github.com/grafana/agent/service/cluster"
	httpservice "github.com/grafana/agent/service/http"
	otel_service "github.com/grafana/agent/service/otel"
	uiservice "github.com/grafana/agent/service/ui"
	"github.com/grafana/ckit/advertise"
	"github.com/grafana/ckit/peer"
	"github.com/grafana/river/diag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/maps"

	// Install Components
	_ "github.com/grafana/agent/component/all"
)

func runCommand() *cobra.Command {
	r := &flowRun{
		inMemoryAddr:          "agent.internal:12345",
		httpListenAddr:        "127.0.0.1:12345",
		storagePath:           "data-agent/",
		uiPrefix:              "/",
		disableReporting:      false,
		enablePprof:           true,
		configFormat:          "flow",
		clusterAdvInterfaces:  advertise.DefaultInterfaces,
		ClusterMaxJoinPeers:   5,
		clusterRejoinInterval: 60 * time.Second,
	}

	cmd := &cobra.Command{
		Use:   "run [flags] path",
		Short: "Run Grafana Agent Flow",
		Long: `The run subcommand runs Grafana Agent Flow in the foreground until an interrupt
is received.

run must be provided an argument pointing at the River dir/file-path to use. If the
River dir/file-path wasn't specified, can't be loaded, or contains errors, run will exit
immediately.

If path is a directory, all *.river files in that directory will be combined
into a single unit. Subdirectories are not recursively searched for further merging.

run starts an HTTP server which can be used to debug Grafana Agent Flow or
force it to reload (by sending a GET or POST request to /-/reload). The listen
address can be changed through the --server.http.listen-addr flag.

By default, the HTTP server exposes a debugging UI at /. The path of the
debugging UI can be changed by providing a different value to
--server.http.ui-path-prefix.

Additionally, the HTTP server exposes the following debug endpoints:

  /debug/pprof   Go performance profiling tools

If reloading the config dir/file-path fails, Grafana Agent Flow will continue running in
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
		StringVar(&r.clusterDiscoverPeers, "cluster.discover-peers", r.clusterDiscoverPeers, "List of key-value tuples for discovering peers")
	cmd.Flags().
		StringSliceVar(&r.clusterAdvInterfaces, "cluster.advertise-interfaces", r.clusterAdvInterfaces, "List of interfaces used to infer an address to advertise")
	cmd.Flags().
		DurationVar(&r.clusterRejoinInterval, "cluster.rejoin-interval", r.clusterRejoinInterval, "How often to rejoin the list of peers")
	cmd.Flags().
		IntVar(&r.ClusterMaxJoinPeers, "cluster.max-join-peers", r.ClusterMaxJoinPeers, "Number of peers to join from the discovered set")
	cmd.Flags().
		StringVar(&r.clusterName, "cluster.name", r.clusterName, "The name of the cluster to join")
	cmd.Flags().
		BoolVar(&r.disableReporting, "disable-reporting", r.disableReporting, "Disable reporting of enabled components to Grafana.")
	cmd.Flags().StringVar(&r.configFormat, "config.format", r.configFormat, fmt.Sprintf("The format of the source file. Supported formats: %s.", supportedFormatsList()))
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
	clusterDiscoverPeers         string
	clusterAdvInterfaces         []string
	clusterRejoinInterval        time.Duration
	ClusterMaxJoinPeers          int
	clusterName                  string
	configFormat                 string
	configBypassConversionErrors bool
}

func (fr *flowRun) Run(configPath string) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := interruptContext()
	defer cancel()

	if configPath == "" {
		return fmt.Errorf("path argument not provided")
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

	// There's a cyclic dependency between the definition of the Flow controller,
	// the reload/ready functions, and the HTTP service.
	//
	// To work around this, we lazily create variables for the functions the HTTP
	// service needs and set them after the Flow controller exists.
	var (
		reload func() (*flow.Source, error)
		ready  func() bool
	)

	clusterService, err := buildClusterService(clusterOptions{
		Log:     l,
		Tracer:  t,
		Metrics: reg,

		EnableClustering:    fr.clusterEnabled,
		NodeName:            fr.clusterNodeName,
		AdvertiseAddress:    fr.clusterAdvAddr,
		ListenAddress:       fr.httpListenAddr,
		JoinPeers:           strings.Split(fr.clusterJoinAddr, ","),
		DiscoverPeers:       fr.clusterDiscoverPeers,
		RejoinInterval:      fr.clusterRejoinInterval,
		AdvertiseInterfaces: fr.clusterAdvInterfaces,
		ClusterMaxJoinPeers: fr.ClusterMaxJoinPeers,
		ClusterName:         fr.clusterName,
	})
	if err != nil {
		return err
	}

	httpService := httpservice.New(httpservice.Options{
		Logger:   log.With(l, "service", "http"),
		Tracer:   t,
		Gatherer: prometheus.DefaultGatherer,

		ReadyFunc:  func() bool { return ready() },
		ReloadFunc: func() (*flow.Source, error) { return reload() },

		HTTPListenAddr:   fr.httpListenAddr,
		MemoryListenAddr: fr.inMemoryAddr,
		EnablePProf:      fr.enablePprof,
	})

	uiService := uiservice.New(uiservice.Options{
		UIPrefix: fr.uiPrefix,
		Cluster:  clusterService.Data().(cluster.Cluster),
	})

	otelService := otel_service.New(l)
	if otelService == nil {
		return fmt.Errorf("failed to create otel service")
	}

	f := flow.New(flow.Options{
		Logger:   l,
		Tracer:   t,
		DataPath: fr.storagePath,
		Reg:      reg,
		Services: []service.Service{
			httpService,
			uiService,
			clusterService,
			otelService,
		},
	})

	ready = f.Ready
	reload = func() (*flow.Source, error) {
		flowSource, err := loadFlowSource(configPath, fr.configFormat, fr.configBypassConversionErrors)
		defer instrumentation.InstrumentSHA256(flowSource.SHA256())
		defer instrumentation.InstrumentLoad(err == nil)

		if err != nil {
			return nil, fmt.Errorf("reading config path %q: %w", configPath, err)
		}
		if err := f.LoadSource(flowSource, nil); err != nil {
			return flowSource, fmt.Errorf("error during the initial grafana/agent load: %w", err)
		}

		return flowSource, nil
	}

	// Flow controller
	{
		wg.Add(1)
		go func() {
			defer wg.Done()
			f.Run(ctx)
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

	// Perform the initial reload. This is done after starting the HTTP server so
	// that /metric and pprof endpoints are available while the Flow controller
	// is loading.
	if source, err := reload(); err != nil {
		var diags diag.Diagnostics
		if errors.As(err, &diags) {
			p := diag.NewPrinter(diag.PrinterConfig{
				Color:              !color.NoColor,
				ContextLinesBefore: 1,
				ContextLinesAfter:  1,
			})
			_ = p.Fprint(os.Stderr, source.RawConfigs(), diags)

			// Print newline after the diagnostics.
			fmt.Println()

			return fmt.Errorf("could not perform the initial load successfully")
		}

		// Exit if the initial load fails.
		return err
	}

	// By now, have either joined or started a new cluster.
	// Nodes initially join in the Viewer state. After the graph has been
	// loaded successfully, we can move to the Participant state to signal that
	// we wish to participate in reading or writing data.
	err = clusterService.ChangeState(ctx, peer.StateParticipant)
	if err != nil {
		return fmt.Errorf("failed to set clusterer state to Participant after initial load")
	}

	reloadSignal := make(chan os.Signal, 1)
	signal.Notify(reloadSignal, syscall.SIGHUP)
	defer signal.Stop(reloadSignal)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-reloadSignal:
			if _, err := reload(); err != nil {
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

func loadFlowSource(path string, converterSourceFormat string, converterBypassErrors bool) (*flow.Source, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		sources := map[string][]byte{}
		err := filepath.WalkDir(path, func(curPath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			// Skip all directories and don't recurse into child dirs that aren't at top-level
			if d.IsDir() {
				if curPath != path {
					return filepath.SkipDir
				}
				return nil
			}
			// Ignore files not ending in .river extension
			if !strings.HasSuffix(curPath, ".river") {
				return nil
			}

			bb, err := os.ReadFile(curPath)
			sources[curPath] = bb
			return err
		})
		if err != nil {
			return nil, err
		}

		return flow.ParseSources(sources)
	}

	bb, err := os.ReadFile(path)
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

	return flow.ParseSource(path, bb)
}

func interruptContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		select {
		case <-sig:
		case <-ctx.Done():
		}
		signal.Stop(sig)

		fmt.Fprintln(os.Stderr, "interrupt received")
	}()

	return ctx, cancel
}
