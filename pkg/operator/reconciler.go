package operator

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-jsonnet"
	grafana_v1alpha1 "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/grafana/agent/pkg/operator/assets"
	"github.com/grafana/agent/pkg/operator/clientutil"
	"github.com/grafana/agent/pkg/operator/config"
	"github.com/grafana/agent/pkg/operator/logutil"
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

	// Update our notifiers with the objects we discovered from building the
	// deployment. This allows us to re-reconcile when any of the objects that
	// composed our final deployment changes.
	for _, secondary := range secondaryResources {
		r.eventHandlers[secondary].Notify(req.NamespacedName, builder.ResourceSelectors[secondary])
	}

	// Fill secrets in store
	if err := r.fillStore(ctx, deployment.AssetReferences(), secrets); err != nil {
		level.Error(l).Log("msg", "unable to cache secrets for building config", "err", err)
		return controller.Result{}, nil
	}

	type reconcileFunc func(context.Context, log.Logger, config.Deployment, assets.SecretStore) error
	actors := []reconcileFunc{
		r.createConfigurationSecret,
		r.createSecrets,
		r.createGoverningService,
		r.createStatefulSets,
	}
	for _, actor := range actors {
		err := actor(ctx, l, deployment, secrets)
		if err != nil {
			level.Error(l).Log("msg", "error during reconciling", "err", err)
			return controller.Result{Requeue: true}, nil
		}
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

// createConfigurationSecret creates the Grafana Agent configuration and stores
// it into a secret.
func (r *reconciler) createConfigurationSecret(
	ctx context.Context,
	l log.Logger,
	d config.Deployment,
	s assets.SecretStore,
) error {

	rawConfig, err := d.BuildConfig(s)

	var jsonnetError jsonnet.RuntimeError
	if errors.As(err, &jsonnetError) {
		// Dump Jsonnet errors to the console to retain newlines and make them
		// easier to digest.
		fmt.Fprintf(os.Stderr, "%s", jsonnetError.Error())
	}
	if err != nil {
		return fmt.Errorf("unable to build config: %w", err)
	}

	blockOwnerDeletion := true

	secret := core_v1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Namespace: d.Agent.Namespace,
			Name:      fmt.Sprintf("%s-config", d.Agent.Name),
			Labels:    r.config.Labels.Merge(managedByOperatorLabels),
			OwnerReferences: []v1.OwnerReference{{
				APIVersion:         d.Agent.APIVersion,
				BlockOwnerDeletion: &blockOwnerDeletion,
				Kind:               d.Agent.Kind,
				Name:               d.Agent.Name,
				UID:                d.Agent.UID,
			}},
		},
		Data: map[string][]byte{"agent.yml": []byte(rawConfig)},
	}

	level.Info(l).Log("msg", "reconciling secret", "secret", secret.Name)
	err = clientutil.CreateOrUpdateSecret(ctx, r.Client, &secret)
	if err != nil {
		return fmt.Errorf("failed to reconcile secret: %w", err)
	}
	return nil
}

// createSecrets creates secrets from the secret store.
func (r *reconciler) createSecrets(
	ctx context.Context,
	l log.Logger,
	d config.Deployment,
	s assets.SecretStore,
) error {

	blockOwnerDeletion := true

	data := make(map[string][]byte)
	for k, value := range s {
		data[config.SanitizeLabelName(string(k))] = []byte(value)
	}

	secret := core_v1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Namespace: d.Agent.Namespace,
			Name:      fmt.Sprintf("%s-secrets", d.Agent.Name),
			OwnerReferences: []v1.OwnerReference{{
				APIVersion:         d.Agent.APIVersion,
				BlockOwnerDeletion: &blockOwnerDeletion,
				Kind:               d.Agent.Kind,
				Name:               d.Agent.Name,
				UID:                d.Agent.UID,
			}},
		},
		Data: data,
	}

	level.Info(l).Log("msg", "reconciling secret", "secret", secret.Name)
	err := clientutil.CreateOrUpdateSecret(ctx, r.Client, &secret)
	if err != nil {
		return fmt.Errorf("failed to reconcile secret: %w", err)
	}
	return nil
}

// createGoverningService creates the service that governs the (eventual)
// StatefulSet. It must be created before the StatefulSet.
func (r *reconciler) createGoverningService(
	ctx context.Context,
	l log.Logger,
	d config.Deployment,
	s assets.SecretStore,
) error {
	svc := generateStatefulSetService(r.config, d)
	level.Info(l).Log("msg", "reconciling statefulset service", "service", svc.Name)
	err := clientutil.CreateOrUpdateService(ctx, r.Client, svc)
	if err != nil {
		return fmt.Errorf("failed to reconcile statefulset governing service: %w", err)
	}
	return nil
}

// createStatefulSets creates a set of Grafana Agent StatefulSets, one per shard.
func (r *reconciler) createStatefulSets(
	ctx context.Context,
	l log.Logger,
	d config.Deployment,
	s assets.SecretStore,
) error {

	shards := minShards
	if reqShards := d.Agent.Spec.Prometheus.Shards; reqShards != nil && *reqShards > 1 {
		shards = *reqShards
	}

	// Keep track of generated stateful sets so we can delete ones that should
	// no longer exist.
	generated := make(map[string]struct{})

	for shard := int32(0); shard < shards; shard++ {
		name := d.Agent.Name
		if shard > 0 {
			name = fmt.Sprintf("%s-shard-%d", name, shard)
		}

		ss, err := generateStatefulSet(r.config, name, d, shard)
		if err != nil {
			return fmt.Errorf("failed to generate statefulset for shard: %w", err)
		}

		level.Info(l).Log("msg", "reconciling statefulset", "statefulset", ss.Name)
		err = clientutil.CreateOrUpdateStatefulSet(ctx, r.Client, ss)
		if err != nil {
			return fmt.Errorf("failed to reconcile statefulset for shard: %w", err)
		}
		generated[ss.Name] = struct{}{}
	}

	// Clean up statefulsets that should no longer exist.
	var statefulSets apps_v1.StatefulSetList
	err := r.List(ctx, &statefulSets, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{
			agentNameLabelName: d.Agent.Name,
		}),
	})
	if err != nil {
		return fmt.Errorf("failed to list statefulsets: %w", err)
	}
	for _, ss := range statefulSets.Items {
		if _, keep := generated[ss.Name]; keep {
			continue
		}
		level.Info(l).Log("msg", "deleting stale statefulset", "name", ss.Name)
		if err := r.Client.Delete(ctx, &ss); err != nil {
			return fmt.Errorf("failed to delete stale statefulset %s: %w", ss.Name, err)
		}
	}

	return nil
}
