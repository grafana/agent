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
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/scrape"
	toolscache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// crdManager is all of the fields required to run the component.
// on update, this entire thing will be recreated and restarted
type crdManager struct {
	mut              sync.Mutex
	discoveryConfigs map[string]discovery.Configs
	scrapeConfigs    map[string]*config.ScrapeConfig
	debugInfo        map[string]*discoveredPodMonitor
	discovery        *discovery.Manager
	scraper          *scrape.Manager

	opts   component.Options
	logger log.Logger
	config *Arguments
	cg     configGenerator
}

func newManager(opts component.Options, logger log.Logger, cfg *Arguments) *crdManager {
	return &crdManager{
		opts:             opts,
		logger:           logger,
		config:           cfg,
		discoveryConfigs: map[string]discovery.Configs{},
		scrapeConfigs:    map[string]*config.ScrapeConfig{},
		debugInfo:        map[string]*discoveredPodMonitor{},
	}
}

func (c *crdManager) run(ctx context.Context) error {

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

	if err := c.runInformers(ctx); err != nil {
		return err
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

func (c *crdManager) runInformers(ctx context.Context) error {
	config, err := c.config.Client.BuildRESTConfig(c.logger)
	if err != nil {
		return errors.Wrap(err, "creating rest config")
	}

	scheme := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{
		v1.AddToScheme,
	} {
		if err := add(scheme); err != nil {
			return fmt.Errorf("unable to register scheme: %w", err)
		}
	}

	ls, err := c.config.LabelSelector.BuildSelector()
	if err != nil {
		return errors.Wrap(err, "building label selector")
	}
	for _, ns := range c.config.Namespaces {
		opts := cache.Options{
			Scheme:    scheme,
			Namespace: ns,
		}

		if ls != labels.Nothing() {
			opts.DefaultSelector.Label = ls
		}
		// TODO: field selector needs to be cloned into config
		cache, err := cache.New(config, opts)
		if err != nil {
			return err
		}

		informers := cache

		go func() {
			err := informers.Start(ctx)
			if err != nil && ctx.Err() != nil {
				level.Error(c.logger).Log("msg", "failed to start informers", "err", err)
			}
		}()
		if !informers.WaitForCacheSync(ctx) {
			return fmt.Errorf("informer caches failed to sync")
		}
		if err := c.configureInformers(ctx, informers); err != nil {
			return fmt.Errorf("failed to configure informers: %w", err)
		}
	}

	return nil
}

// Generous timeout period for configuring all informers
const informerSyncTimeout = 10 * time.Second

func (c *crdManager) configureInformers(ctx context.Context, informers cache.Informers) error {
	types := []client.Object{
		&v1.PodMonitor{},
	}

	informerCtx, cancel := context.WithTimeout(ctx, informerSyncTimeout)
	defer cancel()

	for _, ty := range types {
		informer, err := informers.GetInformer(informerCtx, ty)
		if err != nil {
			if errors.Is(informerCtx.Err(), context.DeadlineExceeded) { // Check the context to prevent GetInformer returning a fake timeout
				return fmt.Errorf("Timeout exceeded while configuring informers. Check the connection"+
					" to the Kubernetes API is stable and that the Agent has appropriate RBAC permissions for %v", ty)
			}

			return err
		}
		informer.AddEventHandler((toolscache.ResourceEventHandlerFuncs{
			AddFunc:    c.onAddPodMonitor,
			UpdateFunc: c.onUpdatePodMonitor,
			DeleteFunc: c.onDeletePodMonitor,
		}))
	}
	return nil
}

func (c *crdManager) apply() error {
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
	return nil
}

func (c *crdManager) clearConfigs(ns string, name string) {
	c.mut.Lock()
	defer c.mut.Unlock()
	prefix := fmt.Sprintf("podMonitor/%s/%s", ns, name)
	for k := range c.discoveryConfigs {
		if strings.HasPrefix(k, prefix) {
			delete(c.discoveryConfigs, k)
			delete(c.scrapeConfigs, k)
		}
	}
	delete(c.debugInfo, prefix)
}

func (c *crdManager) addDebugInfo(ns string, name string, err error) {
	c.mut.Lock()
	defer c.mut.Unlock()
	debug := &discoveredPodMonitor{}
	debug.Namespace = ns
	debug.Name = name
	debug.LastReconcile = time.Now()
	if err != nil {
		debug.ReconcileError = err.Error()
	} else {
		debug.ReconcileError = ""
	}
	prefix := fmt.Sprintf("podMonitor/%s/%s", ns, name)
	c.debugInfo[prefix] = debug
}

func (c *crdManager) addPodMonitor(pm *v1.PodMonitor) {
	var err error
	for i, ep := range pm.Spec.PodMetricsEndpoints {
		var pmc *config.ScrapeConfig
		pmc, err = c.cg.generatePodMonitorConfig(pm, ep, i)
		if err != nil {
			level.Error(c.logger).Log("name", pm.Name, "err", err, "msg", "error generating scrapeconfig from podmonitor")
			break
		}
		c.mut.Lock()
		c.discoveryConfigs[pmc.JobName] = pmc.ServiceDiscoveryConfigs
		c.scrapeConfigs[pmc.JobName] = pmc
		c.mut.Unlock()
	}
	if err != nil {
		c.addDebugInfo(pm.Namespace, pm.Name, err)
		return
	}
	if err = c.apply(); err != nil {
		level.Error(c.logger).Log("name", pm.Name, "err", err, "msg", "error applying scrapeconfig from podmonitor")
	}
	c.addDebugInfo(pm.Namespace, pm.Name, err)
}

func (c *crdManager) onAddPodMonitor(obj interface{}) {
	pm := obj.(*v1.PodMonitor)
	level.Info(c.logger).Log("msg", "found pod monitor", "name", pm.Name)
	c.addPodMonitor(pm)
}
func (c *crdManager) onUpdatePodMonitor(oldObj, newObj interface{}) {
	pm := oldObj.(*v1.PodMonitor)
	c.clearConfigs(pm.Namespace, pm.Name)
	c.addPodMonitor(newObj.(*v1.PodMonitor))
}
func (c *crdManager) onDeletePodMonitor(obj interface{}) {
	pm := obj.(*v1.PodMonitor)
	c.clearConfigs(pm.Namespace, pm.Name)
	c.apply()
}
