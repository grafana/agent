package rules

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/internal/component"
	commonK8s "github.com/grafana/agent/internal/component/common/kubernetes"
	"github.com/grafana/agent/internal/featuregate"
	"github.com/grafana/agent/internal/flow/logging/level"
	mimirClient "github.com/grafana/agent/internal/mimir/client"
	"github.com/grafana/agent/internal/service/cluster"
	"github.com/grafana/dskit/backoff"
	"github.com/grafana/dskit/instrument"
	promListers "github.com/prometheus-operator/prometheus-operator/pkg/client/listers/monitoring/v1"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	coreListers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	_ "k8s.io/component-base/metrics/prometheus/workqueue"
	controller "sigs.k8s.io/controller-runtime"

	promExternalVersions "github.com/prometheus-operator/prometheus-operator/pkg/client/informers/externalversions"
	promVersioned "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
)

func init() {
	component.Register(component.Registration{
		Name:      "mimir.rules.kubernetes",
		Stability: featuregate.StabilityBeta,
		Args:      Arguments{},
		Exports:   nil,
		Build: func(o component.Options, c component.Arguments) (component.Component, error) {
			return New(o, c.(Arguments))
		},
	})
}

type Component struct {
	log     log.Logger
	opts    component.Options
	args    Arguments
	cluster cluster.Cluster

	mimirClient  mimirClient.Interface
	k8sClient    kubernetes.Interface
	promClient   promVersioned.Interface
	ruleLister   promListers.PrometheusRuleLister
	ruleInformer cache.SharedIndexInformer

	namespaceLister   coreListers.NamespaceLister
	namespaceInformer cache.SharedIndexInformer
	informerStopChan  chan struct{}
	ticker            *time.Ticker

	queue          workqueue.RateLimitingInterface
	configUpdates  chan ConfigUpdate
	clusterUpdates chan struct{}

	namespaceSelector labels.Selector
	ruleSelector      labels.Selector

	currentState commonK8s.RuleGroupsByNamespace

	metrics   *metrics
	healthMut sync.RWMutex
	health    component.Health
}

type metrics struct {
	configUpdatesTotal  prometheus.Counter
	clusterUpdatesTotal prometheus.Counter

	eventsTotal   *prometheus.CounterVec
	eventsFailed  *prometheus.CounterVec
	eventsRetried *prometheus.CounterVec

	mimirClientTiming *prometheus.HistogramVec
}

func (m *metrics) Register(r prometheus.Registerer) error {
	r.MustRegister(
		m.configUpdatesTotal,
		m.eventsTotal,
		m.eventsFailed,
		m.eventsRetried,
		m.mimirClientTiming,
	)
	return nil
}

func newMetrics() *metrics {
	return &metrics{
		configUpdatesTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Subsystem: "mimir_rules",
			Name:      "config_updates_total",
			Help:      "Total number of times the configuration has been updated.",
		}),
		clusterUpdatesTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Subsystem: "mimir_rules",
			Name:      "cluster_updates_total",
			Help:      "Total number of times the cluster has changed triggering an update.",
		}),
		eventsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "mimir_rules",
			Name:      "events_total",
			Help:      "Total number of events processed, partitioned by event type.",
		}, []string{"type"}),
		eventsFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "mimir_rules",
			Name:      "events_failed_total",
			Help:      "Total number of events that failed to be processed, even after retries, partitioned by event type.",
		}, []string{"type"}),
		eventsRetried: prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "mimir_rules",
			Name:      "events_retried_total",
			Help:      "Total number of retries across all events, partitioned by event type.",
		}, []string{"type"}),
		mimirClientTiming: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "mimir_rules",
			Name:      "mimir_client_request_duration_seconds",
			Help:      "Duration of requests to the Mimir API.",
			Buckets:   instrument.DefBuckets,
		}, instrument.HistogramCollectorBuckets),
	}
}

type ConfigUpdate struct {
	args Arguments
	err  chan error
}

var _ component.Component = (*Component)(nil)
var _ component.DebugComponent = (*Component)(nil)
var _ component.HealthComponent = (*Component)(nil)

func New(o component.Options, args Arguments) (*Component, error) {
	metrics := newMetrics()
	err := metrics.Register(o.Registerer)
	if err != nil {
		return nil, fmt.Errorf("registering metrics failed: %w", err)
	}

	data, err := o.GetServiceData(cluster.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get information about cluster: %w", err)
	}
	clusterData := data.(cluster.Cluster)

	c := &Component{
		log:            o.Logger,
		opts:           o,
		cluster:        clusterData,
		args:           args,
		configUpdates:  make(chan ConfigUpdate),
		clusterUpdates: make(chan struct{}),
		ticker:         time.NewTicker(args.SyncInterval),
		metrics:        metrics,
	}

	err = c.init()
	if err != nil {
		return nil, fmt.Errorf("initializing component failed: %w", err)
	}

	return c, nil
}

func (c *Component) Run(ctx context.Context) error {
	c.startupWithRetries(ctx)

	for {
		select {
		case update := <-c.configUpdates:
			c.metrics.configUpdatesTotal.Inc()
			c.shutdown()

			c.args = update.args
			err := c.init()
			if err != nil {
				level.Error(c.log).Log("msg", "updating configuration failed", "err", err)
				c.reportUnhealthy(err)
				update.err <- err
				continue
			}

			err = c.startup(ctx)
			if err != nil {
				level.Error(c.log).Log("msg", "updating configuration failed", "err", err)
				c.reportUnhealthy(err)
				update.err <- err
				continue
			}

			update.err <- nil
		case <-c.clusterUpdates:
			c.shutdown()
			err := c.init()
			if err != nil {
				level.Error(c.log).Log("msg", "updating due to cluster changes failed", "err", err)
				c.reportUnhealthy(err)
				continue
			}

			err = c.startup(ctx)
			if err != nil {
				level.Error(c.log).Log("msg", "updating due to cluster changes failed", "err", err)
				c.reportUnhealthy(err)
				continue
			}
		case <-ctx.Done():
			c.shutdown()
			return nil
		case <-c.ticker.C:
			c.queue.Add(commonK8s.Event{
				Typ: eventTypeSyncMimir,
			})
		}
	}
}

// startupWithRetries calls startup indefinitely until it succeeds.
func (c *Component) startupWithRetries(ctx context.Context) {
	startupBackoff := backoff.New(
		ctx,
		backoff.Config{
			MinBackoff: 1 * time.Second,
			MaxBackoff: 10 * time.Second,
			MaxRetries: 0, // infinite retries
		},
	)
	for {
		if err := c.startup(ctx); err != nil {
			level.Error(c.log).Log("msg", "starting up component failed", "err", err)
			c.reportUnhealthy(err)
		} else {
			break
		}
		startupBackoff.Wait()
	}
}

// startup launches the informers and starts the event loop.
func (c *Component) startup(ctx context.Context) error {
	if !c.isLeader() {
		level.Info(c.log).Log("msg", "not leader, skipping start up!")
		return nil
	}

	c.queue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "mimir.rules.kubernetes")
	c.informerStopChan = make(chan struct{})

	if err := c.startNamespaceInformer(); err != nil {
		return err
	}
	if err := c.startRuleInformer(); err != nil {
		return err
	}
	if err := c.syncMimir(ctx); err != nil {
		return err
	}
	go c.eventLoop(ctx)
	return nil
}

func (c *Component) shutdown() {
	if c.informerStopChan != nil {
		close(c.informerStopChan)
	}
	if c.queue != nil {
		c.queue.ShutDownWithDrain()
	}
}

func (c *Component) Update(newConfig component.Arguments) error {
	errChan := make(chan error)
	c.configUpdates <- ConfigUpdate{
		args: newConfig.(Arguments),
		err:  errChan,
	}
	return <-errChan
}

// NotifyClusterChange implements component.ClusterComponent.
func (c *Component) NotifyClusterChange() {
	c.clusterUpdates <- struct{}{}
}

func (c *Component) isLeader() bool {
	// c.cluster.LookupKey() logic goes here
	return true
}

func (c *Component) init() error {
	level.Info(c.log).Log("msg", "initializing with new configuration")

	// TODO: allow overriding some stuff in RestConfig and k8s client options?
	restConfig, err := controller.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get k8s config: %w", err)
	}

	c.k8sClient, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}

	c.promClient, err = promVersioned.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create prometheus operator client: %w", err)
	}

	httpClient := c.args.HTTPClientConfig.Convert()

	c.mimirClient, err = mimirClient.New(c.log, mimirClient.Config{
		ID:                   c.args.TenantID,
		Address:              c.args.Address,
		UseLegacyRoutes:      c.args.UseLegacyRoutes,
		PrometheusHTTPPrefix: c.args.PrometheusHTTPPrefix,
		HTTPClientConfig:     *httpClient,
	}, c.metrics.mimirClientTiming)
	if err != nil {
		return err
	}

	c.ticker.Reset(c.args.SyncInterval)

	c.namespaceSelector, err = commonK8s.ConvertSelectorToListOptions(c.args.RuleNamespaceSelector)
	if err != nil {
		return err
	}

	c.ruleSelector, err = commonK8s.ConvertSelectorToListOptions(c.args.RuleSelector)
	if err != nil {
		return err
	}

	return nil
}

func (c *Component) startNamespaceInformer() error {
	factory := informers.NewSharedInformerFactoryWithOptions(
		c.k8sClient,
		24*time.Hour,
		informers.WithTweakListOptions(func(lo *metav1.ListOptions) {
			lo.LabelSelector = c.namespaceSelector.String()
		}),
	)

	namespaces := factory.Core().V1().Namespaces()
	c.namespaceLister = namespaces.Lister()
	c.namespaceInformer = namespaces.Informer()
	_, err := c.namespaceInformer.AddEventHandler(commonK8s.NewQueuedEventHandler(c.log, c.queue))
	if err != nil {
		return err
	}

	factory.Start(c.informerStopChan)
	factory.WaitForCacheSync(c.informerStopChan)
	return nil
}

func (c *Component) startRuleInformer() error {
	factory := promExternalVersions.NewSharedInformerFactoryWithOptions(
		c.promClient,
		24*time.Hour,
		promExternalVersions.WithTweakListOptions(func(lo *metav1.ListOptions) {
			lo.LabelSelector = c.ruleSelector.String()
		}),
	)

	promRules := factory.Monitoring().V1().PrometheusRules()
	c.ruleLister = promRules.Lister()
	c.ruleInformer = promRules.Informer()
	_, err := c.ruleInformer.AddEventHandler(commonK8s.NewQueuedEventHandler(c.log, c.queue))
	if err != nil {
		return err
	}

	factory.Start(c.informerStopChan)
	factory.WaitForCacheSync(c.informerStopChan)
	return nil
}
