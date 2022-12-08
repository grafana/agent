package rules

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/ghodss/yaml"
	"github.com/go-kit/log/level"
	mimirClient "github.com/grafana/agent/pkg/mimir/client"
	"github.com/grafana/dskit/multierror"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/model/rulefmt"
	"k8s.io/client-go/tools/cache"
)

// This type must be hashable, so it is kept simple. The indexer will maintain a
// cache of current state, so this is mostly used for logging.
type Event struct {
	Type      EventType
	ObjectKey string
}

type EventType string

const (
	EventTypeResourceChanged EventType = "resource-changed"
	EventTypeSyncMimir       EventType = "sync-mimir"
)

func (c *Component) OnAdd(obj interface{}) {
	c.publishEvent(obj)
}

func (c *Component) OnUpdate(oldObj, newObj interface{}) {
	c.publishEvent(newObj)
}

func (c *Component) OnDelete(obj interface{}) {
	c.publishEvent(obj)
}

func (c *Component) publishEvent(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		level.Error(c.log).Log("msg", "failed to get key for object", "err", err)
		return
	}

	c.queue.AddRateLimited(Event{
		Type:      EventTypeResourceChanged,
		ObjectKey: key,
	})
}

func (c *Component) eventLoop(ctx context.Context) {
	for {
		event, shutdown := c.queue.Get()
		if shutdown {
			level.Info(c.log).Log("msg", "shutting down event loop")
			return
		}

		evt := event.(Event)
		c.metrics.eventsTotal.WithLabelValues(string(evt.Type)).Inc()
		err := c.processEvent(ctx, evt)

		if err != nil {
			retries := c.queue.NumRequeues(event)
			if retries < 5 {
				c.metrics.eventsRetried.WithLabelValues(string(evt.Type)).Inc()
				c.queue.AddRateLimited(event)
				level.Error(c.log).Log(
					"msg", "failed to process event, will retry",
					"retries", fmt.Sprintf("%d/5", retries),
					"err", err,
				)
				continue
			} else {
				c.metrics.eventsFailed.WithLabelValues(string(evt.Type)).Inc()
				level.Error(c.log).Log(
					"msg", "failed to process event, max retries exceeded",
					"retries", fmt.Sprintf("%d/5", retries),
					"err", err,
				)
			}
		}

		c.queue.Forget(event)
	}
}

func (c *Component) processEvent(ctx context.Context, e Event) error {
	defer c.queue.Done(e)

	switch e.Type {
	case EventTypeResourceChanged:
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
	rulesByNamespace, err := c.mimirClient.ListRules(ctx, "")
	if err != nil {
		level.Error(c.log).Log("msg", "failed to list rules from mimir", "err", err)
		return
	}

	for ns := range rulesByNamespace {
		if !isManagedMimirNamespace(c.args.MimirNameSpacePrefix, ns) {
			delete(rulesByNamespace, ns)
		}
	}

	c.currentState = rulesByNamespace

	return
}

func (c *Component) reconcileState(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	desiredState, err := c.loadStateFromK8s()

	diffs, err := diffRuleState(desiredState, c.currentState)
	if err != nil {
		return err
	}

	errs := multierror.New()
	for ns, diff := range diffs {
		err = c.applyChanges(ctx, ns, diff)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}

	return errs.Err()
}

func (c *Component) loadStateFromK8s() (map[string][]mimirClient.RuleGroup, error) {
	matchedNamespaces, err := c.namespaceLister.List(c.namespaceSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	desiredState := map[string][]mimirClient.RuleGroup{}
	for _, ns := range matchedNamespaces {
		crdState, err := c.ruleLister.PrometheusRules(ns.Name).List(c.ruleSelector)
		if err != nil {
			return nil, fmt.Errorf("failed to list rules: %w", err)
		}

		for _, pr := range crdState {
			mimirNs := mimirNamespaceForRuleCRD(c.args.MimirNameSpacePrefix, pr)

			groups, err := convertCRDRuleGroupToRuleGroup(pr.Spec)
			if err != nil {
				return nil, fmt.Errorf("failed to convert rule group: %w", err)
			}

			desiredState[mimirNs] = groups
		}
	}

	return desiredState, nil
}

func convertCRDRuleGroupToRuleGroup(crd promv1.PrometheusRuleSpec) ([]mimirClient.RuleGroup, error) {
	buf, err := yaml.Marshal(crd)
	if err != nil {
		return nil, err
	}

	groups, errs := rulefmt.Parse(buf)
	if len(errs) > 0 {
		return nil, multierror.New(errs...).Err()
	}

	mimirGroups := make([]mimirClient.RuleGroup, len(groups.Groups))
	for i, g := range groups.Groups {
		mimirGroups[i] = mimirClient.RuleGroup{
			RuleGroup: g,
			// TODO: allow setting remote write configs?
		}
	}

	return mimirGroups, nil
}

func (c *Component) applyChanges(ctx context.Context, namespace string, diffs []RuleGroupDiff) error {
	if len(diffs) == 0 {
		return nil
	}

	for _, diff := range diffs {
		switch diff.Kind {
		case RuleGroupDiffKindAdd:
			err := c.mimirClient.CreateRuleGroup(ctx, namespace, diff.Desired)
			if err != nil {
				return err
			}
			level.Info(c.log).Log("msg", "added rule group", "namespace", namespace, "group", diff.Desired.Name)
		case RuleGroupDiffKindRemove:
			err := c.mimirClient.DeleteRuleGroup(ctx, namespace, diff.Actual.Name)
			if err != nil {
				return err
			}
			level.Info(c.log).Log("msg", "removed rule group", "namespace", namespace, "group", diff.Actual.Name)
		case RuleGroupDiffKindUpdate:
			err := c.mimirClient.CreateRuleGroup(ctx, namespace, diff.Desired)
			if err != nil {
				return err
			}
			level.Info(c.log).Log("msg", "updated rule group", "namespace", namespace, "group", diff.Desired.Name)
		default:
			level.Error(c.log).Log("msg", "unknown rule group diff kind", "kind", diff.Kind)
		}
	}

	// resync mimir state after applying changes
	c.syncMimir(ctx)

	return nil
}

// mimirNamespaceForRuleCRD returns the namespace that the rule CRD should be
// stored in mimir. This function, along with isManagedNamespace, is used to
// determine if a rule CRD is managed by the agent.
func mimirNamespaceForRuleCRD(prefix string, pr *promv1.PrometheusRule) string {
	return fmt.Sprintf("agent/%s/%s/%s", pr.Namespace, pr.Name, pr.UID)
}

// isManagedMimirNamespace returns true if the namespace is managed by the agent.
// Unmanaged namespaces are left as is by the operator.
func isManagedMimirNamespace(prefix, namespace string) bool {
	prefixPart := regexp.QuoteMeta(prefix)
	namespacePart := `.+`
	namePart := `.+`
	uuidPart := `[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12}`
	managedNamespaceRegex := regexp.MustCompile(
		fmt.Sprintf("^%s/%s/%s/%s$", prefixPart, namespacePart, namePart, uuidPart),
	)
	return managedNamespaceRegex.MatchString(namespace)
}
