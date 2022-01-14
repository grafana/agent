package operator

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	grafana "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/grafana/agent/pkg/operator/hierarchy"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/agent/pkg/util/k8s"
	"github.com/grafana/agent/pkg/util/structwalk"
	prom "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// Test_buildHierarchy checks that an entire resource hierarchy can be
// discovered.
func Test_buildHierarchy(t *testing.T) {
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	l := util.TestLogger(t)
	cluster := NewTestCluster(ctx, t, l)
	cli := cluster.Client()

	resources := k8s.NewResourceSet(l, cluster)
	defer resources.Stop()
	require.NoError(t, resources.AddFile(ctx, "./testdata/test-resource-hierarchy.yaml"))

	// Get root resource
	var root grafana.GrafanaAgent
	err := cli.Get(ctx, client.ObjectKey{Namespace: "default", Name: "grafana-agent-example"}, &root)
	require.NoError(t, err)

	deployment, watchers, err := buildHierarchy(ctx, l, cli, &root)
	require.NoError(t, err)

	// Check resources in hierarchy
	{
		expectedResources := []string{
			"GrafanaAgent/grafana-agent-example",
			"MetricsInstance/primary",
			"LogsInstance/primary",
			"PodMonitor/grafana-agents",
			"PodLogs/grafana-agents",
		}
		var gotResources []string
		structwalk.Walk(&resourceWalker{
			onResource: func(c client.Object) {
				gvk, _ := apiutil.GVKForObject(c, cli.Scheme())

				key := fmt.Sprintf("%s/%s", gvk.Kind, c.GetName())
				gotResources = append(gotResources, key)
			},
		}, deployment)

		require.ElementsMatch(t, expectedResources, gotResources)
	}

	// Check secrets
	{
		expectedSecrets := []string{
			"/secrets/default/prometheus-fake-credentials/fakeUsername",
			"/secrets/default/prometheus-fake-credentials/fakePassword",
		}
		var actualSecrets []string
		for key := range deployment.Secrets {
			actualSecrets = append(actualSecrets, string(key))
		}

		require.ElementsMatch(t, expectedSecrets, actualSecrets)
	}

	// Check configured watchers
	{
		expectedWatchers := []hierarchy.Watcher{
			{
				Object: &grafana.MetricsInstance{},
				Owner:  client.ObjectKey{Namespace: "default", Name: "grafana-agent-example"},
				Selector: &hierarchy.LabelsSelector{
					NamespaceName: "default",
					Labels:        labels.SelectorFromSet(labels.Set{"agent": "grafana-agent-example"}),
				},
			},
			{
				Object: &grafana.LogsInstance{},
				Owner:  client.ObjectKey{Namespace: "default", Name: "grafana-agent-example"},
				Selector: &hierarchy.LabelsSelector{
					NamespaceName: "default",
					Labels:        labels.SelectorFromSet(labels.Set{"agent": "grafana-agent-example"}),
				},
			},
			{
				Object: &prom.ServiceMonitor{},
				Owner:  client.ObjectKey{Namespace: "default", Name: "grafana-agent-example"},
				Selector: &hierarchy.LabelsSelector{
					NamespaceName:   "default",
					NamespaceLabels: labels.Everything(),
					Labels:          labels.SelectorFromSet(labels.Set{"instance": "primary"}),
				},
			},
			{
				Object: &prom.PodMonitor{},
				Owner:  client.ObjectKey{Namespace: "default", Name: "grafana-agent-example"},
				Selector: &hierarchy.LabelsSelector{
					NamespaceName:   "default",
					NamespaceLabels: labels.Everything(),
					Labels:          labels.SelectorFromSet(labels.Set{"instance": "primary"}),
				},
			},
			{
				Object: &prom.Probe{},
				Owner:  client.ObjectKey{Namespace: "default", Name: "grafana-agent-example"},
				Selector: &hierarchy.LabelsSelector{
					NamespaceName: "default",
					Labels:        labels.Nothing(),
				},
			},
			{
				Object: &grafana.PodLogs{},
				Owner:  client.ObjectKey{Namespace: "default", Name: "grafana-agent-example"},
				Selector: &hierarchy.LabelsSelector{
					NamespaceName:   "default",
					NamespaceLabels: labels.Everything(),
					Labels:          labels.SelectorFromSet(labels.Set{"instance": "primary"}),
				},
			},
			{
				Object: &v1.Secret{},
				Owner:  client.ObjectKey{Namespace: "default", Name: "grafana-agent-example"},
				Selector: &hierarchy.KeySelector{
					Namespace: "default",
					Name:      "prometheus-fake-credentials",
				},
			},
		}
		require.ElementsMatch(t, expectedWatchers, watchers)
	}
}

type resourceWalker struct {
	onResource func(c client.Object)
}

func (w *resourceWalker) Visit(v interface{}) (next structwalk.Visitor) {
	if v == nil {
		return nil
	}
	if obj, ok := v.(client.Object); ok {
		w.onResource(obj)
	}
	return w
}
