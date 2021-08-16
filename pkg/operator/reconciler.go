package operator

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	grafana_v1alpha1 "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/grafana/agent/pkg/operator/assets"
	"github.com/grafana/agent/pkg/operator/clientutil"
	"github.com/grafana/agent/pkg/operator/config"
	"github.com/grafana/agent/pkg/operator/logutil"
	core_v1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		ResourceSelectors: make(map[secondaryResource][]resourceSelector),
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

	assetRefs := deployment.AssetReferences()

	// Update our notifiers with asset references discovered through the deployment.
	r.watchSecrets(req, assetRefs)

	// Fill secrets in store
	if err := r.fillStore(ctx, assetRefs, secrets); err != nil {
		level.Error(l).Log("msg", "unable to cache secrets for building config", "err", err)
		return controller.Result{}, nil
	}

	type reconcileFunc func(context.Context, log.Logger, config.Deployment, assets.SecretStore) error
	actors := []reconcileFunc{
		// Operator-wide resources
		r.createSecrets,

		// Metrics resources (may be a no-op if no metrics configured)
		r.createMetricsConfigurationSecret,
		r.createMetricsGoverningService,
		r.createMetricsStatefulSets,

		// Logs resources (may be a no-op if no logs configured)
		r.createLogsConfigurationSecret,
		r.createLogsDaemonSet,
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

// watchSecrets will go iterate over asset references and configure them to be
// watched for updates. This allows reconciles to trigger when a referenced
// Secret or ConfigMap changes.
func (r *reconciler) watchSecrets(req controller.Request, refs []config.AssetReference) {
	var (
		configMapSelectors []resourceSelector
		secretSelectors    []resourceSelector
	)

	for _, ref := range refs {
		switch {
		case ref.Reference.ConfigMap != nil:
			configMapSelectors = append(configMapSelectors, &assetReferenceSelector{Reference: ref})
		case ref.Reference.Secret != nil:
			secretSelectors = append(secretSelectors, &assetReferenceSelector{Reference: ref})
		default:
			panic("unknown AssetReference")
		}
	}

	r.eventHandlers[resourceConfigMap].Notify(req.NamespacedName, configMapSelectors)
	r.eventHandlers[resourceSecret].Notify(req.NamespacedName, secretSelectors)
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
