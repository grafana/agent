package rules

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/rivertypes"
	mimirClient "github.com/grafana/agent/pkg/mimir/client"
	"github.com/grafana/dskit/crypto/tls"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	controller "sigs.k8s.io/controller-runtime"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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

	mimirClient *mimirClient.MimirClient
	k8sClient   k8sClient.Client
	ticker      *time.Ticker

	namespaceSelector labels.Selector
	ruleSelector      labels.Selector
}

var _ component.Component = (*Component)(nil)
var _ reconcile.Reconciler = (*Component)(nil)

func NewComponent(o component.Options, c Arguments) (*Component, error) {
	return &Component{
		log:  o.Logger,
		opts: o,
		args: c,
	}, nil
}

func (c *Component) Run(ctx context.Context) error {
	err := c.init()
	if err != nil {
		return err
	}

	c.start(ctx)

	return nil
}

func (c *Component) Update(newConfig component.Arguments) error {
	c.args = newConfig.(Arguments)
	return c.init()
}

func (c *Component) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func (c *Component) init() error {
	if c.args.SyncInterval == 0 {
		c.args.SyncInterval = 30 * time.Second
	}

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

	c.ticker = time.NewTicker(c.args.SyncInterval)

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

func (c *Component) start(ctx context.Context) {
	err := c.syncRules(ctx)
	if err != nil {
		level.Error(c.log).Log("msg", "failed to sync rules", "err", err)
	}

	for {
		select {
		case <-c.ticker.C:
			err := c.syncRules(ctx)
			if err != nil {
				level.Error(c.log).Log("msg", "failed to sync rules", "err", err)
			}
		case <-ctx.Done():
			level.Info(c.log).Log("msg", "shutting down")
			return
		}
	}
}

func (c *Component) syncRules(ctx context.Context) error {
	level.Info(c.log).Log("msg", "syncing rules")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	desiredState, err := c.discoverRuleCRDs(ctx)
	if err != nil {
		return err
	}

	actualState, err := c.loadActiveRules(ctx)
	if err != nil {
		return err
	}

	diffs := c.diffRuleStates(desiredState, actualState)

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

func (c *Component) discoverRuleCRDs(ctx context.Context) ([]*promv1.PrometheusRule, error) {
	// List namespaces
	var namespaces corev1.NamespaceList
	err := c.k8sClient.List(ctx, &namespaces, &k8sClient.ListOptions{
		LabelSelector: c.namespaceSelector,
	})
	if err != nil {
		return nil, err
	}

	var crds []*promv1.PrometheusRule
	// List rules in each namespace
	for _, namespace := range namespaces.Items {
		var crdList promv1.PrometheusRuleList
		err := c.k8sClient.List(ctx, &crdList, &k8sClient.ListOptions{
			LabelSelector: c.ruleSelector,
			Namespace:     namespace.Name,
		})
		if err != nil {
			return nil, err
		}

		crds = append(crds, crdList.Items...)
	}
	return crds, nil
}

func (c *Component) loadActiveRules(ctx context.Context) ([]mimirClient.RuleGroup, error) {
	rulesByNamespace, err := c.mimirClient.ListRules(ctx, c.args.MimirRuleNamespace)
	if err != nil {
		return nil, err
	}

	return rulesByNamespace[c.args.MimirRuleNamespace], nil
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
	Desired promv1.RuleGroup
}

func (c *Component) diffRuleStates(desired []*promv1.PrometheusRule, actual []mimirClient.RuleGroup) []RuleGroupDiff {
	var diff []RuleGroupDiff

	seenGroups := map[string]bool{}

	for _, desiredRule := range desired {
	desiredGroups:
		for _, desiredRuleGroup := range desiredRule.Spec.Groups {
			for _, actualRuleGroup := range actual {
				if desiredRuleGroup.Name == actualRuleGroup.Name {
					diff = append(diff, RuleGroupDiff{
						Kind:    RuleGroupDiffKindUpdate,
						Actual:  actualRuleGroup,
						Desired: desiredRuleGroup,
					})
					continue desiredGroups
				}
			}

			diff = append(diff, RuleGroupDiff{
				Kind:    RuleGroupDiffKindAdd,
				Desired: desiredRuleGroup,
			})
		}
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

	return diff
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
		case RuleGroupDiffKindRemove:
			level.Info(c.log).Log("msg", "removing rule group", "group", diff.Actual.Name)
		case RuleGroupDiffKindUpdate:
			level.Info(c.log).Log("msg", "updating rule group", "group", diff.Desired.Name)
		default:
			level.Error(c.log).Log("msg", "unknown rule group diff kind", "kind", diff.Kind)
		}
	}

	return nil
}
