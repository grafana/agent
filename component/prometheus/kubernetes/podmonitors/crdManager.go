package podmonitors

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
	"github.com/pkg/errors"
	informers "github.com/prometheus-operator/prometheus-operator/pkg/client/informers/externalversions"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/scrape"
	"k8s.io/client-go/tools/cache"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// crdManager is all of the fields required to run the component.
// on update, this entire thing will be recreated and restarted
type crdManager struct {
	opts   component.Options
	logger log.Logger
	config *Arguments

	discovery        *discovery.Manager
	scraper          *scrape.Manager
	cg               configGenerator
	discoveryConfigs map[string]discovery.Configs
	scrapeConfigs    map[string]*config.ScrapeConfig

	mut sync.Mutex
}

func newManager(opts component.Options, logger log.Logger, cfg *Arguments) *crdManager {
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
	err := c.discovery.ApplyConfig(c.discoveryConfigs)
	if err != nil {
		level.Error(c.logger).Log("msg", "error applying discovery configs", "err", err)
	}
	scs := []*config.ScrapeConfig{}
	for _, sc := range c.scrapeConfigs {
		scs = append(scs, sc)
	}
	err = c.scraper.ApplyConfig(&config.Config{
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

func (c *crdManager) addPodMonitor(pm *v1.PodMonitor) {
	for i, ep := range pm.Spec.PodMetricsEndpoints {
		pmc, err := c.cg.generatePodMonitorConfig(pm, ep, i)
		if err != nil {
			level.Error(c.logger).Log("name", pm.Name, "err", err, "msg", "error generating scrapeconfig from podmonitor")
			continue
		}
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
	c.addPodMonitor(pm)
}
func (c *crdManager) onUpdatePodMonitor(oldObj, newObj interface{}) {
	pm := oldObj.(*v1.PodMonitor)
	c.clearConfigs("podMonitor", pm.Namespace, pm.Name)
	c.addPodMonitor(newObj.(*v1.PodMonitor))
}
func (c *crdManager) onDeletePodMonitor(obj interface{}) {
	pm := obj.(*v1.PodMonitor)
	c.clearConfigs("podMonitor", pm.Namespace, pm.Name)
	c.apply()
}

func (c *crdManager) run(ctx context.Context) error {
	config, err := c.config.restConfig()
	if err != nil {
		return errors.Wrap(err, "creating rest config")
	}
	promClientset, err := versioned.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "creating prometheus clientset")
	}

	c.cg = configGenerator{
		config: c.config,
	}

	c.discovery = discovery.NewManager(ctx, c.logger, discovery.Name(c.opts.ID))
	go func() {
		err := c.discovery.Run()
		if err != nil {
			level.Error(c.logger).Log("msg", "discovery manager stopped", "err", err)
		}
	}()

	for _, namespace := range c.config.Namespaces {
		factory := informers.NewSharedInformerFactoryWithOptions(promClientset,
			5*time.Minute,
			informers.WithNamespace(namespace),
			informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
				if c.config.FieldSelector != "" {
					opts.FieldSelector = c.config.FieldSelector
				}
				if c.config.LabelSelector != "" {
					opts.LabelSelector = c.config.LabelSelector
				}
			}))

		pminf := factory.Monitoring().V1().PodMonitors().Informer()
		pminf.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    c.onAddPodMonitor,
			UpdateFunc: c.onUpdatePodMonitor,
			DeleteFunc: c.onDeletePodMonitor,
		})
		err = pminf.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
			level.Error(c.logger).Log("msg", "pod monitor watcher error", "err", err)
		})
		if err != nil {
			return errors.Wrap(err, "setting watch handler")
		}

		factory.Start(ctx.Done())
	}

	level.Info(c.logger).Log("msg", "informers  started")

	flowAppendable := prometheus.NewFanout(c.config.ForwardTo, c.opts.ID, c.opts.Registerer)
	opts := &scrape.Options{}
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
			return nil
		case m := <-c.discovery.SyncCh():
			//TODO: are there cases where we modify targets?
			targetSetsChan <- m
		}
	}
}
