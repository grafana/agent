package operator

import (
	"flag"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/weaveworks/common/logging"
	"k8s.io/apimachinery/pkg/runtime"
	controller "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	grafana_v1alpha1 "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	promop_v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	promop "github.com/prometheus-operator/prometheus-operator/pkg/operator"
	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// Needed for clients.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// Config controls the configuration of the Operator.
type Config struct {
	LogLevel      logging.Level
	LogFormat     logging.Format
	Labels        promop.Labels
	Controller    controller.Options
	AgentSelector string

	// TODO(rfratto): extra settings from Prometheus Operator:
	//
	// 1. Reloader container image/requests/limits
	// 2. Namespaces allow/denylist.
	// 3. Namespaces for Prometheus resources.
}

// NewConfig creates a new Config and initializes default values.
// Flags will be regsitered against f if it is non-nil.
func NewConfig(f *flag.FlagSet) (*Config, error) {
	if f == nil {
		f = flag.NewFlagSet("temp", flag.PanicOnError)
	}

	var c Config
	err := c.registerFlags(f)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) registerFlags(f *flag.FlagSet) error {
	c.LogLevel.RegisterFlags(f)
	c.LogFormat.RegisterFlags(f)
	f.Var(&c.Labels, "labels", "Labels to add to all created operator resources")
	f.StringVar(&c.AgentSelector, "agent-selector", "", "Label selector to discover GrafanaAgent CRs. Defaults to all GrafanaAgent CRs.")

	f.StringVar(&c.Controller.Namespace, "namespace", "", "Namespace to restrict the Operator to.")
	f.StringVar(&c.Controller.Host, "listen-host", "", "Host to listen on. Empty string means all interfaces.")
	f.IntVar(&c.Controller.Port, "listen-port", 9443, "Port to listen on.")
	f.StringVar(&c.Controller.MetricsBindAddress, "metrics-listen-address", ":8080", "Address to expose Operator metrics on")
	f.StringVar(&c.Controller.HealthProbeBindAddress, "health-listen-address", "", "Address to expose Operator health probes on")

	// Custom initial values for the endpoint names.
	c.Controller.ReadinessEndpointName = "/-/ready"
	c.Controller.LivenessEndpointName = "/-/healthy"

	c.Controller.Scheme = runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{
		core_v1.AddToScheme,
		apps_v1.AddToScheme,
		grafana_v1alpha1.AddToScheme,
		promop_v1.AddToScheme,
	} {
		if err := add(c.Controller.Scheme); err != nil {
			return fmt.Errorf("unable to register scheme: %w", err)
		}
	}

	return nil
}

// Manager initializes the controller for this Config.
func (c *Config) Manager() (manager.Manager, error) {
	return controller.NewManager(controller.GetConfigOrDie(), c.Controller)
}

// New creates a new Operator managed by a specific Manager. Start the Manager
// to also start the Operator.
func New(l log.Logger, c *Config, m manager.Manager) error {
	var (
		events = newResourceEventHandlers(m.GetClient(), l)

		applyGVK  = func(obj client.Object) client.Object { return applyGVK(obj, m) }
		watchType = func(obj client.Object) source.Source { return watchType(obj, m) }

		agentPredicates []predicate.Predicate
	)

	// Initialize agentPredicates if an GrafanaAgent selector is configured.
	if c.AgentSelector != "" {
		sel, err := meta_v1.ParseToLabelSelector(c.AgentSelector)
		if err != nil {
			return fmt.Errorf("unable to create predicate for selecting GrafanaAgent CRs: %w", err)
		}
		selPredicate, err := predicate.LabelSelectorPredicate(*sel)
		if err != nil {
			return fmt.Errorf("unable to create predicate for selecting GrafanaAgent CRs: %w", err)
		}
		agentPredicates = append(agentPredicates, selPredicate)
	}

	err := controller.NewControllerManagedBy(m).
		For(applyGVK(&grafana_v1alpha1.GrafanaAgent{}), builder.WithPredicates(agentPredicates...)).
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
			config:        c,
		})
	if err != nil {
		return fmt.Errorf("failed to create controller: %w", err)
	}

	return nil
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
// nolint: interfacer
func applyGVK(obj client.Object, m manager.Manager) client.Object {
	gvk, err := apiutil.GVKForObject(obj, m.GetScheme())
	if err != nil {
		panic(err)
	}
	obj.GetObjectKind().SetGroupVersionKind(gvk)
	return obj
}
