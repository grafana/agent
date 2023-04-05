package common

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
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/scrape"
	toolscache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/grafana/agent/component/prometheus/operator"
	"github.com/grafana/agent/component/prometheus/operator/configgen"
	compscrape "github.com/grafana/agent/component/prometheus/scrape"
	promopv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// Generous timeout period for configuring all informers
const informerSyncTimeout = 10 * time.Second

// Manager is all of the fields required to run a crd based component.
// on update, this entire thing should be recreated and restarted
type crdManager struct {
	mut              sync.Mutex
	discoveryConfigs map[string]discovery.Configs
	scrapeConfigs    map[string]*config.ScrapeConfig
	debugInfo        map[string]*operator.DiscoveredPodMonitor
	discoveryManager *discovery.Manager
	scrapeManager    *scrape.Manager

	opts      component.Options
	logger    log.Logger
	args      *operator.Arguments
	configGen configgen.ConfigGenerator
}

func newCrdManager(opts component.Options, logger log.Logger, args *operator.Arguments) *crdManager {
	return &crdManager{
		opts:             opts,
		logger:           logger,
		args:             args,
		discoveryConfigs: map[string]discovery.Configs{},
		scrapeConfigs:    map[string]*config.ScrapeConfig{},
		debugInfo:        map[string]*operator.DiscoveredPodMonitor{},
	}
}

func (c *crdManager) Run(ctx context.Context) error {
	c.configGen = configgen.ConfigGenerator{
		Client: &c.args.Client,
	}

	// Start prometheus service discovery manager
	c.discoveryManager = discovery.NewManager(ctx, c.logger, discovery.Name(c.opts.ID))
	go func() {
		err := c.discoveryManager.Run()
		if err != nil {
			level.Error(c.logger).Log("msg", "discovery manager stopped", "err", err)
		}
	}()

	if err := c.runInformers(ctx); err != nil {
		return err
	}
	level.Info(c.logger).Log("msg", "informers  started")

	// Start prometheus scrape manager.
	flowAppendable := prometheus.NewFanout(c.args.ForwardTo, c.opts.ID, c.opts.Registerer)
	opts := &scrape.Options{}
	c.scrapeManager = scrape.NewManager(opts, c.logger, flowAppendable)
	defer c.scrapeManager.Stop()
	targetSetsChan := make(chan map[string][]*targetgroup.Group)
	go func() {
		err := c.scrapeManager.Run(targetSetsChan)
		level.Info(c.logger).Log("msg", "scrape manager stopped")
		if err != nil {
			level.Error(c.logger).Log("msg", "scrape manager failed", "err", err)
		}
	}()

	// Start the target discovery loop to update the scrape manager with new targets.
	for {
		select {
		case <-ctx.Done():
			return nil
		case m := <-c.discoveryManager.SyncCh():
			targetSetsChan <- m
		}
	}
}

// DebugInfo returns debug information for the CRDManager.
func (c *crdManager) DebugInfo() interface{} {
	c.mut.Lock()
	defer c.mut.Unlock()

	var info operator.DebugInfo
	for _, pm := range c.debugInfo {
		info.DiscoveredPodMonitors = append(info.DiscoveredPodMonitors, pm)
	}
	info.Targets = compscrape.BuildTargetStatuses(c.scrapeManager.TargetsActive())
	return info
}

// runInformers starts all the informers that are required to discover PodMonitors.
func (c *crdManager) runInformers(ctx context.Context) error {
	config, err := c.args.Client.BuildRESTConfig(c.logger)
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

	ls, err := c.args.LabelSelector.BuildSelector()
	if err != nil {
		return fmt.Errorf("building label selector: %w", err)
	}
	for _, ns := range c.args.Namespaces {
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
			// If the context was canceled, we don't want to log an error.
			if err != nil && ctx.Err() == nil {
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
func (c *crdManager) configureInformers(ctx context.Context, informers cache.Informers) error {
	podMonitor := &promopv1.PodMonitor{}

	informerCtx, cancel := context.WithTimeout(ctx, informerSyncTimeout)
	defer cancel()

	informer, err := informers.GetInformer(informerCtx, podMonitor)
	if err != nil {
		if errors.Is(informerCtx.Err(), context.DeadlineExceeded) { // Check the context to prevent GetInformer returning a fake timeout
			return fmt.Errorf("timeout exceeded while configuring informers. Check the connection"+
				" to the Kubernetes API is stable and that the Agent has appropriate RBAC permissions for %v", podMonitor)
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
	return nil
}

// apply applies the current state of the Manager to the Prometheus discovery manager and scrape manager.
func (c *crdManager) apply() error {
	c.mut.Lock()
	defer c.mut.Unlock()
	err := c.discoveryManager.ApplyConfig(c.discoveryConfigs)
	if err != nil {
		level.Error(c.logger).Log("msg", "error applying discovery configs", "err", err)
		return err
	}
	scs := []*config.ScrapeConfig{}
	for _, sc := range c.scrapeConfigs {
		scs = append(scs, sc)
	}
	err = c.scrapeManager.ApplyConfig(&config.Config{
		ScrapeConfigs: scs,
	})
	if err != nil {
		level.Error(c.logger).Log("msg", "error applying scrape configs", "err", err)
		return err
	}
	level.Debug(c.logger).Log("msg", "scrape config was updated")
	return nil
}

func (c *crdManager) addDebugInfo(ns string, name string, err error) {
	c.mut.Lock()
	defer c.mut.Unlock()
	debug := &operator.DiscoveredPodMonitor{}
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

func (c *crdManager) addPodMonitor(pm *promopv1.PodMonitor) {
	var err error
	for i, ep := range pm.Spec.PodMetricsEndpoints {
		var pmc *config.ScrapeConfig
		pmc, err = c.configGen.GeneratePodMonitorConfig(pm, ep, i)
		if err != nil {
			// TODO(jcreixell): Generate Kubernetes event to inform of this error when running `kubectl get <podmonitor>`.
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

func (c *crdManager) onAddPodMonitor(obj interface{}) {
	pm := obj.(*promopv1.PodMonitor)
	level.Info(c.logger).Log("msg", "found pod monitor", "name", pm.Name)
	c.addPodMonitor(pm)
}
func (c *crdManager) onUpdatePodMonitor(oldObj, newObj interface{}) {
	pm := oldObj.(*promopv1.PodMonitor)
	c.clearConfigs(pm.Namespace, pm.Name)
	c.addPodMonitor(newObj.(*promopv1.PodMonitor))
}
func (c *crdManager) onDeletePodMonitor(obj interface{}) {
	pm := obj.(*promopv1.PodMonitor)
	c.clearConfigs(pm.Namespace, pm.Name)
	if err := c.apply(); err != nil {
		level.Error(c.logger).Log("name", pm.Name, "err", err, "msg", "error applying scrape configs after podmonitor deletion")
	}
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
