package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/version"
	"github.com/weaveworks/common/logging"
	"k8s.io/apimachinery/pkg/runtime"
	controller "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clog "sigs.k8s.io/controller-runtime/pkg/log"

	grafana_v1alpha1 "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	promop_v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// Config controls the configuration of the operator.
type Config struct {
	LogLevel   logging.Level
	LogFormat  logging.Format
	Controller controller.Options
}

// RegisterFlags registers command-line flags for controlling the Config to the
// given FlagSet.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.LogLevel.RegisterFlags(f)
	c.LogFormat.RegisterFlags(f)
}

func main() {
	var (
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
		cfg    = loadConfig(logger)

		err error
	)

	logger = setupLogger(logger, cfg)

	// Register all types that we will be dealing with to schemeBuilder.
	cfg.Controller.Scheme = runtime.NewScheme()

	for _, add := range []func(*runtime.Scheme) error{
		grafana_v1alpha1.AddToScheme,
		promop_v1.AddToScheme,
	} {
		if err := add(cfg.Controller.Scheme); err != nil {
			level.Error(logger).Log("msg", "unable to register to scheme", "err", err)
			os.Exit(1)
		}
	}

	// Initialize the operator by bringing up a new manager and all controllers.
	m, err := controller.NewManager(controller.GetConfigOrDie(), cfg.Controller)
	if err != nil {
		level.Error(logger).Log("msg", "unable to start manager", "err", err)
		os.Exit(1)
	}

	// TODO(rfratto): reconciler for more resources
	err = controller.NewControllerManagedBy(m).
		For(&grafana_v1alpha1.GrafanaAgent{}).
		Complete(&reconciler{
			Client: m.GetClient(),
			scheme: m.GetScheme(),
		})
	if err != nil {
		level.Error(logger).Log("msg", "unable to create controller", "err", err)
		os.Exit(1)
	}

	// Run the manager and wait for a signal to shut down.
	level.Info(logger).Log("msg", "starting manager")
	if err := m.Start(controller.SetupSignalHandler()); err != nil {
		level.Error(logger).Log("msg", "problem running manager", "err", err)
		os.Exit(1)
	}
}

// loadConfig will read command line flags and populate a Config. loadConfig
// will exit the program on failure.
func loadConfig(l log.Logger) *Config {
	var (
		printVersion bool
		cfg          Config
	)

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.BoolVar(&printVersion, "version", false, "Print this build's version information")
	cfg.RegisterFlags(fs)

	if err := fs.Parse(os.Args[1:]); err != nil {
		level.Error(l).Log("msg", "failed to parse flags", "err", err)
		os.Exit(1)
	}
	if printVersion {
		fmt.Println(version.Print("agent-operator"))
		os.Exit(0)
	}

	return &cfg
}

type reconciler struct {
	client.Client
	scheme *runtime.Scheme
}

func (r *reconciler) Reconcile(ctx context.Context, req controller.Request) (controller.Result, error) {
	l := clog.FromContext(ctx).WithValues("grafana-agent", req.NamespacedName)
	l.Info("reconciling grafana-agent")

	var agent grafana_v1alpha1.GrafanaAgent
	if err := r.Get(ctx, req.NamespacedName, &agent); err != nil {
		l.Error(err, "unable to get grafana-agent")
		return controller.Result{}, nil
	}

	// TODO(rfratto): do everything now :)

	return controller.Result{}, nil
}
