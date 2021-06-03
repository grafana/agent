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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
	ResourceSelectors map[secondaryResource][]ResourceSelector
}

func (b *deploymentBuilder) Build(ctx context.Context, l log.Logger) (config.Deployment, error) {
	instances, err := b.getPrometheusInstances(ctx)
	if err != nil {
		return config.Deployment{}, err
	}
	promInstances := make([]config.PrometheusInstance, 0, len(instances))

	for _, inst := range instances {
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

		promInstances = append(promInstances, config.PrometheusInstance{
			Instance:        inst,
			ServiceMonitors: sMons,
			PodMonitors:     pMons,
			Probes:          probes,
		})
	}

	return config.Deployment{
		Agent:      b.Agent,
		Prometheis: promInstances,
	}, nil
}

func (b *deploymentBuilder) getPrometheusInstances(ctx context.Context) ([]*grafana_v1alpha1.PrometheusInstance, error) {
	sel, err := b.getResourceSelector(
		b.Agent.Namespace,
		b.Agent.Spec.Prometheus.InstanceNamespaceSelector,
		b.Agent.Spec.Prometheus.InstanceSelector,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to build prometheus resource selector: %w", err)
	}
	b.ResourceSelectors[resourcePromInstance] = append(b.ResourceSelectors[resourcePromInstance], sel)

	var (
		list        grafana_v1alpha1.PrometheusInstanceList
		namespace   = namespaceFromSelector(sel)
		listOptions = &client.ListOptions{LabelSelector: sel.Labels, Namespace: namespace}
	)
	if err := b.List(ctx, &list, listOptions); err != nil {
		return nil, err
	}

	items := make([]*grafana_v1alpha1.PrometheusInstance, 0, len(list.Items))
	for _, item := range list.Items {
		if match, err := b.matchNamespace(ctx, &item.ObjectMeta, sel); match {
			items = append(items, item)
		} else if err != nil {
			return nil, fmt.Errorf("failed getting namespace: %w", err)
		}
	}
	return items, nil
}

func (b *deploymentBuilder) getResourceSelector(
	currentNamespace string,
	namespaceSelector *v1.LabelSelector,
	objectSelector *v1.LabelSelector,
) (sel ResourceSelector, err error) {

	// Set up our namespace label and object label selectors. By default, we'll
	// match everything (the inverse of the k8s default). If we specify anything,
	// we'll narrow it down.
	var (
		nsLabels  = labels.Everything()
		objLabels = labels.Everything()
	)
	if namespaceSelector != nil {
		nsLabels, err = v1.LabelSelectorAsSelector(namespaceSelector)
		if err != nil {
			return sel, err
		}
	}
	if objectSelector != nil {
		objLabels, err = v1.LabelSelectorAsSelector(objectSelector)
		if err != nil {
			return sel, err
		}
	}

	sel = ResourceSelector{
		NamespaceName: prom.NamespaceSelector{
			MatchNames: []string{currentNamespace},
		},
		NamespaceLabels: nsLabels,
		Labels:          objLabels,
	}

	// If we have a namespace selector, that means we're matching more than one
	// namespace and we should adjust NamespaceName appropriatel.
	if namespaceSelector != nil {
		sel.NamespaceName = prom.NamespaceSelector{Any: true}
	}

	return
}

// namespaceFromSelector returns the namespace string that should be used for
// querying lists of objects. If the ResourceSelector is looking at more than
// one namespace, an empty string will be returned. Otherwise, it will return
// the first namespace.
func namespaceFromSelector(sel ResourceSelector) string {
	if !sel.NamespaceName.Any && len(sel.NamespaceName.MatchNames) == 1 {
		return sel.NamespaceName.MatchNames[0]
	}
	return ""
}

func (b *deploymentBuilder) matchNamespace(
	ctx context.Context,
	obj *v1.ObjectMeta,
	sel ResourceSelector,
) (bool, error) {
	// If we were matching on a specific namespace, there's no
	// further work to do here.
	if namespaceFromSelector(sel) != "" {
		return true, nil
	}

	var ns core_v1.Namespace
	if err := b.Get(ctx, types.NamespacedName{Name: obj.Namespace}, &ns); err != nil {
		return false, fmt.Errorf("failed getting namespace: %w", err)
	}

	return sel.NamespaceLabels.Matches(labels.Set(ns.Labels)), nil
}

func (b *deploymentBuilder) getServiceMonitors(
	ctx context.Context,
	l log.Logger,
	inst *grafana_v1alpha1.PrometheusInstance,
) ([]*prom.ServiceMonitor, error) {
	sel, err := b.getResourceSelector(
		inst.Namespace,
		inst.Spec.ServiceMonitorNamespaceSelector,
		inst.Spec.ServiceMonitorSelector,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to build service monitor resource selector: %w", err)
	}
	b.ResourceSelectors[resourceServiceMonitor] = append(b.ResourceSelectors[resourceServiceMonitor], sel)

	var (
		list        prom.ServiceMonitorList
		namespace   = namespaceFromSelector(sel)
		listOptions = &client.ListOptions{LabelSelector: sel.Labels, Namespace: namespace}
	)
	if err := b.List(ctx, &list, listOptions); err != nil {
		return nil, err
	}

	items := make([]*prom.ServiceMonitor, 0, len(list.Items))

Item:
	for _, item := range list.Items {
		match, err := b.matchNamespace(ctx, &item.ObjectMeta, sel)
		if err != nil {
			return nil, fmt.Errorf("failed getting namespace: %w", err)
		} else if !match {
			continue
		}

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
	sel, err := b.getResourceSelector(
		inst.Namespace,
		inst.Spec.PodMonitorNamespaceSelector,
		inst.Spec.PodMonitorSelector,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to build service monitor resource selector: %w", err)
	}
	b.ResourceSelectors[resourcePodMonitor] = append(b.ResourceSelectors[resourcePodMonitor], sel)

	var (
		list        prom.PodMonitorList
		namespace   = namespaceFromSelector(sel)
		listOptions = &client.ListOptions{LabelSelector: sel.Labels, Namespace: namespace}
	)
	if err := b.List(ctx, &list, listOptions); err != nil {
		return nil, err
	}

	items := make([]*prom.PodMonitor, 0, len(list.Items))
	for _, item := range list.Items {
		if match, err := b.matchNamespace(ctx, &item.ObjectMeta, sel); match {
			items = append(items, item)
		} else if err != nil {
			return nil, fmt.Errorf("failed getting namespace: %w", err)
		}
	}
	return items, nil
}

func (b *deploymentBuilder) getProbes(
	ctx context.Context,
	inst *grafana_v1alpha1.PrometheusInstance,
) ([]*prom.Probe, error) {
	sel, err := b.getResourceSelector(
		inst.Namespace,
		inst.Spec.ProbeNamespaceSelector,
		inst.Spec.ProbeSelector,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to build service monitor resource selector: %w", err)
	}
	b.ResourceSelectors[resourceProbe] = append(b.ResourceSelectors[resourceProbe], sel)

	var namespace string
	if !sel.NamespaceName.Any && len(sel.NamespaceName.MatchNames) == 1 {
		namespace = sel.NamespaceName.MatchNames[0]
	}

	var list prom.ProbeList
	listOptions := &client.ListOptions{LabelSelector: sel.Labels, Namespace: namespace}
	if err := b.List(ctx, &list, listOptions); err != nil {
		return nil, err
	}

	items := make([]*prom.Probe, 0, len(list.Items))
	for _, item := range list.Items {
		if match, err := b.matchNamespace(ctx, &item.ObjectMeta, sel); match {
			items = append(items, item)
		} else if err != nil {
			return nil, fmt.Errorf("failed getting namespace: %w", err)
		}
	}
	return items, nil
}
