package rules

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/google/go-cmp/cmp"
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	promListers "github.com/prometheus-operator/prometheus-operator/pkg/client/listers/monitoring/v1"
	"github.com/prometheus/prometheus/model/rulefmt"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	coreListers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/grafana/agent/internal/component/common/kubernetes"
	mimirClient "github.com/grafana/agent/internal/mimir/client"
)

type fakeMimirClient struct {
	rulesMut sync.RWMutex
	rules    map[string][]mimirClient.RuleGroup
}

var _ mimirClient.Interface = &fakeMimirClient{}

func newFakeMimirClient() *fakeMimirClient {
	return &fakeMimirClient{
		rules: make(map[string][]mimirClient.RuleGroup),
	}
}

func (m *fakeMimirClient) CreateRuleGroup(ctx context.Context, namespace string, rule mimirClient.RuleGroup) error {
	m.rulesMut.Lock()
	defer m.rulesMut.Unlock()
	m.deleteLocked(namespace, rule.Name)
	m.rules[namespace] = append(m.rules[namespace], rule)
	return nil
}

func (m *fakeMimirClient) DeleteRuleGroup(ctx context.Context, namespace, group string) error {
	m.rulesMut.Lock()
	defer m.rulesMut.Unlock()
	m.deleteLocked(namespace, group)
	return nil
}

func (m *fakeMimirClient) deleteLocked(namespace, group string) {
	for ns, v := range m.rules {
		if namespace != "" && namespace != ns {
			continue
		}
		for i, g := range v {
			if g.Name == group {
				m.rules[ns] = append(m.rules[ns][:i], m.rules[ns][i+1:]...)

				if len(m.rules[ns]) == 0 {
					delete(m.rules, ns)
				}

				return
			}
		}
	}
}

func (m *fakeMimirClient) ListRules(ctx context.Context, namespace string) (map[string][]mimirClient.RuleGroup, error) {
	m.rulesMut.RLock()
	defer m.rulesMut.RUnlock()
	output := make(map[string][]mimirClient.RuleGroup)
	for ns, v := range m.rules {
		if namespace != "" && namespace != ns {
			continue
		}
		output[ns] = v
	}
	return output, nil
}

func TestEventLoop(t *testing.T) {
	nsIndexer := cache.NewIndexer(
		cache.DeletionHandlingMetaNamespaceKeyFunc,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	nsLister := coreListers.NewNamespaceLister(nsIndexer)

	ruleIndexer := cache.NewIndexer(
		cache.DeletionHandlingMetaNamespaceKeyFunc,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	ruleLister := promListers.NewPrometheusRuleLister(ruleIndexer)

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "namespace",
			UID:  types.UID("33f8860c-bd06-4c0d-a0b1-a114d6b9937b"),
		},
	}

	rule := &v1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "name",
			Namespace: "namespace",
			UID:       types.UID("64aab764-c95e-4ee9-a932-cd63ba57e6cf"),
			Annotations: map[string]string{
				"monitoring.grafana.com/source_tenants": "tenant1, tenant2",
			},
		},
		Spec: v1.PrometheusRuleSpec{
			Groups: []v1.RuleGroup{
				{
					Name: "group",
					Rules: []v1.Rule{
						{
							Alert: "alert",
							Expr:  intstr.FromString("expr"),
						},
					},
				},
			},
		},
	}

	component := Component{
		log:               log.NewLogfmtLogger(os.Stdout),
		queue:             workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		namespaceLister:   nsLister,
		namespaceSelector: labels.Everything(),
		ruleLister:        ruleLister,
		ruleSelector:      labels.Everything(),
		mimirClient:       newFakeMimirClient(),
		args:              Arguments{MimirNameSpacePrefix: "agent"},
		metrics:           newMetrics(),
	}
	eventHandler := kubernetes.NewQueuedEventHandler(component.log, component.queue)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go component.eventLoop(ctx)

	// Add a namespace and rule to kubernetes
	nsIndexer.Add(ns)
	ruleIndexer.Add(rule)
	eventHandler.OnAdd(rule, false)

	expectedRules := map[string][]mimirClient.RuleGroup{
		"agent/namespace/name/64aab764-c95e-4ee9-a932-cd63ba57e6cf": {
			{
				SourceTenants: []string{"tenant1", "tenant2"},
				RuleGroup: rulefmt.RuleGroup{
					Name:     "group",
					Interval: 0,
					Limit:    0,
					Rules: []rulefmt.RuleNode{
						{
							Record: yaml.Node{},
							Alert: yaml.Node{
								Kind:   8,
								Tag:    "!!str",
								Value:  "alert",
								Line:   4,
								Column: 12,
							},
							Expr: yaml.Node{
								Kind:   8,
								Tag:    "!!str",
								Value:  "expr",
								Line:   5,
								Column: 11,
							},
						},
					},
				},
			},
		},
	}

	// Wait for the rule to be added to mimir
	require.Eventually(t, func() bool {
		rules, err := component.mimirClient.ListRules(ctx, "")
		require.NoError(t, err)
		if !cmp.Equal(rules, expectedRules) {
			t.Errorf("rules not equal: %v", cmp.Diff(expectedRules, rules))

			return false
		}

		return true
	}, time.Second, 10*time.Millisecond)
	component.queue.AddRateLimited(kubernetes.Event{Typ: eventTypeSyncMimir})

	// Update the rule in kubernetes
	rule.Spec.Groups[0].Rules = append(rule.Spec.Groups[0].Rules, v1.Rule{
		Alert: "alert2",
		Expr:  intstr.FromString("expr2"),
	})
	ruleIndexer.Update(rule)
	eventHandler.OnUpdate(rule, rule)

	// Wait for the rule to be updated in mimir
	require.Eventually(t, func() bool {
		allRules, err := component.mimirClient.ListRules(ctx, "")
		require.NoError(t, err)
		rules := allRules[mimirNamespaceForRuleCRD("agent", rule)][0].Rules
		return len(rules) == 2
	}, time.Second, 10*time.Millisecond)
	component.queue.AddRateLimited(kubernetes.Event{Typ: eventTypeSyncMimir})

	// Remove the rule from kubernetes
	ruleIndexer.Delete(rule)
	eventHandler.OnDelete(rule)

	// Wait for the rule to be removed from mimir
	require.Eventually(t, func() bool {
		rules, err := component.mimirClient.ListRules(ctx, "")
		require.NoError(t, err)
		return len(rules) == 0
	}, time.Second, 10*time.Millisecond)
}
