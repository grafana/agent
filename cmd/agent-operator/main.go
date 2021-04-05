package main

import (
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
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"

	grafana_v1alpha1 "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	promop_v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"

	// Needed for clients.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
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
		core_v1.AddToScheme,
		apps_v1.AddToScheme,
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

	events := newResourceEventHandlers(m.GetClient(), logger)

	applyGVK := func(obj client.Object) client.Object { return applyGVK(obj, m) }
	watchType := func(obj client.Object) source.Source { return watchType(obj, m) }

	err = controller.NewControllerManagedBy(m).
		For(applyGVK(&grafana_v1alpha1.GrafanaAgent{})).
		Owns(applyGVK(&core_v1.Service{})).
		Owns(applyGVK(&core_v1.Secret{})).
		Owns(applyGVK(&apps_v1.StatefulSet{})).
		Watches(watchType(&grafana_v1alpha1.PrometheusInstance{}), events[resourcePromInstance]).
		Watches(watchType(&promop_v1.ServiceMonitor{}), events[resourceServiceMonitor]).
		Watches(watchType(&promop_v1.PodMonitor{}), events[resourcePodMonitor]).
		Watches(watchType(&promop_v1.Probe{}), events[resourceProbe]).
		Watches(watchType(&core_v1.Secret{}), events[resourceSecret]).
		Complete(&reconciler{
			Client:        m.GetClient(),
			scheme:        m.GetScheme(),
			eventHandlers: events,
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

// watchType applies the GVK to an object and returns a source to watch it.
// watchType is a convenience function; without it, the GVK won't show up in
// logs.
func watchType(obj client.Object, m manager.Manager) source.Source {
	applyGVK(obj, m)
	return &source.Kind{Type: obj}
}

// applyGVK applies a GVK to an object based on the scheme. applyGVK is a
// convenience function; without it, the GVK won't show up in logs.
func applyGVK(obj client.Object, m manager.Manager) client.Object {
	gvk, err := apiutil.GVKForObject(obj, m.GetScheme())
	if err != nil {
		panic(err)
	}
	obj.GetObjectKind().SetGroupVersionKind(gvk)
	return obj
}
