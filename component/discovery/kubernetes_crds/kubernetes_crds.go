package kubernetes_crds

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus"
	promCommonConfig "github.com/prometheus/common/config"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/scrape"
	"github.com/psanford/memfs"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	promop "github.com/prometheus-operator/prometheus-operator/pkg/client/informers/externalversions"
	"github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
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

// TODO: make new type for most of the settable fields for run, that we just replace entirely on update.
// less locking that way.
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
		// TODO: need to pass in number of shards
		pmc := c.cg.generatePodMonitorConfig(pm, ep, i, 0)
		c.mut.Lock()
		c.discoveryConfigs[pmc.JobName] = pmc.ServiceDiscoveryConfigs
		c.scrapeConfigs[pmc.JobName] = pmc
		c.mut.Unlock()
	}
	c.apply()
}

func (c *Component) onAddPodMonitor(obj interface{}) {
	pm := obj.(*v1.PodMonitor)
	level.Info(c.opts.Logger).Log("msg", "found pod monitor", "name", pm.Name)
	c.addConfig(pm)
}

func (c *Component) onUpdatePodMonitor(oldObj, newObj interface{}) {
	pm := oldObj.(*v1.PodMonitor)
	level.Info(c.opts.Logger).Log("msg", "found pod monitor", "name", pm.Name)
	c.clearConfigs("podMonitor", pm.Namespace, pm.Name)
	c.addConfig(newObj.(*v1.PodMonitor))
}
func (c *Component) onDeletePodMonitor(obj interface{}) {
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
	clientset, err := kubernetes.NewForConfig(config)

	promClientset := versioned.New(clientset.RESTClient())
	if err != nil {
		level.Error(c.opts.Logger).Log("msg", "failed to create rest client", "err", err)
		return
	}

	fs := memfs.New()
	sm := &secretManager{
		fs:     fs,
		client: clientset,
	}
	c.cg = configGenerator{
		config:  c.config,
		secrets: sm,
	}

	factory := promop.NewSharedInformerFactory(promClientset, 5*time.Minute)
	inf := factory.Monitoring().V1().PodMonitors().Informer()
	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAddPodMonitor,
		UpdateFunc: c.onUpdatePodMonitor,
		DeleteFunc: c.onDeletePodMonitor,
	})
	inf.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
		level.Error(c.opts.Logger).Log("msg", "kubernetes watcher error", "err", err)
	})
	factory.Start(ctx.Done())

	level.Info(c.opts.Logger).Log("msg", "Informer factory started")

	c.discovery = discovery.NewManager(ctx, c.opts.Logger, discovery.Name(c.opts.ID))
	go func() {
		err := c.discovery.Run()
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "discovery manager stopped", "err", err)
		}
	}()

	flowAppendable := prometheus.NewFanout(componentCfg.ForwardTo, c.opts.ID, c.opts.Registerer)
	opts := &scrape.Options{
		HTTPClientOptions: []promCommonConfig.HTTPClientOption{
			promCommonConfig.WithFS(fs),
		},
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
	c.discoveryConfigs = map[string]discovery.Configs{}
	c.scrapeConfigs = map[string]*config.ScrapeConfig{}
	c.mut.Unlock()
	select {
	case c.onUpdate <- struct{}{}:
	default:
	}
	return nil
}
