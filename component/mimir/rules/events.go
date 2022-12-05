package rules

import (
	"context"
	"fmt"
	"time"

	"github.com/ghodss/yaml"
	"github.com/go-kit/log/level"
	"github.com/grafana/dskit/multierror"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/model/rulefmt"
)

// This type must be hashable, so it is kept simple. The indexer will maintain a
// cache of current state, so this is mostly used for logging.
type Event struct {
	Type      EventType
	ObjectKey string
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
func (c *Component) processEvent(ctx context.Context, e Event) error {
	switch e.Type {
	case EventTypeAddRule, EventTypeUpdateRule, EventTypeDeleteRule,
		EventTypeAddNamespace, EventTypeUpdateNamespace, EventTypeDeleteNamespace:
		level.Info(c.log).Log("msg", "processing event", "type", e.Type, "key", e.ObjectKey)
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

func (c *Component) reconcileState(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	desiredState, err := c.loadStateFromK8s()

	diffs, err := diffRuleStates(desiredState, c.currentState)
	if err != nil {
		return err
	}

	return c.applyChanges(ctx, diffs)
}

func (c *Component) loadStateFromK8s() ([]rulefmt.RuleGroup, error) {
	matchedNamespaces, err := c.namespaceLister.List(c.namespaceSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	desiredState := []rulefmt.RuleGroup{}
	for _, ns := range matchedNamespaces {
		crdState, err := c.ruleLister.PrometheusRules(ns.Name).List(c.ruleSelector)
		if err != nil {
			return nil, fmt.Errorf("failed to list rules: %w", err)
		}

		for _, pr := range crdState {
			groups, err := convertCRDRuleGroupToRuleGroup(pr.Spec)
			if err != nil {
				return nil, fmt.Errorf("failed to convert rule group: %w", err)
			}

			desiredState = append(desiredState, groups.Groups...)
		}
	}

	return desiredState, nil
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
