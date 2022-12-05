package rules

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	mimirClient "github.com/grafana/agent/pkg/mimir/client"
	"github.com/grafana/dskit/crypto/tls"
	promListers "github.com/prometheus-operator/prometheus-operator/pkg/client/listers/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	coreListers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	controller "sigs.k8s.io/controller-runtime"

	promExternalVersions "github.com/prometheus-operator/prometheus-operator/pkg/client/informers/externalversions"
	promVersioned "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
)

func init() {
	component.Register(component.Registration{
		Name:    "mimir.rules",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(o component.Options, c component.Arguments) (component.Component, error) {
			return NewComponent(o, c.(Arguments))
		},
	})
}

type Exports struct {
}

type Component struct {
	log  log.Logger
	opts component.Options
	args Arguments

	mimirClient  *mimirClient.MimirClient
	k8sClient    kubernetes.Interface
	promClient   promVersioned.Interface
	ruleLister   promListers.PrometheusRuleLister
	ruleInformer cache.SharedIndexInformer

	namespaceLister   coreListers.NamespaceLister
	namespaceInformer cache.SharedIndexInformer
	informerStopChan  chan struct{}
	ticker            *time.Ticker

	queue         workqueue.RateLimitingInterface
	configUpdates chan ConfigUpdate

	namespaceSelector labels.Selector
	ruleSelector      labels.Selector

	currentState []mimirClient.RuleGroup
}

type ConfigUpdate struct {
	args Arguments
	err  chan error
}

var _ component.Component = (*Component)(nil)

func NewComponent(o component.Options, c Arguments) (*Component, error) {
	setDefaultArguments(&c)
	return &Component{
		log:           o.Logger,
		opts:          o,
		args:          c,
		configUpdates: make(chan ConfigUpdate),
		ticker:        time.NewTicker(c.SyncInterval),
	}, nil
}

func (c *Component) Run(ctx context.Context) error {
	err := c.startup(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case update := <-c.configUpdates:
			c.shutdown()

			c.args = update.args
			err := c.startup(ctx)
			update.err <- err
			if err != nil {
				return err
			}
		case <-ctx.Done():
			c.shutdown()
			return nil
		case <-c.ticker.C:
			c.queue.Add(Event{
				Type: EventTypeSyncMimir,
			})
		}
	}
}

func (c *Component) startup(ctx context.Context) error {
	err := c.init()
	if err != nil {
		return err
	}

	c.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	c.informerStopChan = make(chan struct{})

	c.startNamespaceInformer()
	c.startRuleInformer()
	c.syncMimir(ctx)
	go c.eventLoop(ctx)
	return nil
}

func (c *Component) shutdown() {
	close(c.informerStopChan)
	c.queue.ShutDownWithDrain()
}

func (c *Component) Update(newConfig component.Arguments) error {
	errChan := make(chan error)
	c.configUpdates <- ConfigUpdate{
		args: newConfig.(Arguments),
		err:  errChan,
	}
	return <-errChan
}

func (c *Component) init() error {
	level.Info(c.log).Log("msg", "initializing with new configuration")

	setDefaultArguments(&c.args)

	// TODO: allow overriding some stuff in RestConfig and k8s client options?
	restConfig := controller.GetConfigOrDie()

	var err error
	c.k8sClient, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}

	c.promClient, err = promVersioned.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create prometheus operator client: %w", err)
	}

	c.mimirClient, err = mimirClient.New(mimirClient.Config{
		User:    c.args.ClientParams.User,
		Key:     string(c.args.ClientParams.Key),
		Address: c.args.ClientParams.Address,
		ID:      c.args.ClientParams.ID,
		TLS: tls.ClientConfig{
			CertPath:           c.args.ClientParams.TLS.CertPath,
			KeyPath:            c.args.ClientParams.TLS.KeyPath,
			CAPath:             c.args.ClientParams.TLS.CAPath,
			ServerName:         c.args.ClientParams.TLS.ServerName,
			InsecureSkipVerify: c.args.ClientParams.TLS.InsecureSkipVerify,
			CipherSuites:       c.args.ClientParams.TLS.CipherSuites,
			MinVersion:         c.args.ClientParams.TLS.MinVersion,
		},
		UseLegacyRoutes: c.args.ClientParams.UseLegacyRoutes,
		AuthToken:       string(c.args.ClientParams.AuthToken),
	})
	if err != nil {
		return err
	}

	c.ticker.Reset(c.args.SyncInterval)

	c.namespaceSelector, err = convertSelectorToListOptions(c.args.RuleNamespaceSelector)
	if err != nil {
		return err
	}

	c.ruleSelector, err = convertSelectorToListOptions(c.args.RuleSelector)
	if err != nil {
		return err
	}

	return nil
}

func convertSelectorToListOptions(selector LabelSelector) (labels.Selector, error) {
	matchExpressions := []metav1.LabelSelectorRequirement{}

	for _, me := range selector.MatchExpressions {
		matchExpressions = append(matchExpressions, metav1.LabelSelectorRequirement{
			Key:      me.Key,
			Operator: metav1.LabelSelectorOperator(me.Operator),
			Values:   me.Values,
		})
	}

	return metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels:      selector.MatchLabels,
		MatchExpressions: matchExpressions,
	})
}

func (c *Component) startNamespaceInformer() {
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
	c.namespaceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				level.Error(c.log).Log("msg", "failed to get key from object", "err", err)
				return
			}

			c.queue.AddRateLimited(Event{
				Type:      EventTypeAddNamespace,
				ObjectKey: key,
			})
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newKey, err := cache.MetaNamespaceKeyFunc(newObj)
			if err != nil {
				level.Error(c.log).Log("msg", "failed to get key from object", "err", err)
				return
			}

			c.queue.AddRateLimited(Event{
				Type:      EventTypeUpdateNamespace,
				ObjectKey: newKey,
			})
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				level.Error(c.log).Log("msg", "failed to get key from object", "err", err)
				return
			}

			c.queue.AddRateLimited(Event{
				Type:      EventTypeDeleteNamespace,
				ObjectKey: key,
			})
		},
	})

	factory.Start(c.informerStopChan)
	factory.WaitForCacheSync(c.informerStopChan)
}

func (c *Component) startRuleInformer() {
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
	c.ruleInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				level.Error(c.log).Log("msg", "failed to get key from object", "err", err)
				return
			}

			c.queue.AddRateLimited(Event{
				Type:      EventTypeAddRule,
				ObjectKey: key,
			})
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newKey, err := cache.MetaNamespaceKeyFunc(newObj)
			if err != nil {
				level.Error(c.log).Log("msg", "failed to get key from object", "err", err)
				return
			}

			c.queue.AddRateLimited(Event{
				Type:      EventTypeUpdateRule,
				ObjectKey: newKey,
			})
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				level.Error(c.log).Log("msg", "failed to get key from object", "err", err)
				return
			}

			c.queue.AddRateLimited(Event{
				Type:      EventTypeDeleteRule,
				ObjectKey: key,
			})
		},
	})

	factory.Start(c.informerStopChan)
	factory.WaitForCacheSync(c.informerStopChan)
}

func setDefaultArguments(args *Arguments) {
	if args.SyncInterval == 0 {
		args.SyncInterval = 30 * time.Second
	}
}
