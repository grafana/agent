package operator

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	grafana "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/grafana/agent/pkg/operator/assets"
	"github.com/grafana/agent/pkg/operator/config"
	prom "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type deploymentBuilder struct {
	client.Client

	Logger  log.Logger
	Config  *Config
	Agent   *grafana.GrafanaAgent
	Secrets assets.SecretStore

	// ResourceSelectors is filled as objects are found and can be used to
	// trigger future reconciles.
	ResourceSelectors map[secondaryResource][]resourceSelector
}

func (b *deploymentBuilder) Build(ctx context.Context) (config.Deployment, error) {
	metricsInstanceSel, err := b.buildResourceSelector(b.Agent.MetricsInstanceSelector())
	if err != nil {
		return config.Deployment{}, fmt.Errorf("failed to build MetricsInstance selector: %w", err)
	}
	b.addSelector(resourcePromInstance, metricsInstanceSel)

	rootMetricInstances, err := b.getMetricsInstances(ctx, metricsInstanceSel)
	if err != nil {
		return config.Deployment{}, err
	}
	metricInstances := make([]config.MetricsInstance, 0, len(rootMetricInstances))

	for _, inst := range rootMetricInstances {
		// Get resource selectors for ServiceMonitors, PodMonitors, Probes
		var (
			sMonSel, pMonSel, probeSel resourceSelector
		)
		getters := []struct {
			name string
			rsel *resourceSelector
			osel grafana.ObjectSelector
		}{
			{"ServiceMonitor", &sMonSel, inst.ServiceMonitorSelector()},
			{"PodMonitor", &pMonSel, inst.PodMonitorSelector()},
			{"Probe", &probeSel, inst.ProbeSelector()},
		}
		for _, g := range getters {
			var err error
			*g.rsel, err = b.buildResourceSelector(g.osel)
			if err != nil {
				return config.Deployment{}, fmt.Errorf("failed to build %s selector: %w", g.name, err)
			}
		}

		// Use resource selectors to look up objects
		sMons, err := b.getServiceMonitors(ctx, sMonSel)
		if err != nil {
			return config.Deployment{}, fmt.Errorf("unable to fetch ServiceMonitors: %w", err)
		}
		pMons, err := b.getPodMonitors(ctx, pMonSel)
		if err != nil {
			return config.Deployment{}, fmt.Errorf("unable to fetch PodMonitors: %w", err)
		}
		probes, err := b.getProbes(ctx, probeSel)
		if err != nil {
			return config.Deployment{}, fmt.Errorf("unable to fetch Probes: %w", err)
		}

		metricInstances = append(metricInstances, config.MetricsInstance{
			Instance:        inst,
			ServiceMonitors: sMons,
			PodMonitors:     pMons,
			Probes:          probes,
		})

		b.addSelector(resourceServiceMonitor, sMonSel)
		b.addSelector(resourcePodMonitor, pMonSel)
		b.addSelector(resourceProbe, probeSel)
	}

	logsInstanceSel, err := b.buildResourceSelector(b.Agent.LogsInstanceSelector())
	if err != nil {
		return config.Deployment{}, fmt.Errorf("failed to build LogsInstance selector: %w", err)
	}
	b.addSelector(resourceLogsInstance, logsInstanceSel)

	rootLogsInstances, err := b.getLogsInstances(ctx, logsInstanceSel)
	if err != nil {
		return config.Deployment{}, err
	}
	logsInstances := make([]config.LogInstance, 0, len(rootLogsInstances))

	for _, inst := range rootLogsInstances {
		podLogsSel, err := b.buildResourceSelector(inst.PodLogsInstanceSelector())
		if err != nil {
			return config.Deployment{}, fmt.Errorf("failed to build PodLogs selector: %w", err)
		}
		podLogs, err := b.getPodLogs(ctx, podLogsSel)
		if err != nil {
			return config.Deployment{}, fmt.Errorf("unable to fetch PodLogs: %w", err)
		}

		logsInstances = append(logsInstances, config.LogInstance{
			Instance: inst,
			PodLogs:  podLogs,
		})

		b.addSelector(resourcePodLogs, podLogsSel)
	}

	return config.Deployment{
		Agent:   b.Agent,
		Metrics: metricInstances,
		Logs:    logsInstances,
	}, nil
}

// buildResourceSelector builds a selector for discovering objects with a
// namespace selector and an object selector. If the namespace
// selector, it will default to finding everything in the parent
// namespace.
func (b *deploymentBuilder) buildResourceSelector(sel grafana.ObjectSelector) (resourceSelector, error) {
	var namespaceFilter string
	if sel.NamespaceSelector == nil {
		// When there's no namespaceSelector defined, default to looking in the
		// current namespace and matching everything within that namespace.
		namespaceFilter = sel.ParentNamespace
		sel.NamespaceSelector = &metav1.LabelSelector{}
	}

	nsLabels, err := metav1.LabelSelectorAsSelector(sel.NamespaceSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to convert object namespace selector into label selector: %w", err)
	}
	objectLabels, err := metav1.LabelSelectorAsSelector(sel.Labels)
	if err != nil {
		return nil, fmt.Errorf("failed to convert object selector into label selector: %w", err)
	}

	var ss []resourceSelector

	if namespaceFilter != "" {
		ss = append(ss, &namespaceSelector{Namespace: namespaceFilter})
	}
	ss = append(ss, &namespaceLabelSelector{Selector: nsLabels})
	ss = append(ss, &labelSelector{Selector: objectLabels})
	return &multiSelector{Selectors: ss}, nil
}

// addSelector registers the given selector for the resource to use for update
// tracking.
func (b *deploymentBuilder) addSelector(res secondaryResource, sel resourceSelector) {
	b.ResourceSelectors[res] = append(b.ResourceSelectors[res], sel)
}

func (b *deploymentBuilder) getMetricsInstances(ctx context.Context, sel resourceSelector) ([]*grafana.MetricsInstance, error) {
	var list grafana.MetricsInstanceList
	if err := b.list(ctx, &list, sel); err != nil {
		return nil, fmt.Errorf("unable to discover MetricsInstance: %w", err)
	}
	return list.Items, nil
}

// list finds all objects for sel and fills them into list.
func (b *deploymentBuilder) list(
	ctx context.Context,
	list client.ObjectList,
	sel resourceSelector,
) error {

	var lo client.ListOptions
	sel.SetListOptions(&lo)
	if err := b.List(ctx, list, &lo); err != nil {
		return fmt.Errorf("failed to list objects: %w", err)
	}

	elements, err := meta.ExtractList(list)
	if err != nil {
		return fmt.Errorf("failed to get list: %w", err)
	}

	filteredElements := make([]runtime.Object, 0, len(elements))
	for _, e := range elements {
		o, ok := e.(client.Object)
		if !ok {
			return fmt.Errorf("unexpected object returned")
		}

		if sel.Matches(b.Logger, b.Client, o) {
			filteredElements = append(filteredElements, e)
		}
	}

	if err := meta.SetList(list, filteredElements); err != nil {
		return fmt.Errorf("failed to populate list of objects: %w", err)
	}
	return nil
}

func (b *deploymentBuilder) getServiceMonitors(ctx context.Context, sel resourceSelector) ([]*prom.ServiceMonitor, error) {
	var list prom.ServiceMonitorList
	if err := b.list(ctx, &list, sel); err != nil {
		return nil, fmt.Errorf("unable to discover ServiceMonitors: %w", err)
	}

	items := make([]*prom.ServiceMonitor, 0, len(list.Items))
Item:
	for _, item := range list.Items {
		if b.Agent.Spec.Metrics.ArbitraryFSAccessThroughSMs.Deny {
			for _, ep := range item.Spec.Endpoints {
				err := testForArbitraryFSAccess(ep)
				if err == nil {
					continue
				}

				level.Warn(b.Logger).Log(
					"msg", "skipping service monitor",
					"agent", client.ObjectKeyFromObject(b.Agent),
					"servicemonitor", client.ObjectKeyFromObject(item),
					"err", err,
				)

				continue Item
			}
		}
		items = append(items, item)
	}

	return items, nil
}

func testForArbitraryFSAccess(e prom.Endpoint) error {
	if e.BearerTokenFile != "" {
		return fmt.Errorf("it accesses file system via bearer token file which is disallowed via GrafanaAgent specification")
	}

	if e.TLSConfig == nil {
		return nil
	}

	if e.TLSConfig.CAFile != "" || e.TLSConfig.CertFile != "" || e.TLSConfig.KeyFile != "" {
		return fmt.Errorf("it accesses file system via TLS config which is disallowed via GrafanaAgent specification")
	}

	return nil
}

func (b *deploymentBuilder) getPodMonitors(ctx context.Context, sel resourceSelector) ([]*prom.PodMonitor, error) {
	var list prom.PodMonitorList
	if err := b.list(ctx, &list, sel); err != nil {
		return nil, fmt.Errorf("unable to discover PodMonitors: %w", err)
	}
	return list.Items, nil
}

func (b *deploymentBuilder) getProbes(ctx context.Context, sel resourceSelector) ([]*prom.Probe, error) {
	var list prom.ProbeList
	if err := b.list(ctx, &list, sel); err != nil {
		return nil, fmt.Errorf("unable to discover Probes: %w", err)
	}
	return list.Items, nil
}

func (b *deploymentBuilder) getLogsInstances(ctx context.Context, sel resourceSelector) ([]*grafana.LogsInstance, error) {
	var list grafana.LogsInstanceList
	if err := b.list(ctx, &list, sel); err != nil {
		return nil, fmt.Errorf("unable to discover LogsInstances: %w", err)
	}
	return list.Items, nil
}

func (b *deploymentBuilder) getPodLogs(ctx context.Context, sel resourceSelector) ([]*grafana.PodLogs, error) {
	var list grafana.PodLogsList
	if err := b.list(ctx, &list, sel); err != nil {
		return nil, fmt.Errorf("unable to discover PodLogs: %w", err)
	}
	return list.Items, nil
}
