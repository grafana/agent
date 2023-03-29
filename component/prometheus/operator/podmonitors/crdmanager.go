package podmonitors

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus"
	"github.com/grafana/agent/component/prometheus/operator/podmonitors/config_gen"
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

// Generous timeout period for configuring all informers
const informerSyncTimeout = 10 * time.Second

// CRDManager is all of the fields required to run the component.
// on update, this entire thing will be recreated and restarted
// TODO: make fields public
type CRDManager struct {
	mut              sync.Mutex
	discoveryConfigs map[string]discovery.Configs
	scrapeConfigs    map[string]*config.ScrapeConfig
	debugInfo        map[string]*DiscoveredPodMonitor
	discovery        *discovery.Manager
	scraper          *scrape.Manager

	opts   component.Options
	logger log.Logger
	config *Arguments
	cg     config_gen.ConfigGenerator
}

func NewCRDManager(opts component.Options, logger log.Logger, cfg *Arguments) *CRDManager {
	return &CRDManager{
		opts:             opts,
		logger:           logger,
		config:           cfg,
		discoveryConfigs: map[string]discovery.Configs{},
		scrapeConfigs:    map[string]*config.ScrapeConfig{},
		debugInfo:        map[string]*DiscoveredPodMonitor{},
	}
}

func (c *CRDManager) Run(ctx context.Context) error {
	c.cg = config_gen.ConfigGenerator{
		Client: &c.config.Client,
	}

	// Start prometheus discovery manager (service discovery)
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

	// Start prometheus scrape manager
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

func (c *CRDManager) runInformers(ctx context.Context) error {
	config, err := c.config.Client.BuildRESTConfig(c.logger)
	if err != nil {
		return fmt.Errorf("creating rest config: %w", err)
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
		return fmt.Errorf("building label selector: %w", err)
	}
	for _, ns := range c.config.Namespaces {
		opts := cache.Options{
			Scheme:    scheme,
			Namespace: ns,
		}

		if ls != labels.Nothing() {
			opts.DefaultSelector.Label = ls
		}
		// TODO: field selector needs to be cloned into config (ask Craig about this)
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

// configureInformers configures the informers for the CRDManager to watch for PodMonitors changes.
func (c *CRDManager) configureInformers(ctx context.Context, informers cache.Informers) error {
	objects := []client.Object{
		&v1.PodMonitor{},
	}

	informerCtx, cancel := context.WithTimeout(ctx, informerSyncTimeout)
	defer cancel()

	for _, obj := range objects {
		informer, err := informers.GetInformer(informerCtx, obj)
		if err != nil {
			if errors.Is(informerCtx.Err(), context.DeadlineExceeded) { // Check the context to prevent GetInformer returning a fake timeout
				return fmt.Errorf("Timeout exceeded while configuring informers. Check the connection"+
					" to the Kubernetes API is stable and that the Agent has appropriate RBAC permissions for %v", obj)
			}

			return err
		}
		_, err = informer.AddEventHandler((toolscache.ResourceEventHandlerFuncs{
			AddFunc:    c.onAddPodMonitor,
			UpdateFunc: c.onUpdatePodMonitor,
			DeleteFunc: c.onDeletePodMonitor,
		}))
		if err != nil {
			return err
		}
	}
	return nil
}

// apply applies the current state of the CRDManager to the Prometheus discovery manager and scrape manager.
func (c *CRDManager) apply() error {
	c.mut.Lock()
	defer c.mut.Unlock()
	err := c.discovery.ApplyConfig(c.discoveryConfigs)
	if err != nil {
		level.Error(c.logger).Log("msg", "error applying discovery configs", "err", err)
		return err
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
		return err
	}
	level.Debug(c.logger).Log("msg", "scrape config was updated")
	return nil
}

func (c *CRDManager) clearConfigs(ns string, name string) {
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

func (c *CRDManager) addDebugInfo(ns string, name string, err error) {
	c.mut.Lock()
	defer c.mut.Unlock()
	debug := &DiscoveredPodMonitor{}
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

func (c *CRDManager) addPodMonitor(pm *v1.PodMonitor) {
	var err error
	for i, ep := range pm.Spec.PodMetricsEndpoints {
		var pmc *config.ScrapeConfig
		pmc, err = c.cg.GeneratePodMonitorConfig(pm, ep, i)
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
		level.Error(c.logger).Log("name", pm.Name, "err", err, "msg", "error applying scrape configs from podmonitor")
	}
	c.addDebugInfo(pm.Namespace, pm.Name, err)
}

func (c *CRDManager) onAddPodMonitor(obj interface{}) {
	pm := obj.(*v1.PodMonitor)
	level.Info(c.logger).Log("msg", "found pod monitor", "name", pm.Name)
	c.addPodMonitor(pm)
}
func (c *CRDManager) onUpdatePodMonitor(oldObj, newObj interface{}) {
	pm := oldObj.(*v1.PodMonitor)
	c.clearConfigs(pm.Namespace, pm.Name)
	c.addPodMonitor(newObj.(*v1.PodMonitor))
}
func (c *CRDManager) onDeletePodMonitor(obj interface{}) {
	pm := obj.(*v1.PodMonitor)
	c.clearConfigs(pm.Namespace, pm.Name)
	if err := c.apply(); err != nil {
		level.Error(c.logger).Log("name", pm.Name, "err", err, "msg", "error applying scrape configs after podmonitor deletion")
	}
}
