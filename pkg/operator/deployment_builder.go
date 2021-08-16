package operator

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	grafana_v1alpha1 "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/grafana/agent/pkg/operator/assets"
	"github.com/grafana/agent/pkg/operator/config"
	prom "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	core_v1 "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type deploymentBuilder struct {
	client.Client

	Config  *Config
	Agent   *grafana_v1alpha1.GrafanaAgent
	Secrets assets.SecretStore

	// ResourceSelectors is filled as objects are found and can be used to
	// trigger future reconciles.
	ResourceSelectors map[secondaryResource][]resourceSelector
}

func (b *deploymentBuilder) Build(ctx context.Context, l log.Logger) (config.Deployment, error) {
	rootMetricInstances, err := b.getPrometheusInstances(ctx)
	if err != nil {
		return config.Deployment{}, err
	}
	metricInstances := make([]config.PrometheusInstance, 0, len(rootMetricInstances))

	for _, inst := range rootMetricInstances {
		sMons, err := b.getServiceMonitors(ctx, l, inst)
		if err != nil {
			return config.Deployment{}, fmt.Errorf("unable to fetch ServiceMonitors: %w", err)
		}
		pMons, err := b.getPodMonitors(ctx, inst)
		if err != nil {
			return config.Deployment{}, fmt.Errorf("unable to fetch PodMonitors: %w", err)
		}
		probes, err := b.getProbes(ctx, inst)
		if err != nil {
			return config.Deployment{}, fmt.Errorf("unable to fetch Probes: %w", err)
		}

		metricInstances = append(metricInstances, config.PrometheusInstance{
			Instance:        inst,
			ServiceMonitors: sMons,
			PodMonitors:     pMons,
			Probes:          probes,
		})
	}

	rootLogsInstances, err := b.getLogsInstances(ctx)
	if err != nil {
		return config.Deployment{}, err
	}
	logsInstances := make([]config.LogInstance, 0, len(rootLogsInstances))

	for _, inst := range rootLogsInstances {
		podLogs, err := b.getPodLogs(ctx, inst)
		if err != nil {
			return config.Deployment{}, fmt.Errorf("unable to fetch PodLogs: %w", err)
		}

		logsInstances = append(logsInstances, config.LogInstance{
			Instance: inst,
			PodLogs:  podLogs,
		})
	}

	return config.Deployment{
		Agent:      b.Agent,
		Prometheis: metricInstances,
		Logs:       logsInstances,
	}, nil
}

// list will search for all objects in list using the provided
// namespaceSelector and objectSelector. When namespaceSelector is nil, only
// objects in parentNamespace are returned.
//
// Returns a resourceSelector that can be used to be notified of one of the
// discovered objects changing.
func (b *deploymentBuilder) list(
	ctx context.Context,
	list client.ObjectList,
	parentNamespace string,
	namespaceSelector *v1.LabelSelector,
	objectSelector *v1.LabelSelector,
) (resourceSelector, error) {
	var namespace string
	if namespaceSelector == nil {
		namespace = parentNamespace
	}

	nsLabels, err := v1.LabelSelectorAsSelector(namespaceSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to convert object namespace selector into label selector: %w", err)
	}
	objectLabels, err := v1.LabelSelectorAsSelector(objectSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to convert object selector into label selector: %w", err)
	}

	listOptions := &client.ListOptions{
		LabelSelector: objectLabels,
		Namespace:     namespace,
	}
	if err := b.List(ctx, list, listOptions); err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	elements, err := api_meta.ExtractList(list)
	if err != nil {
		return nil, fmt.Errorf("failed to get list: %w", err)
	}

	filteredElements := make([]runtime.Object, 0, len(elements))
	for _, e := range elements {
		o, ok := e.(client.Object)
		if !ok {
			return nil, fmt.Errorf("unexpected object returned")
		}

		// We aren't using a namespaceSelector when namespace is set.
		// Don't do any matching.
		if namespace != "" {
			continue
		}

		var ns core_v1.Namespace
		if err := b.Get(ctx, types.NamespacedName{Name: o.GetNamespace()}, &ns); err != nil {
			return nil, fmt.Errorf("failed getting namespace for selector filtering: %w", err)
		}
		if nsLabels.Matches(labels.Set(ns.Labels)) {
			filteredElements = append(filteredElements, e)
		}
	}

	if err := api_meta.SetList(list, elements); err != nil {
		return nil, fmt.Errorf("failed to populate list of objects: %w", err)
	}
	return b.getResourceSelector(namespace, nsLabels, objectLabels), nil
}

func (b *deploymentBuilder) getResourceSelector(
	matchNamespace string,
	nsSel labels.Selector,
	oSel labels.Selector,
) resourceSelector {
	var ss []resourceSelector

	if matchNamespace != "" {
		ss = append(ss, &namespaceSelector{Namespace: matchNamespace})
	} else {
		ss = append(ss, &namespaceLabelSelector{Selector: nsSel})
	}

	ss = append(ss, &labelSelector{Selector: oSel})

	return &multiSelector{Selectors: ss}
}

func (b *deploymentBuilder) getPrometheusInstances(ctx context.Context) ([]*grafana_v1alpha1.PrometheusInstance, error) {
	var list grafana_v1alpha1.PrometheusInstanceList
	sel, err := b.list(
		ctx,
		&list,
		b.Agent.Namespace,
		b.Agent.Spec.Prometheus.InstanceNamespaceSelector,
		b.Agent.Spec.Prometheus.InstanceSelector,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to discover PrometheusInstances: %w", err)
	}
	b.ResourceSelectors[resourcePromInstance] = append(b.ResourceSelectors[resourcePromInstance], sel)
	return list.Items, nil
}

func (b *deploymentBuilder) getServiceMonitors(
	ctx context.Context,
	l log.Logger,
	inst *grafana_v1alpha1.PrometheusInstance,
) ([]*prom.ServiceMonitor, error) {
	var list prom.ServiceMonitorList
	sel, err := b.list(
		ctx,
		&list,
		inst.Namespace,
		inst.Spec.ServiceMonitorNamespaceSelector,
		inst.Spec.ServiceMonitorSelector,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to discover ServiceMonitors: %w", err)
	}
	b.ResourceSelectors[resourceServiceMonitor] = append(b.ResourceSelectors[resourceServiceMonitor], sel)

	items := make([]*prom.ServiceMonitor, 0, len(list.Items))
Item:
	for _, item := range list.Items {
		if b.Agent.Spec.Prometheus.ArbitraryFSAccessThroughSMs.Deny {
			for _, ep := range item.Spec.Endpoints {
				err := testForArbitraryFSAccess(ep)
				if err == nil {
					continue
				}

				level.Warn(l).Log(
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

func (b *deploymentBuilder) getPodMonitors(
	ctx context.Context,
	inst *grafana_v1alpha1.PrometheusInstance,
) ([]*prom.PodMonitor, error) {
	var list prom.PodMonitorList
	sel, err := b.list(
		ctx,
		&list,
		inst.Namespace,
		inst.Spec.PodMonitorNamespaceSelector,
		inst.Spec.PodMonitorSelector,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to discover PodMonitors: %w", err)
	}
	b.ResourceSelectors[resourcePodMonitor] = append(b.ResourceSelectors[resourcePodMonitor], sel)
	return list.Items, nil
}

func (b *deploymentBuilder) getProbes(
	ctx context.Context,
	inst *grafana_v1alpha1.PrometheusInstance,
) ([]*prom.Probe, error) {
	var list prom.ProbeList
	sel, err := b.list(
		ctx,
		&list,
		inst.Namespace,
		inst.Spec.ProbeNamespaceSelector,
		inst.Spec.ProbeSelector,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to discover PodMonitors: %w", err)
	}
	b.ResourceSelectors[resourceProbe] = append(b.ResourceSelectors[resourceProbe], sel)
	return list.Items, nil
}

func (b *deploymentBuilder) getLogsInstances(ctx context.Context) ([]*grafana_v1alpha1.LogsInstance, error) {
	sel, err := b.getResourceSelector(
		b.Agent.Namespace,
		b.Agent.Spec.Logs.InstanceNamespaceSelector,
		b.Agent.Spec.Logs.InstanceSelector,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to build logs resource selector: %w", err)
	}
	b.ResourceSelectors[resourceLogsInstance] = append(b.ResourceSelectors[resourceLogsInstance], sel)

	var (
		list        grafana_v1alpha1.LogsInstanceList
		namespace   = namespaceFromSelector(sel)
		listOptions = &client.ListOptions{LabelSelector: sel.Labels, Namespace: namespace}
	)
	if err := b.List(ctx, &list, listOptions); err != nil {
		return nil, err
	}

	items := make([]*grafana_v1alpha1.LogsInstance, 0, len(list.Items))
	for _, item := range list.Items {
		if match, err := b.matchNamespace(ctx, &item.ObjectMeta, sel); match {
			items = append(items, item)
		} else if err != nil {
			return nil, fmt.Errorf("failed getting namespace: %w", err)
		}
	}
	return items, nil
}

func (b *deploymentBuilder) getPodLogs(
	ctx context.Context,
	inst *grafana_v1alpha1.LogsInstance,
) ([]*grafana_v1alpha1.PodLogs, error) {
	sel, err := b.getResourceSelector(
		inst.Namespace,
		inst.Spec.PodLogsNamespaceSelector,
		inst.Spec.PodLogsSelector,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to build service monitor resource selector: %w", err)
	}
	b.ResourceSelectors[resourcePodLogs] = append(b.ResourceSelectors[resourcePodLogs], sel)

	var (
		list        grafana_v1alpha1.PodLogsList
		namespace   = namespaceFromSelector(sel)
		listOptions = &client.ListOptions{LabelSelector: sel.Labels, Namespace: namespace}
	)
	if err := b.List(ctx, &list, listOptions); err != nil {
		return nil, err
	}

	items := make([]*grafana_v1alpha1.PodLogs, 0, len(list.Items))
	for _, item := range list.Items {
		if match, err := b.matchNamespace(ctx, &item.ObjectMeta, sel); match {
			items = append(items, item)
		} else if err != nil {
			return nil, fmt.Errorf("failed getting namespace: %w", err)
		}
	}
	return items, nil
}
