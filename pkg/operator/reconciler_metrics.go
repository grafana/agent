package operator

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-jsonnet"
	"github.com/grafana/agent/pkg/operator/assets"
	"github.com/grafana/agent/pkg/operator/clientutil"
	"github.com/grafana/agent/pkg/operator/config"
	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// createMetricsConfigurationSecret creates the Grafana Agent metrics configuration and stores
// it into a secret.
func (r *reconciler) createMetricsConfigurationSecret(
	ctx context.Context,
	l log.Logger,
	d config.Deployment,
	s assets.SecretStore,
) error {
	return r.createTelemetryConfigurationSecret(ctx, l, d, s, config.MetricsType)
}

func (r *reconciler) createTelemetryConfigurationSecret(
	ctx context.Context,
	l log.Logger,
	d config.Deployment,
	s assets.SecretStore,
	ty config.Type,
) error {

	var shouldCreate bool
	key := types.NamespacedName{Namespace: d.Agent.Namespace}

	switch ty {
	case config.MetricsType:
		key.Name = fmt.Sprintf("%s-config", d.Agent.Name)
		shouldCreate = len(d.Metrics) > 0
	case config.LogsType:
		key.Name = fmt.Sprintf("%s-logs-config", d.Agent.Name)
		shouldCreate = len(d.Logs) > 0
	default:
		return fmt.Errorf("unknown telemetry type %s", ty)
	}

	// Delete the old Secret if one exists and we have nothing to create.
	if !shouldCreate {
		var secret core_v1.Secret
		err := r.Client.Get(ctx, key, &secret)
		if k8s_errors.IsNotFound(err) || !isManagedResource(&secret) {
			return nil
		} else if err != nil {
			return fmt.Errorf("failed to find stale secret %s: %w", key, err)
		}

		err = r.Client.Delete(ctx, &secret)
		if err != nil {
			return fmt.Errorf("failed to delete stale secret %s: %w", key, err)
		}
		return nil
	}

	rawConfig, err := d.BuildConfig(s, ty)

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
			Namespace: key.Namespace,
			Name:      key.Name,
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

// createMetricsGoverningService creates the service that governs the (eventual)
// StatefulSet. It must be created before the StatefulSet.
func (r *reconciler) createMetricsGoverningService(
	ctx context.Context,
	l log.Logger,
	d config.Deployment,
	s assets.SecretStore,
) error {
	svc := generateMetricsStatefulSetService(r.config, d)

	// Delete the old Secret if one exists and we have no prometheus instances.
	if len(d.Metrics) == 0 {
		key := types.NamespacedName{Namespace: svc.Namespace, Name: svc.Name}

		var service core_v1.Service
		err := r.Client.Get(ctx, key, &service)
		if k8s_errors.IsNotFound(err) || !isManagedResource(&service) {
			return nil
		} else if err != nil {
			return fmt.Errorf("failed to find stale Service %s: %w", key, err)
		}

		err = r.Client.Delete(ctx, &service)
		if err != nil {
			return fmt.Errorf("failed to delete stale Service %s: %w", key, err)
		}
		return nil
	}

	level.Info(l).Log("msg", "reconciling statefulset service", "service", svc.Name)
	err := clientutil.CreateOrUpdateService(ctx, r.Client, svc)
	if err != nil {
		return fmt.Errorf("failed to reconcile statefulset governing service: %w", err)
	}
	return nil
}

// createMetricsStatefulSets creates a set of Grafana Agent StatefulSets, one per shard.
func (r *reconciler) createMetricsStatefulSets(
	ctx context.Context,
	l log.Logger,
	d config.Deployment,
	s assets.SecretStore,
) error {

	shards := minShards
	if reqShards := d.Agent.Spec.Metrics.Shards; reqShards != nil && *reqShards > 1 {
		shards = *reqShards
	}

	// Keep track of generated stateful sets so we can delete ones that should
	// no longer exist.
	generated := make(map[string]struct{})

	for shard := int32(0); shard < shards; shard++ {
		// Don't generate anything if there weren't any instances.
		if len(d.Metrics) == 0 {
			continue
		}

		name := d.Agent.Name
		if shard > 0 {
			name = fmt.Sprintf("%s-shard-%d", name, shard)
		}

		ss, err := generateMetricsStatefulSet(r.config, name, d, shard)
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
			managedByOperatorLabel: managedByOperatorLabelValue,
			agentNameLabelName:     d.Agent.Name,
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
