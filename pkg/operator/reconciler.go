package operator

import (
	"context"
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	grafana_v1alpha1 "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/grafana/agent/pkg/operator/assets"
	"github.com/grafana/agent/pkg/operator/clientutil"
	"github.com/grafana/agent/pkg/operator/config"
	"github.com/grafana/agent/pkg/operator/hierarchy"
	"github.com/grafana/agent/pkg/operator/logutil"
	core_v1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	controller "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type reconciler struct {
	client.Client
	scheme *runtime.Scheme
	config *Config

	notifier *hierarchy.Notifier
}

func (r *reconciler) Reconcile(ctx context.Context, req controller.Request) (controller.Result, error) {
	l := logutil.FromContext(ctx)
	level.Info(l).Log("msg", "reconciling grafana-agent")
	defer level.Debug(l).Log("msg", "done reconciling grafana-agent")

	// Reset our notifications while we re-handle the reconcile.
	r.notifier.StopNotify(req.NamespacedName)

	var agent grafana_v1alpha1.GrafanaAgent
	if err := r.Get(ctx, req.NamespacedName, &agent); k8s_errors.IsNotFound(err) {
		level.Debug(l).Log("msg", "detected deleted agent")
		return controller.Result{}, nil
	} else if err != nil {
		level.Error(l).Log("msg", "unable to get grafana-agent", "err", err)
		return controller.Result{}, nil
	}

	if agent.Spec.Paused {
		return controller.Result{}, nil
	}

	deployment, watchers, err := buildHierarchy(ctx, l, r.Client, &agent)
	if err != nil {
		level.Error(l).Log("msg", "unable to build hierarchy", "err", err)
		return controller.Result{}, nil
	}
	if err := r.notifier.Notify(watchers...); err != nil {
		level.Error(l).Log("msg", "unable to update notifier", "err", err)
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
		err := actor(ctx, l, deployment, deployment.Secrets)
		if err != nil {
			level.Error(l).Log("msg", "error during reconciling", "err", err)
			return controller.Result{Requeue: true}, nil
		}
	}

	return controller.Result{}, nil
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
			Labels: map[string]string{
				managedByOperatorLabel: managedByOperatorLabelValue,
			},
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
