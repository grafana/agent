package kubernetes

import (
	"context"
	"fmt"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/metrics/scrape"
	promk8s "github.com/prometheus/prometheus/discovery/kubernetes"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

// TODO: cpeterson: use defaults from here for hcl default
var _ = promk8s.DefaultSDConfig

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

type SDConfig struct {
	APIServer          URL                `hcl:"api_server,optional"`
	Role               string             `hcl:"role"`
	KubeConfig         string             `hcl:"kubeconfig_file,optional"`
	HTTPClientConfig   HTTPClientConfig   `hcl:"http_client_config,optional"`
	NamespaceDiscovery NamespaceDiscovery `hcl:"namespaces,optional"`
	Selectors          []SelectorConfig   `hcl:"selectors,optional"`
}

func (sd *SDConfig) Convert() *promk8s.SDConfig {
	selectors := make([]promk8s.SelectorConfig, len(sd.Selectors))
	for i, s := range sd.Selectors {
		selectors[i] = *s.Convert()
	}
	return &promk8s.SDConfig{
		APIServer:          sd.APIServer.Convert(),
		Role:               promk8s.Role(sd.Role),
		KubeConfig:         sd.KubeConfig,
		HTTPClientConfig:   *sd.HTTPClientConfig.Convert(),
		NamespaceDiscovery: *sd.NamespaceDiscovery.Convert(),
		Selectors:          selectors,
	}
}

type NamespaceDiscovery struct {
	IncludeOwnNamespace bool     `hcl:"own_namespace,optional"`
	Names               []string `hcl:"names,optional"`
}

func (nd *NamespaceDiscovery) Convert() *promk8s.NamespaceDiscovery {
	return &promk8s.NamespaceDiscovery{
		IncludeOwnNamespace: nd.IncludeOwnNamespace,
		Names:               nd.Names,
	}
}

type SelectorConfig struct {
	Role  string `hcl:"role,optional"`
	Label string `hcl:"label,optional"`
	Field string `hcl:"field,optional"`
}

func (sc *SelectorConfig) Convert() *promk8s.SelectorConfig {
	return &promk8s.SelectorConfig{
		Role:  promk8s.Role(sc.Role),
		Label: sc.Label,
		Field: sc.Field,
	}
}

// Exports holds values which are exported by the discovery.k8s component.
type Exports struct {
	Targets []scrape.Target `hcl:"targets,optional"`
}

// Component implements the discovery.k8s component.
type Component struct {
	opts   component.Options
	args   SDConfig
	cancel context.CancelFunc

	ch chan []*targetgroup.Group
}

// New creates a new discovery.k8s component.
func New(o component.Options, args SDConfig) (*Component, error) {
	c := &Component{
		opts: o,
		// TODO: check buffering in prometheus service discovery manager
		ch: make(chan []*targetgroup.Group),
	}

	// Perform an update which will immediately set our exports
	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	f := func(t []scrape.Target) {
		c.opts.OnStateChange(Exports{Targets: t})
	}
	discovery.RunDiscovery(ctx, c.ch, f)
	// if we get here, we've canceled
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}

// Update implements component.Compnoent.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(SDConfig)
	fmt.Println(newArgs)
	disc, err := promk8s.New(c.opts.Logger, newArgs.Convert())
	if err != nil {
		return err
	}
	if c.cancel != nil {
		c.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	go disc.Run(ctx, c.ch)
	return nil
}
