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

	compscrape "github.com/grafana/agent/component/prometheus/scrape"
	promopv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// Generous timeout period for configuring all informers
const informerSyncTimeout = 10 * time.Second

// CRDManager is all of the fields required to run the component.
// on update, this entire thing will be recreated and restarted
type CRDManager struct {
	Mut              sync.Mutex
	DiscoveryConfigs map[string]discovery.Configs
	ScrapeConfigs    map[string]*config.ScrapeConfig
	debugInfo        map[string]*DiscoveredPodMonitor
	DiscoveryManager *discovery.Manager
	ScrapeManager    *scrape.Manager

	Opts      component.Options
	Logger    log.Logger
	Args      *Arguments
	ConfigGen config_gen.ConfigGenerator
}

func NewCRDManager(opts component.Options, logger log.Logger, args *Arguments) *CRDManager {
	return &CRDManager{
		Opts:             opts,
		Logger:           logger,
		Args:             args,
		DiscoveryConfigs: map[string]discovery.Configs{},
		ScrapeConfigs:    map[string]*config.ScrapeConfig{},
		debugInfo:        map[string]*DiscoveredPodMonitor{},
	}
}

func (c *CRDManager) Run(ctx context.Context) error {
	c.ConfigGen = config_gen.ConfigGenerator{
		Client: &c.Args.Client,
	}

	// Start prometheus service discovery manager
	c.DiscoveryManager = discovery.NewManager(ctx, c.Logger, discovery.Name(c.Opts.ID))
	go func() {
		err := c.DiscoveryManager.Run()
		if err != nil {
			level.Error(c.Logger).Log("msg", "discovery manager stopped", "err", err)
		}
	}()

	if err := c.runInformers(ctx); err != nil {
		return err
	}
	level.Info(c.Logger).Log("msg", "informers  started")

	// Start prometheus scrape manager.
	flowAppendable := prometheus.NewFanout(c.Args.ForwardTo, c.Opts.ID, c.Opts.Registerer)
	opts := &scrape.Options{}
	c.ScrapeManager = scrape.NewManager(opts, c.Logger, flowAppendable)
	defer c.ScrapeManager.Stop()
	targetSetsChan := make(chan map[string][]*targetgroup.Group)
	go func() {
		err := c.ScrapeManager.Run(targetSetsChan)
		level.Info(c.Logger).Log("msg", "scrape manager stopped")
		if err != nil {
			level.Error(c.Logger).Log("msg", "scrape manager failed", "err", err)
		}
	}()

	// Start the target discovery loop to update the scrape manager with new targets.
	for {
		select {
		case <-ctx.Done():
			return nil
		case m := <-c.DiscoveryManager.SyncCh():
			targetSetsChan <- m
		}
	}
}

// DebugInfo returns debug information for the CRDManager.
func (c *CRDManager) DebugInfo() interface{} {
	c.Mut.Lock()
	defer c.Mut.Unlock()

	var info DebugInfo
	for _, pm := range c.debugInfo {
		info.DiscoveredPodMonitors = append(info.DiscoveredPodMonitors, pm)
	}
	info.Targets = compscrape.BuildTargetStatuses(c.ScrapeManager.TargetsActive())
	return info
}

// runInformers starts all the informers that are required to discover PodMonitors.
func (c *CRDManager) runInformers(ctx context.Context) error {
	config, err := c.Args.Client.BuildRESTConfig(c.Logger)
	if err != nil {
		return fmt.Errorf("creating rest config: %w", err)
	}

	scheme := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{
		promopv1.AddToScheme,
	} {
		if err := add(scheme); err != nil {
			return fmt.Errorf("unable to register scheme: %w", err)
		}
	}

	ls, err := c.Args.LabelSelector.BuildSelector()
	if err != nil {
		return fmt.Errorf("building label selector: %w", err)
	}
	for _, ns := range c.Args.Namespaces {
		opts := cache.Options{
			Scheme:    scheme,
			Namespace: ns,
		}

		if ls != labels.Nothing() {
			opts.DefaultSelector.Label = ls
		}
		cache, err := cache.New(config, opts)
		if err != nil {
			return err
		}

		informers := cache

		go func() {
			err := informers.Start(ctx)
			if err != nil && ctx.Err() != nil {
				level.Error(c.Logger).Log("msg", "failed to start informers", "err", err)
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
		&promopv1.PodMonitor{},
	}

	informerCtx, cancel := context.WithTimeout(ctx, informerSyncTimeout)
	defer cancel()

	for _, obj := range objects {
		informer, err := informers.GetInformer(informerCtx, obj)
		if err != nil {
			if errors.Is(informerCtx.Err(), context.DeadlineExceeded) { // Check the context to prevent GetInformer returning a fake timeout
				return fmt.Errorf("timeout exceeded while configuring informers. Check the connection"+
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
	c.Mut.Lock()
	defer c.Mut.Unlock()
	err := c.DiscoveryManager.ApplyConfig(c.DiscoveryConfigs)
	if err != nil {
		level.Error(c.Logger).Log("msg", "error applying discovery configs", "err", err)
		return err
	}
	scs := []*config.ScrapeConfig{}
	for _, sc := range c.ScrapeConfigs {
		scs = append(scs, sc)
	}
	err = c.ScrapeManager.ApplyConfig(&config.Config{
		ScrapeConfigs: scs,
	})
	if err != nil {
		level.Error(c.Logger).Log("msg", "error applying scrape configs", "err", err)
		return err
	}
	level.Debug(c.Logger).Log("msg", "scrape config was updated")
	return nil
}

func (c *CRDManager) addDebugInfo(ns string, name string, err error) {
	c.Mut.Lock()
	defer c.Mut.Unlock()
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

func (c *CRDManager) addPodMonitor(pm *promopv1.PodMonitor) {
	var err error
	for i, ep := range pm.Spec.PodMetricsEndpoints {
		var pmc *config.ScrapeConfig
		pmc, err = c.ConfigGen.GeneratePodMonitorConfig(pm, ep, i)
		if err != nil {
			// TODO(jcreixell): Generate Kubernetes event to inform of this error when runing `kubectl get <podmonitor>`.
			level.Error(c.Logger).Log("name", pm.Name, "err", err, "msg", "error generating scrapeconfig from podmonitor")
			break
		}
		c.Mut.Lock()
		c.DiscoveryConfigs[pmc.JobName] = pmc.ServiceDiscoveryConfigs
		c.ScrapeConfigs[pmc.JobName] = pmc
		c.Mut.Unlock()
	}
	if err != nil {
		c.addDebugInfo(pm.Namespace, pm.Name, err)
		return
	}
	if err = c.apply(); err != nil {
		level.Error(c.Logger).Log("name", pm.Name, "err", err, "msg", "error applying scrape configs from podmonitor")
	}
	c.addDebugInfo(pm.Namespace, pm.Name, err)
}

func (c *CRDManager) onAddPodMonitor(obj interface{}) {
	pm := obj.(*promopv1.PodMonitor)
	level.Info(c.Logger).Log("msg", "found pod monitor", "name", pm.Name)
	c.addPodMonitor(pm)
}
func (c *CRDManager) onUpdatePodMonitor(oldObj, newObj interface{}) {
	pm := oldObj.(*promopv1.PodMonitor)
	c.clearConfigs(pm.Namespace, pm.Name)
	c.addPodMonitor(newObj.(*promopv1.PodMonitor))
}
func (c *CRDManager) onDeletePodMonitor(obj interface{}) {
	pm := obj.(*promopv1.PodMonitor)
	c.clearConfigs(pm.Namespace, pm.Name)
	if err := c.apply(); err != nil {
		level.Error(c.Logger).Log("name", pm.Name, "err", err, "msg", "error applying scrape configs after podmonitor deletion")
	}
}

func (c *CRDManager) clearConfigs(ns string, name string) {
	c.Mut.Lock()
	defer c.Mut.Unlock()
	prefix := fmt.Sprintf("podMonitor/%s/%s", ns, name)
	for k := range c.DiscoveryConfigs {
		if strings.HasPrefix(k, prefix) {
			delete(c.DiscoveryConfigs, k)
			delete(c.ScrapeConfigs, k)
		}
	}
	delete(c.debugInfo, prefix)
}
