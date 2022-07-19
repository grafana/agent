package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/grafana/agent/component"
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
	runDiscovery(ctx, c.ch, f)
	// if we get here, we've canceled
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}

func runDiscovery(ctx context.Context, ch <-chan []*targetgroup.Group, f func([]scrape.Target)) {
	cache := map[string]*targetgroup.Group{}

	dirty := false

	const maxChangeFreq = 5 * time.Second
	// this should give us 2 seconds at startup to collect some changes before sending
	var lastChange time.Time = time.Now().Add(-3 * time.Second)
	for {
		var timeChan <-chan time.Time = nil
		if dirty {
			now := time.Now()
			nextValidTime := lastChange.Add(5 * time.Second)
			if now.Unix() > nextValidTime.Unix() {
				// We are past the threshold, send change notification now
				lastChange = now
				t := []scrape.Target{}
				for _, group := range cache {
					for _, target := range group.Targets {
						m := map[string]string{}
						for k, v := range group.Labels {
							m[string(k)] = string(v)
						}
						for k, v := range target {
							m[string(k)] = string(v)
						}
						t = append(t, m)
					}
				}
				f(t)
			} else {
				// else set a timer
				timeToWait := nextValidTime.Sub(now)
				timeChan = time.After(timeToWait)
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-timeChan:
			continue
		case groups := <-ch:
			for _, group := range groups {
				if len(group.Targets) == 0 {
					delete(cache, group.Source)
				} else {
					cache[group.Source] = group
					dirty = true
				}
			}
		}
	}
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
