package kubernetes_crds

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	commonConfig "github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/prometheus"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	promop "github.com/prometheus-operator/prometheus-operator/pkg/client/informers/externalversions"
	versioned "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.kubernetes_crds",
		Args:    Config{},
		Exports: struct{}{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args)
		},
	})
}

type Config struct {
	// Local kubeconfig to access cluster
	KubeConfig string `river:"kubeconfig_file,attr,optional"`
	// APIServerConfig allows specifying a host and auth methods to access apiserver.
	// If left empty, Prometheus is assumed to run inside of the cluster
	// and will discover API servers automatically and use the pod's CA certificate
	// and bearer token file at /var/run/secrets/kubernetes.io/serviceaccount/.
	ApiServerConfig *APIServerConfig `river:"api_server,block,optional"`

	ForwardTo []storage.Appendable `river:"forward_to,attr"`
}

// APIServerConfig defines a host and auth methods to access apiserver.
// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config
type APIServerConfig struct {
	// Host of apiserver.
	// A valid string consisting of a hostname or IP followed by an optional port number
	Host string `json:"host"`
	// BasicAuth allow an endpoint to authenticate over basic authentication
	BasicAuth *commonConfig.BasicAuth `json:"basicAuth,omitempty"`
	// Bearer token for accessing apiserver.
	BearerToken string `json:"bearerToken,omitempty"`
	// File to read bearer token for accessing apiserver.
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`
	// TLS Config to use for accessing apiserver.
	TLSConfig commonConfig.TLSConfig `json:"tlsConfig,omitempty"`
	// Authorization section for accessing apiserver
	Authorization commonConfig.Authorization `json:"authorization,omitempty"`
}

func (c *Config) restConfig() (*rest.Config, error) {
	if c.KubeConfig != "" {
		return clientcmd.BuildConfigFromFlags("", c.KubeConfig)
	}
	if c.ApiServerConfig == nil {
		return rest.InClusterConfig()
	}
	// TODO
	log.Fatal("Convert apiserverconfig directly")
	return nil, nil
}

type Component struct {
	opts      component.Options
	discovery *discovery.Manager
	scraper   *scrape.Manager

	config           *Config
	cg               configGenerator
	discoveryConfigs map[string]discovery.Configs
	scrapeConfigs    map[string]*config.ScrapeConfig

	onUpdate chan struct{}
	mut      sync.Mutex
}

func New(o component.Options, args component.Arguments) (*Component, error) {

	c := &Component{
		opts:             o,
		discoveryConfigs: map[string]discovery.Configs{},
		scrapeConfigs:    map[string]*config.ScrapeConfig{},
		onUpdate:         make(chan struct{}, 1),
	}
	return c, c.Update(args)
}

func (c *Component) apply() {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.discovery.ApplyConfig(c.discoveryConfigs)
	scs := []*config.ScrapeConfig{}
	for _, sc := range c.scrapeConfigs {
		// set defaults from prom defaults. TODO: better way to do this
		for _, r := range sc.RelabelConfigs {
			if r.Action == "" {
				r.Action = "replace"
			}
			if r.Separator == "" {
				r.Separator = ";"
			}
			if r.Regex.Regexp == nil {
				r.Regex = relabel.MustNewRegexp("(.*)")
			}
			if r.Replacement == "" {
				r.Replacement = "$1"
			}
		}
		scs = append(scs, sc)
	}
	err := c.scraper.ApplyConfig(&config.Config{
		ScrapeConfigs: scs,
	})
	if err != nil {
		level.Error(c.opts.Logger).Log("msg", "error applying scrape configs", "err", err)
	}
	level.Debug(c.opts.Logger).Log("msg", "scrape config was updated")
}

func (c *Component) clearConfigs(kind string, ns string, name string) {
	c.mut.Lock()
	defer c.mut.Unlock()
	prefix := fmt.Sprintf("%s/%s/%s", kind, ns, name)
	for k := range c.discoveryConfigs {
		if strings.HasPrefix(k, prefix) {
			delete(c.discoveryConfigs, k)
			delete(c.scrapeConfigs, k)
		}
	}
}

func (c *Component) addConfig(pm *v1.PodMonitor) {
	for i, ep := range pm.Spec.PodMetricsEndpoints {
		pmc := c.cg.generatePodMonitorConfig(pm, ep, i)
		c.mut.Lock()
		c.discoveryConfigs[pmc.JobName] = pmc.ServiceDiscoveryConfigs
		c.scrapeConfigs[pmc.JobName] = pmc
		c.mut.Unlock()
	}
	c.apply()
}

func (c *Component) OnAddPodMonitor(obj interface{}) {
	pm := obj.(*v1.PodMonitor)
	level.Info(c.opts.Logger).Log("msg", "found pod monitor", "name", pm.Name)
	c.addConfig(pm)
}

func (c *Component) OnUpdatePodMonitor(oldObj, newObj interface{}) {
	pm := oldObj.(*v1.PodMonitor)
	level.Info(c.opts.Logger).Log("msg", "found pod monitor", "name", pm.Name)
	c.clearConfigs("podMonitor", pm.Namespace, pm.Name)
	c.addConfig(newObj.(*v1.PodMonitor))
}
func (c *Component) OnDeletePodMonitor(obj interface{}) {
	pm := obj.(*v1.PodMonitor)
	c.clearConfigs("podMonitor", pm.Namespace, pm.Name)
	c.apply()
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	// innerCtx gets passed to things we create, so we can restart everything anytime we get an update.
	// Ideally, this component has very little dynamic config, and won't have frequent updates.
	var innerCtx context.Context
	// cancel is the func we use to trigger a stop to all downstream processors we create
	var cancel func()
	for {
		select {
		case <-ctx.Done():
			if cancel != nil {
				cancel()
			}
			return nil
		case <-c.onUpdate:
			if cancel != nil {
				cancel()
			}
			innerCtx, cancel = context.WithCancel(ctx)
			c.mut.Lock()
			componentCfg := c.config
			c.mut.Unlock()
			go c.run(innerCtx, componentCfg)
		}
	}
}

func (c *Component) run(ctx context.Context, componentCfg *Config) {
	config, err := componentCfg.restConfig()
	if err != nil {
		level.Error(c.opts.Logger).Log("msg", "failed to create rest config", "err", err)
		return
	}
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		level.Error(c.opts.Logger).Log("msg", "failed to create rest client", "err", err)
		return
	}

	factory := promop.NewSharedInformerFactory(clientset, 5*time.Minute)
	inf := factory.Monitoring().V1().PodMonitors().Informer()
	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.OnAddPodMonitor,
		UpdateFunc: c.OnUpdatePodMonitor,
		DeleteFunc: c.OnDeletePodMonitor,
	})
	inf.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
		level.Error(c.opts.Logger).Log("msg", "kubernetes watcher error", "err", err)
	})
	factory.Start(ctx.Done())

	level.Info(c.opts.Logger).Log("msg", "Informer factory started")

	// TODO: mutex on setting component variables
	c.discovery = discovery.NewManager(ctx, c.opts.Logger, discovery.Name(c.opts.ID))
	go func() {
		err := c.discovery.Run()
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "discovery manager stopped", "err", err)
		}
	}()

	flowAppendable := prometheus.NewFanout(componentCfg.ForwardTo, c.opts.ID, c.opts.Registerer)
	opts := &scrape.Options{
		// TODO: any options we need to set globally? ExtraMetrics?
	}
	c.scraper = scrape.NewManager(opts, c.opts.Logger, flowAppendable)
	defer c.scraper.Stop()
	targetSetsChan := make(chan map[string][]*targetgroup.Group)

	go func() {
		err := c.scraper.Run(targetSetsChan)
		level.Info(c.opts.Logger).Log("msg", "scrape manager stopped")
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "scrape manager failed", "err", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case m := <-c.discovery.SyncCh():
			targetSetsChan <- m
		}
	}
}

// Update implements component.Compnoent.
func (c *Component) Update(args component.Arguments) error {
	c.mut.Lock()
	cfg := args.(Config)
	c.config = &cfg
	c.cg = configGenerator{
		config: c.config,
	}
	c.discoveryConfigs = map[string]discovery.Configs{}
	c.mut.Unlock()
	select {
	case c.onUpdate <- struct{}{}:
	default:
	}
	return nil
}
