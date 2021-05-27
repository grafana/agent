package operator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-jsonnet"
	grafana_v1alpha1 "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/grafana/agent/pkg/operator/assets"
	"github.com/grafana/agent/pkg/operator/clientutil"
	"github.com/grafana/agent/pkg/operator/config"
	"github.com/grafana/agent/pkg/operator/logutil"
	prom "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	controller "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type reconciler struct {
	client.Client
	scheme *runtime.Scheme
	config *Config

	eventHandlers eventHandlers
}

func (r *reconciler) Reconcile(ctx context.Context, req controller.Request) (controller.Result, error) {
	l := logutil.FromContext(ctx)
	level.Info(l).Log("msg", "reconciling grafana-agent")

	var agent grafana_v1alpha1.GrafanaAgent
	if err := r.Get(ctx, req.NamespacedName, &agent); k8s_errors.IsNotFound(err) {
		level.Debug(l).Log("msg", "detected deleted agent, cleaning up watchers")
		r.eventHandlers.Clear(req.NamespacedName)

		return controller.Result{}, nil
	} else if err != nil {
		level.Error(l).Log("msg", "unable to get grafana-agent", "err", err)
		return controller.Result{}, nil
	}

	if agent.Spec.Paused {
		return controller.Result{}, nil
	}

	secrets := make(assets.SecretStore)
	builder := deploymentBuilder{
		Config:            r.config,
		Client:            r.Client,
		Agent:             &agent,
		Secrets:           secrets,
		ResourceSelectors: make(map[secondaryResource][]ResourceSelector),
	}

	deployment, err := builder.Build(ctx, l)
	if err != nil {
		level.Error(l).Log("msg", "unable to collect resources", "err", err)
		return controller.Result{}, nil
	}

	// Fill secrets in store
	if err := r.fillStore(ctx, deployment.AssetReferences(), secrets); err != nil {
		level.Error(l).Log("msg", "unable to cache secrets for building config", "err", err)
		return controller.Result{}, nil
	}

	// Create configuration in a secret
	{
		rawConfig, err := deployment.BuildConfig(secrets)

		var jsonnetError jsonnet.RuntimeError
		if errors.As(err, &jsonnetError) {
			// Jump Jsonnet errors to the console to retain newlines and make them
			// easier to digest.
			fmt.Fprintf(os.Stderr, "%s", jsonnetError.Error())
		}
		if err != nil {
			level.Error(l).Log("msg", "unable to build config", "err", err)
			return controller.Result{}, nil
		}

		blockOwnerDeletion := true

		secret := core_v1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Namespace: agent.Namespace,
				Name:      fmt.Sprintf("%s-config", agent.Name),
				Labels:    r.config.Labels.Merge(managedByOperatorLabels),
				OwnerReferences: []v1.OwnerReference{{
					APIVersion:         agent.APIVersion,
					BlockOwnerDeletion: &blockOwnerDeletion,
					Kind:               agent.Kind,
					Name:               agent.Name,
					UID:                agent.UID,
				}},
			},
			Data: map[string][]byte{"agent.yml": []byte(rawConfig)},
		}

		level.Info(l).Log("msg", "creating or updating secret", "secret", secret.Name)
		err = r.Client.Update(ctx, &secret)
		if k8s_errors.IsNotFound(err) {
			err = r.Client.Create(ctx, &secret)
		}
		if k8s_errors.IsAlreadyExists(err) {
			return controller.Result{Requeue: true}, nil
		}
		if err != nil {
			level.Error(l).Log("msg", "failed to create Secret for storing config", "err", err)
			return controller.Result{Requeue: true}, nil
		}
	}

	// Create secrets from asset store. These will be used to create volume
	// mounts into pods.
	{
		blockOwnerDeletion := true

		data := make(map[string][]byte)
		for k, value := range secrets {
			data[config.SanitizeLabelName(string(k))] = []byte(value)
		}

		secret := core_v1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Namespace: agent.Namespace,
				Name:      fmt.Sprintf("%s-secrets", agent.Name),
				OwnerReferences: []v1.OwnerReference{{
					APIVersion:         agent.APIVersion,
					BlockOwnerDeletion: &blockOwnerDeletion,
					Kind:               agent.Kind,
					Name:               agent.Name,
					UID:                agent.UID,
				}},
			},
			Data: data,
		}

		level.Info(l).Log("msg", "creating or updating secret", "secret", secret.Name)
		err = r.Client.Update(ctx, &secret)
		if k8s_errors.IsNotFound(err) {
			err = r.Client.Create(ctx, &secret)
		}
		if k8s_errors.IsAlreadyExists(err) {
			return controller.Result{Requeue: true}, nil
		}
		if err != nil {
			level.Error(l).Log("msg", "failed to create Secret for storing config", "err", err)
			return controller.Result{Requeue: true}, nil
		}
	}

	// Generate governing service.
	{
		svc := generateStatefulSetService(r.config, deployment)
		level.Info(l).Log("msg", "creating or updating statefulset service", "service", svc.Name)
		err := clientutil.CreateOrUpdateService(ctx, r.Client, svc)
		if err != nil {
			level.Error(l).Log("msg", "failed to create statefulset service", "err", err)
			return controller.Result{}, nil
		}
	}

	// Generate and create StatefulSet for all shards.
	{
		shards := minShards
		if reqShards := deployment.Agent.Spec.Prometheus.Shards; reqShards != nil && *reqShards > 1 {
			shards = *reqShards
		}

		// TODO(rfratto): when refactoring, keep in mind that returning an error
		// will cause it to requeue.

		// Keep track of generated stateful sets so we can delete ones that should
		// no longer exist.
		generated := make(map[string]struct{})

		for shard := int32(0); shard < shards; shard++ {
			name := deployment.Agent.Name
			if shard > 0 {
				name = fmt.Sprintf("%s-shard-%d", name, shard)
			}

			ss, err := generateStatefulSet(r.config, name, deployment, shard)
			if err != nil {
				level.Error(l).Log("msg", "failed to generate statefulset", "err", err)
				return controller.Result{Requeue: false}, nil
			}

			level.Info(l).Log("msg", "creating or updating statefulset", "statefulset", ss.Name)
			err = r.Client.Update(ctx, ss)
			if k8s_errors.IsNotFound(err) {
				err = r.Client.Create(ctx, ss)
			}
			if k8s_errors.IsNotAcceptable(err) {
				level.Error(l).Log("msg", "unacceptable change to statefulset. deleting and re-querying", "err", err)
				err = r.Client.Delete(ctx, ss)
				if err != nil {
					level.Error(l).Log("msg", "failed to delete unacceptable statefulset")
				}
				return controller.Result{
					Requeue:      true,
					RequeueAfter: 500 * time.Millisecond,
				}, nil
			}
			if k8s_errors.IsAlreadyExists(err) {
				return controller.Result{Requeue: true}, nil
			}
			if err != nil {
				level.Error(l).Log("msg", "failed to create statefulset", "err", err)
				return controller.Result{}, nil
			}

			generated[ss.Name] = struct{}{}
		}

		// Clean up statefulsets that should no longer exist.
		var statefulSets apps_v1.StatefulSetList
		err = r.List(ctx, &statefulSets, &client.ListOptions{
			LabelSelector: labels.SelectorFromSet(labels.Set{
				agentNameLabelName: deployment.Agent.Name,
			}),
		})
		if err != nil {
			level.Error(l).Log("msg", "failed to list statefulsets", "err", err)
			return controller.Result{}, nil
		}
		for _, ss := range statefulSets.Items {
			if _, keep := generated[ss.Name]; keep {
				continue
			}
			level.Info(l).Log("msg", "deleting stale statefulset", "name", ss.Name)
			if err := r.Client.Delete(ctx, &ss); err != nil {
				level.Warn(l).Log("msg", "failed to delete stale statefulset", "err", err)
			}
		}
	}

	// Update our notifiers with every object we discovered. This ensures that we
	// will re-reconcile whenever any of these objects (which composed our configs)
	// changes.
	for _, secondary := range secondaryResources {
		r.eventHandlers[secondary].Notify(req.NamespacedName, builder.ResourceSelectors[secondary])
	}

	return controller.Result{}, nil
}

// fillStore retrieves all the values from refs and caches them in the provided store.
func (r *reconciler) fillStore(ctx context.Context, refs []config.AssetReference, store assets.SecretStore) error {
	for _, ref := range refs {
		var value string

		if ref.Reference.ConfigMap != nil {
			var cm core_v1.ConfigMap
			name := types.NamespacedName{
				Namespace: ref.Namespace,
				Name:      ref.Reference.ConfigMap.Name,
			}

			if err := r.Get(ctx, name, &cm); err != nil {
				return err
			}

			rawValue, ok := cm.BinaryData[ref.Reference.ConfigMap.Key]
			if !ok {
				return fmt.Errorf("no key %s in ConfigMap %s", ref.Reference.ConfigMap.Key, name)
			}
			value = string(rawValue)
		} else if ref.Reference.Secret != nil {
			var secret core_v1.Secret
			name := types.NamespacedName{
				Namespace: ref.Namespace,
				Name:      ref.Reference.Secret.Name,
			}

			if err := r.Get(ctx, name, &secret); err != nil {
				return err
			}

			rawValue, ok := secret.Data[ref.Reference.Secret.Key]
			if !ok {
				return fmt.Errorf("no key %s in Secret %s", ref.Reference.ConfigMap.Key, name)
			}
			value = string(rawValue)
		}

		store[assets.KeyForSelector(ref.Namespace, &ref.Reference)] = value
	}

	return nil
}

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
