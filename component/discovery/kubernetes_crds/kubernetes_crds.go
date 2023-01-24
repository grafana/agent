package kubernetes_crds

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
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

type Component struct {
	opts   component.Options
	config *Config

	onUpdate chan struct{}
	mut      sync.Mutex
}

// crdManager is all of the fields required to run the component.
// on update, this entire thing will be recreated and restarted
type crdManager struct {
	opts   component.Options
	logger log.Logger
	config *Config

	discovery        *discovery.Manager
	scraper          *scrape.Manager
	cg               configGenerator
	discoveryConfigs map[string]discovery.Configs
	scrapeConfigs    map[string]*config.ScrapeConfig

	// TODO: mutex only needed here if informer calls event handlers concurrently
	mut sync.Mutex
}

func New(o component.Options, args component.Arguments) (*Component, error) {
	c := &Component{
		opts:     o,
		onUpdate: make(chan struct{}, 1),
	}
	return c, c.Update(args)
}

func newManager(opts component.Options, logger log.Logger, cfg *Config) *crdManager {
	return &crdManager{
		opts:             opts,
		logger:           logger,
		config:           cfg,
		discoveryConfigs: map[string]discovery.Configs{},
		scrapeConfigs:    map[string]*config.ScrapeConfig{},
	}
}

func (c *crdManager) apply() {
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
		level.Error(c.logger).Log("msg", "error applying scrape configs", "err", err)
	}
	level.Debug(c.logger).Log("msg", "scrape config was updated")
}

func (c *crdManager) clearConfigs(kind string, ns string, name string) {
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

func (c *crdManager) addConfig(pm *v1.PodMonitor) {
	for i, ep := range pm.Spec.PodMetricsEndpoints {
		pmc := c.cg.generatePodMonitorConfig(pm, ep, i)
		c.mut.Lock()
		c.discoveryConfigs[pmc.JobName] = pmc.ServiceDiscoveryConfigs
		c.scrapeConfigs[pmc.JobName] = pmc
		c.mut.Unlock()
	}
	c.apply()
}

func (c *crdManager) onAddPodMonitor(obj interface{}) {
	pm := obj.(*v1.PodMonitor)
	level.Info(c.logger).Log("msg", "found pod monitor", "name", pm.Name)
	c.addConfig(pm)
}

func (c *crdManager) onUpdatePodMonitor(oldObj, newObj interface{}) {
	pm := oldObj.(*v1.PodMonitor)
	level.Info(c.logger).Log("msg", "found pod monitor", "name", pm.Name)
	c.clearConfigs("podMonitor", pm.Namespace, pm.Name)
	c.addConfig(newObj.(*v1.PodMonitor))
}
func (c *crdManager) onDeletePodMonitor(obj interface{}) {
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
			crdMan := newManager(c.opts, c.opts.Logger, componentCfg)
			go crdMan.run(innerCtx)
		}
	}
}

func (c *crdManager) run(ctx context.Context) {

	config, err := c.config.restConfig()
	if err != nil {
		level.Error(c.logger).Log("msg", "failed to create rest config", "err", err)
		return
	}
	clientset, err := kubernetes.NewForConfig(config)

	promClientset := versioned.New(clientset.RESTClient())
	if err != nil {
		level.Error(c.logger).Log("msg", "failed to create rest client", "err", err)
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
		level.Error(c.logger).Log("msg", "kubernetes watcher error", "err", err)
	})
	factory.Start(ctx.Done())

	level.Info(c.logger).Log("msg", "Informer factory started")

	c.discovery = discovery.NewManager(ctx, c.logger, discovery.Name(c.opts.ID))
	go func() {
		err := c.discovery.Run()
		if err != nil {
			level.Error(c.logger).Log("msg", "discovery manager stopped", "err", err)
		}
	}()

	flowAppendable := prometheus.NewFanout(c.config.ForwardTo, c.opts.ID, c.opts.Registerer)
	opts := &scrape.Options{
		HTTPClientOptions: []promCommonConfig.HTTPClientOption{
			promCommonConfig.WithFS(fs),
		},
	}
	c.scraper = scrape.NewManager(opts, c.logger, flowAppendable)
	defer c.scraper.Stop()
	targetSetsChan := make(chan map[string][]*targetgroup.Group)
	go func() {
		err := c.scraper.Run(targetSetsChan)
		level.Info(c.logger).Log("msg", "scrape manager stopped")
		if err != nil {
			level.Error(c.logger).Log("msg", "scrape manager failed", "err", err)
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
	c.mut.Unlock()
	select {
	case c.onUpdate <- struct{}{}:
	default:
	}
	return nil
}
