package rules

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/ghodss/yaml"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/rivertypes"
	mimirClient "github.com/grafana/agent/pkg/mimir/client"
	"github.com/grafana/dskit/crypto/tls"
	"github.com/grafana/dskit/multierror"
	"github.com/pkg/errors"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	promListers "github.com/prometheus-operator/prometheus-operator/pkg/client/listers/monitoring/v1"
	"github.com/prometheus/prometheus/model/rulefmt"
	yamlv3 "gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	controller "sigs.k8s.io/controller-runtime"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"

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

type Arguments struct {
	ClientParams       ClientArguments `river:"client,block"`
	SyncInterval       time.Duration   `river:"sync_interval,attr,optional"`
	MimirRuleNamespace string          `river:"mimir_rule_namespace,attr"`

	RuleSelector          LabelSelector `river:"rule_selector,block,optional"`
	RuleNamespaceSelector LabelSelector `river:"rule_namespace_selector,block,optional"`
}

type LabelSelector struct {
	MatchLabels      map[string]string `river:"match_labels,attr"`
	MatchExpressions []MatchExpression `river:"match_expressions,attr"`
}

type MatchExpression struct {
	Key      string   `river:"key,attr"`
	Operator string   `river:"operator,attr"`
	Values   []string `river:"values,attr"`
}

type ClientArguments struct {
	User            string            `river:"user,attr,optional"`
	Key             rivertypes.Secret `river:"key,attr,optional"`
	Address         string            `river:"address,attr"`
	ID              string            `river:"id,attr,optional"`
	TLS             TLSArguments      `river:"tls,block,optional"`
	UseLegacyRoutes bool              `river:"use_legacy_routes,attr,optional"`
	AuthToken       rivertypes.Secret `river:"auth_token,attr,optional"`
}

type TLSArguments struct {
	CertPath           string `river:"tls_cert_path,attr,optional"`
	KeyPath            string `river:"tls_key_path,attr,optional"`
	CAPath             string `river:"tls_ca_path,attr,optional"`
	ServerName         string `river:"tls_server_name,attr,optional"`
	InsecureSkipVerify bool   `river:"tls_insecure_skip_verify,attr,optional"`
	CipherSuites       string `river:"tls_cipher_suites,attr,optional"`
	MinVersion         string `river:"tls_min_version,attr,optional"`
}

type Exports struct {
}

type Component struct {
	log  log.Logger
	opts component.Options
	args Arguments

	mimirClient      *mimirClient.MimirClient
	k8sClient        k8sClient.Client
	promClient       promVersioned.Interface
	ruleLister       promListers.PrometheusRuleLister
	ruleInformer     cache.SharedIndexInformer
	informerStopChan chan struct{}
	ticker           *time.Ticker

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

func setDefaultArguments(args *Arguments) {
	if args.SyncInterval == 0 {
		args.SyncInterval = 30 * time.Second
	}
}

func (c *Component) init() error {
	level.Info(c.log).Log("msg", "initializing with new configuration")

	setDefaultArguments(&c.args)

	// TODO: allow overriding some stuff in RestConfig and k8s client options?
	restConfig := controller.GetConfigOrDie()

	scheme := runtime.NewScheme()
	err := corev1.AddToScheme(scheme)
	if err != nil {
		return fmt.Errorf("failed to add prometheus operator scheme: %w", err)
	}
	err = promv1.AddToScheme(scheme)
	if err != nil {
		return fmt.Errorf("failed to add prometheus operator scheme: %w", err)
	}

	c.k8sClient, err = k8sClient.New(restConfig, k8sClient.Options{
		Scheme: scheme,
	})
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

func (c *Component) reconcileState(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	crdState, err := c.ruleLister.List(c.ruleSelector)
	if err != nil {
		return fmt.Errorf("failed to list rules: %w", err)
	}

	desiredState := []rulefmt.RuleGroup{}
	for _, pr := range crdState {
		groups, err := convertCRDRuleGroupToRuleGroup(pr.Spec)
		if err != nil {
			return fmt.Errorf("failed to convert rule group: %w", err)
		}

		desiredState = append(desiredState, groups.Groups...)
	}

	diffs, err := c.diffRuleStates(desiredState, c.currentState)
	if err != nil {
		return err
	}

	return c.applyChanges(ctx, diffs)
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

type RuleGroupDiffKind string

const (
	RuleGroupDiffKindAdd    RuleGroupDiffKind = "add"
	RuleGroupDiffKindRemove RuleGroupDiffKind = "remove"
	RuleGroupDiffKindUpdate RuleGroupDiffKind = "update"
)

type RuleGroupDiff struct {
	Kind    RuleGroupDiffKind
	Actual  mimirClient.RuleGroup
	Desired mimirClient.RuleGroup
}

func (c *Component) diffRuleStates(desired []rulefmt.RuleGroup, actual []mimirClient.RuleGroup) ([]RuleGroupDiff, error) {
	var diff []RuleGroupDiff

	seenGroups := map[string]bool{}

desiredGroups:
	for _, desiredRuleGroup := range desired {
		mimirRuleGroup := mimirClient.RuleGroup{
			RuleGroup: desiredRuleGroup,
			// TODO: allow setting the remote write configs?
			// RWConfigs: ,
		}

		seenGroups[desiredRuleGroup.Name] = true

		for _, actualRuleGroup := range actual {
			if desiredRuleGroup.Name == actualRuleGroup.Name {
				if equalRuleGroups(desiredRuleGroup, actualRuleGroup.RuleGroup) {
					continue desiredGroups
				}

				// TODO: check if the rules are the same
				diff = append(diff, RuleGroupDiff{
					Kind:    RuleGroupDiffKindUpdate,
					Actual:  actualRuleGroup,
					Desired: mimirRuleGroup,
				})
				continue desiredGroups
			}
		}

		diff = append(diff, RuleGroupDiff{
			Kind:    RuleGroupDiffKindAdd,
			Desired: mimirRuleGroup,
		})
	}

	for _, actualRuleGroup := range actual {
		if seenGroups[actualRuleGroup.Name] {
			continue
		}

		diff = append(diff, RuleGroupDiff{
			Kind:   RuleGroupDiffKindRemove,
			Actual: actualRuleGroup,
		})
	}

	return diff, nil
}

func (c *Component) applyChanges(ctx context.Context, diffs []RuleGroupDiff) error {
	if len(diffs) == 0 {
		return nil
	}

	level.Info(c.log).Log("msg", "applying rule changes", "num_changes", len(diffs))

	for _, diff := range diffs {
		switch diff.Kind {
		case RuleGroupDiffKindAdd:
			level.Info(c.log).Log("msg", "adding rule group", "group", diff.Desired.Name)
			err := c.mimirClient.CreateRuleGroup(ctx, c.args.MimirRuleNamespace, diff.Desired)
			if err != nil {
				return err
			}
		case RuleGroupDiffKindRemove:
			level.Info(c.log).Log("msg", "removing rule group", "group", diff.Actual.Name)
			err := c.mimirClient.DeleteRuleGroup(ctx, c.args.MimirRuleNamespace, diff.Actual.Name)
			if err != nil {
				return err
			}
		case RuleGroupDiffKindUpdate:
			level.Info(c.log).Log("msg", "updating rule group", "group", diff.Desired.Name)
			err := c.mimirClient.CreateRuleGroup(ctx, c.args.MimirRuleNamespace, diff.Desired)
			if err != nil {
				return err
			}
		default:
			level.Error(c.log).Log("msg", "unknown rule group diff kind", "kind", diff.Kind)
		}
	}

	c.syncMimir(ctx)

	return nil
}

func convertCRDRuleGroupToRuleGroup(crd promv1.PrometheusRuleSpec) (*rulefmt.RuleGroups, error) {
	buf, err := yaml.Marshal(crd)
	if err != nil {
		return &rulefmt.RuleGroups{}, err
	}

	groups, errs := rulefmt.Parse(buf)
	if len(errs) > 0 {
		return &rulefmt.RuleGroups{}, multierror.New(errs...).Err()
	}

	return groups, nil
}

func (c *Component) startRuleInformer() {
	c.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	factory := promExternalVersions.NewSharedInformerFactory(c.promClient, 24*time.Hour)

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
				Type:     EventTypeAddRule,
				NewRules: key,
			})
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldKey, err := cache.MetaNamespaceKeyFunc(oldObj)
			if err != nil {
				level.Error(c.log).Log("msg", "failed to get key from object", "err", err)
				return
			}

			newKey, err := cache.MetaNamespaceKeyFunc(newObj)
			if err != nil {
				level.Error(c.log).Log("msg", "failed to get key from object", "err", err)
				return
			}

			c.queue.AddRateLimited(Event{
				Type:     EventTypeUpdateRule,
				NewRules: newKey,
				OldRules: oldKey,
			})
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				level.Error(c.log).Log("msg", "failed to get key from object", "err", err)
				return
			}

			c.queue.AddRateLimited(Event{
				Type:     EventTypeDeleteRule,
				OldRules: key,
			})
		},
	})

	c.informerStopChan = make(chan struct{})
	factory.Start(c.informerStopChan)
	factory.WaitForCacheSync(c.informerStopChan)
}

func (c *Component) eventLoop(ctx context.Context) {
	for {
		event, shutdown := c.queue.Get()
		if shutdown {
			level.Info(c.log).Log("msg", "shutting down event loop")
			return
		}

		evt := event.(Event)
		err := c.processEvent(ctx, evt)

		if err != nil {
			// TODO: retry limits?
			level.Error(c.log).Log("msg", "failed to process event", "err", err)
			// c.queue.AddRateLimited(event)
		} else {
			c.queue.Forget(event)
			c.queue.Done(event)
		}
	}
}

func (c *Component) getRuleGroupsFromKey(key string) (*rulefmt.RuleGroups, error) {
	obj, _, err := c.ruleInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rule from informer")
	}

	groups, err := convertCRDRuleGroupToRuleGroup(obj.(*promv1.PrometheusRule).Spec)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert CRD rule group to rule group")
	}

	return groups, nil
}

func (c *Component) processEvent(ctx context.Context, e Event) error {
	switch e.Type {
	case EventTypeAddRule:
		level.Info(c.log).Log("msg", "processing add rule event", "key", e.NewRules)
	case EventTypeUpdateRule:
		level.Info(c.log).Log("msg", "processing update rule event", "key", e.NewRules)
	case EventTypeDeleteRule:
		level.Info(c.log).Log("msg", "processing delete rule event", "key", e.OldRules)
	case EventTypeAddNamespace:
	case EventTypeDeleteNamespace:
	case EventTypeUpdateNamespace:
	case EventTypeSyncMimir:
		level.Debug(c.log).Log("msg", "syncing current state from ruler")
		c.syncMimir(ctx)
	default:
		return fmt.Errorf("unknown event type: %s", e.Type)
	}

	return c.reconcileState(ctx)
}

func (c *Component) syncMimir(ctx context.Context) {
	rulesByNamespace, err := c.mimirClient.ListRules(ctx, c.args.MimirRuleNamespace)
	if err != nil {
		level.Error(c.log).Log("msg", "failed to list rules from mimir", "err", err)
		return
	}

	c.currentState = rulesByNamespace[c.args.MimirRuleNamespace]
}

// This type must be hashable, so it is kept simple. The indexer will maintain a
// cache of current state, so this is only used for logging.
type Event struct {
	Type EventType

	NewRules string
	OldRules string

	NewNamespace string
	OldNamespace string
}

type EventType string

const (
	EventTypeAddRule    EventType = "add-rule"
	EventTypeUpdateRule EventType = "update-rule"
	EventTypeDeleteRule EventType = "delete-rule"

	EventTypeAddNamespace    EventType = "add-namespace"
	EventTypeUpdateNamespace EventType = "update-namespace"
	EventTypeDeleteNamespace EventType = "delete-namespace"

	EventTypeSyncMimir EventType = "sync-mimir"
)

func equalRuleGroups(a, b rulefmt.RuleGroup) bool {
	aBuf, err := yamlv3.Marshal(a)
	if err != nil {
		return false
	}
	bBuf, err := yamlv3.Marshal(b)
	if err != nil {
		return false
	}

	if !bytes.Equal(aBuf, bBuf) {

		fmt.Println("----")
		fmt.Println(string(aBuf))
		fmt.Println("----")
		fmt.Println(string(bBuf))

		return false
	}

	return bytes.Equal(aBuf, bBuf)
}
