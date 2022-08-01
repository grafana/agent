package kubernetes

import (
	"context"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/metrics/scrape"
	promk8s "github.com/prometheus/prometheus/discovery/kubernetes"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.k8s",
		Args:    SDConfig{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(SDConfig))
		},
	})
}

// SDConfig is a conversion of discover/kubernetes/SDConfig to be compatible with flow
type SDConfig struct {
	APIServer          config.URL              `river:"api_server,attr,optional"`
	Role               string                  `river:"role,attr"`
	KubeConfig         string                  `river:"kubeconfig_file,attr,optional"`
	HTTPClientConfig   config.HTTPClientConfig `river:"http_client_config,attr,optional"`
	NamespaceDiscovery NamespaceDiscovery      `river:"namespaces,attr,optional"`
	Selectors          []SelectorConfig        `river:"selectors,attr,optional"`
}

// DefaultConfig holds defaults for SDConfig. (copied from prometheus)
var DefaultConfig = SDConfig{
	HTTPClientConfig: config.DefaultHTTPClientConfig,
}

// UnmarshalRiver simply applies defaults then unmarshals regularly
func (sd *SDConfig) UnmarshalRiver(f func(interface{}) error) error {
	*sd = DefaultConfig
	type arguments SDConfig
	return f((*arguments)(sd))
}

// Convert to prometheus config type
func (sd *SDConfig) Convert() *promk8s.SDConfig {
	selectors := make([]promk8s.SelectorConfig, len(sd.Selectors))
	for i, s := range sd.Selectors {
		selectors[i] = *s.convert()
	}
	return &promk8s.SDConfig{
		APIServer:          sd.APIServer.Convert(),
		Role:               promk8s.Role(sd.Role),
		KubeConfig:         sd.KubeConfig,
		HTTPClientConfig:   *sd.HTTPClientConfig.Convert(),
		NamespaceDiscovery: *sd.NamespaceDiscovery.convert(),
		Selectors:          selectors,
	}
}

// NamespaceDiscovery mirroring prometheus type
type NamespaceDiscovery struct {
	IncludeOwnNamespace bool     `river:"own_namespace,attr,optional"`
	Names               []string `river:"names,attr,optional"`
}

func (nd *NamespaceDiscovery) convert() *promk8s.NamespaceDiscovery {
	return &promk8s.NamespaceDiscovery{
		IncludeOwnNamespace: nd.IncludeOwnNamespace,
		Names:               nd.Names,
	}
}

// SelectorConfig mirroring prometheus type
type SelectorConfig struct {
	Role  string `river:"role,attr,optional"`
	Label string `river:"label,attr,optional"`
	Field string `river:"field,attr,optional"`
}

func (sc *SelectorConfig) convert() *promk8s.SelectorConfig {
	return &promk8s.SelectorConfig{
		Role:  promk8s.Role(sc.Role),
		Label: sc.Label,
		Field: sc.Field,
	}
}

// Exports holds values which are exported by the discovery.k8s component.
type Exports struct {
	Targets []scrape.Target `river:"targets,attr"`
}

// Component implements the discovery.k8s component.
type Component struct {
	opts   component.Options
	cancel context.CancelFunc
}

// New creates a new discovery.k8s component.
func New(o component.Options, args SDConfig) (*Component, error) {
	c := &Component{
		opts: o,
	}
	return c, c.Update(args)
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}

// Update implements component.Compnoent.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(SDConfig)
	disc, err := promk8s.New(c.opts.Logger, newArgs.Convert())
	if err != nil {
		return err
	}
	// cancel any previously running discovery
	if c.cancel != nil {
		c.cancel()
	}
	// function to send updates on change
	f := func(t []scrape.Target) {
		c.opts.OnStateChange(Exports{Targets: t})
	}
	// create new context so we can cancel it if we get any future updates
	newCtx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	// finally run discovery
	go discovery.RunDiscovery(newCtx, disc, f)
	return nil
}
