//go:build linux && (amd64 || arm64)

package beyla

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"
	"sync"

	"github.com/grafana/agent/internal/component"
	"github.com/grafana/agent/internal/component/discovery"
	"github.com/grafana/agent/internal/component/otelcol"
	"github.com/grafana/agent/internal/featuregate"
	http_service "github.com/grafana/agent/internal/service/http"
	"github.com/grafana/beyla/pkg/beyla"
	"github.com/grafana/beyla/pkg/components"
	"github.com/grafana/beyla/pkg/services"
	"github.com/grafana/beyla/pkg/transform"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
)

func init() {
	component.Register(component.Registration{
		Name:      "beyla.ebpf",
		Stability: featuregate.StabilityBeta,
		Args:      Arguments{},
		Exports:   Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Component struct {
	opts   component.Options
	mut    sync.Mutex
	args   Arguments
	reload chan struct{}
	reg    *prometheus.Registry
}

// Arguments configures the Beyla component.
type Arguments struct {
	Port           string                     `river:"open_port,attr,optional"`
	ExecutableName string                     `river:"executable_name,attr,optional"`
	Routes         Routes                     `river:"routes,block,optional"`
	Attributes     Attributes                 `river:"attributes,block,optional"`
	Discovery      Discovery                  `river:"discovery,block,optional"`
	Output         *otelcol.ConsumerArguments `river:"output,block,optional"`
}

type Routes struct {
	Unmatch        string   `river:"unmatched,attr,optional"`
	Patterns       []string `river:"patterns,attr,optional"`
	IgnorePatterns []string `river:"ignored_patterns,attr,optional"`
	IgnoredEvents  string   `river:"ignore_mode,attr,optional"`
}

func (args Routes) Convert() *transform.RoutesConfig {
	return &transform.RoutesConfig{
		Unmatch:        transform.UnmatchType(args.Unmatch),
		Patterns:       args.Patterns,
		IgnorePatterns: args.IgnorePatterns,
		IgnoredEvents:  transform.IgnoreMode(args.IgnoredEvents),
	}
}

type Attributes struct {
	Kubernetes KubernetesDecorator `river:"kubernetes,block"`
}

type KubernetesDecorator struct {
	Enable string `river:"enable,attr"`
}

func (args Attributes) Convert() beyla.Attributes {
	return beyla.Attributes{
		Kubernetes: transform.KubernetesDecorator{
			Enable: transform.KubeEnableFlag(args.Kubernetes.Enable),
		},
	}
}

type Discovery struct {
	Services Services `river:"services,block"`
}

func (args Discovery) Convert() (services.DiscoveryConfig, error) {
	srv, err := args.Services.Convert()
	if err != nil {
		return services.DiscoveryConfig{}, err
	}
	return services.DiscoveryConfig{
		Services: srv,
	}, nil
}

type Services []Service

type Service struct {
	Name      string `river:"name,attr,optional"`
	Namespace string `river:"namespace,attr,optional"`
	OpenPorts string `river:"open_ports,attr,optional"`
	Path      string `river:"exe_path,attr,optional"`
}

func (args Services) Convert() (services.DefinitionCriteria, error) {
	var attrs services.DefinitionCriteria
	for _, s := range args {
		ports, err := stringToPortEnum(s.OpenPorts)
		if err != nil {
			return nil, err
		}
		paths, err := stringToRegexpAttr(s.Path)
		if err != nil {
			return nil, err
		}
		attrs = append(attrs, services.Attributes{
			Name:      s.Name,
			Namespace: s.Namespace,
			OpenPorts: ports,
			Path:      paths,
		})
	}
	return attrs, nil
}

type Exports struct {
	Targets []discovery.Target `river:"targets,attr"`
}

func New(opts component.Options, args Arguments) (*Component, error) {
	reg := prometheus.NewRegistry()
	c := &Component{
		opts:   opts,
		args:   args,
		reload: make(chan struct{}, 1),
		reg:    reg,
	}

	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	var cancel context.CancelFunc
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-c.reload:
			// cancel any previously running exporter
			if cancel != nil {
				cancel()
			}
			newCtx, cancelFunc := context.WithCancel(ctx)
			cancel = cancelFunc

			c.mut.Lock()
			cfg, err := c.args.Convert()
			if err != nil {
				return fmt.Errorf("failed to convert arguments: %w", err)
			}
			cfg.Prometheus.Registry = c.reg
			c.mut.Unlock()
			components.RunBeyla(newCtx, cfg)
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	c.mut.Lock()
	baseTarget, err := c.baseTarget()
	if err != nil {
		return err
	}
	c.opts.OnStateChange(Exports{
		Targets: []discovery.Target{baseTarget},
	})
	c.mut.Unlock()
	select {
	case c.reload <- struct{}{}:
	default:
	}
	return nil
}

func (c *Component) baseTarget() (discovery.Target, error) {
	data, err := c.opts.GetServiceData(http_service.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get HTTP information: %w", err)
	}
	httpData := data.(http_service.Data)

	return discovery.Target{
		model.AddressLabel:     httpData.MemoryListenAddr,
		model.SchemeLabel:      "http",
		model.MetricsPathLabel: path.Join(httpData.HTTPPathForComponent(c.opts.ID), "metrics"),
		"instance":             defaultInstance(),
		"job":                  "beyla",
	}, nil

}

func (c *Component) Handler() http.Handler {
	return promhttp.HandlerFor(c.reg, promhttp.HandlerOpts{})
}

func (a *Arguments) Convert() (*beyla.Config, error) {
	var err error
	cfg := &beyla.DefaultConfig
	if a.Output != nil {
		cfg.TracesReceiver = convertTraceConsumers(a.Output.Traces)
	}
	cfg.Port, err = stringToPortEnum(a.Port)
	if err != nil {
		return nil, err
	}
	cfg.Exec, err = stringToRegexpAttr(a.ExecutableName)
	if err != nil {
		return nil, err
	}
	cfg.Routes = a.Routes.Convert()
	cfg.Attributes = a.Attributes.Convert()
	cfg.Discovery, err = a.Discovery.Convert()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func stringToRegexpAttr(s string) (services.RegexpAttr, error) {
	if s == "" {
		return services.RegexpAttr{}, nil
	}
	re, err := regexp.Compile(s)
	if err != nil {
		return services.RegexpAttr{}, err
	}
	return services.NewPathRegexp(re), nil
}

func stringToPortEnum(s string) (services.PortEnum, error) {
	if s == "" {
		return services.PortEnum{}, nil
	}
	p := services.PortEnum{}
	err := p.UnmarshalText([]byte(s))
	if err != nil {
		return services.PortEnum{}, err
	}
	return p, nil
}

func convertTraceConsumers(consumers []otelcol.Consumer) beyla.TracesReceiverConfig {
	convertedConsumers := make([]beyla.Consumer, len(consumers))
	for i, trace := range consumers {
		convertedConsumers[i] = trace
	}
	return beyla.TracesReceiverConfig{
		Traces: convertedConsumers,
	}
}

func defaultInstance() string {
	hostname := os.Getenv("HOSTNAME")
	if hostname != "" {
		return hostname
	}

	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
