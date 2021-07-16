package operator

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/operator/assets"
	"github.com/grafana/agent/pkg/operator/clientutil"
	"github.com/grafana/agent/pkg/operator/config"
	apps_v1 "k8s.io/api/apps/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// createLogsConfigurationSecret creates the Grafana Agent logs configuration
// and stores it into a secret.
func (r *reconciler) createLogsConfigurationSecret(
	ctx context.Context,
	l log.Logger,
	d config.Deployment,
	s assets.SecretStore,
) error {
	return r.createTelemetryConfigurationSecret(ctx, l, d, s, config.LogsType)
}

// createLogsDaemonSet creates a DaemonSet for logs.
func (r *reconciler) createLogsDaemonSet(
	ctx context.Context,
	l log.Logger,
	d config.Deployment,
	s assets.SecretStore,
) error {
	name := fmt.Sprintf("%s-logs", d.Agent.Name)
	ds, err := generateLogsDaemonSet(r.config, name, d)
	if err != nil {
		return fmt.Errorf("failed to generate DaemonSet: %w", err)
	}
	key := types.NamespacedName{Namespace: ds.Namespace, Name: ds.Name}

	if len(d.Logs) == 0 {

		var ds apps_v1.DaemonSet
		err := r.Client.Get(ctx, key, &ds)
		if k8s_errors.IsNotFound(err) {
			return nil
		} else if err != nil {
			return fmt.Errorf("failed to find stale DaemonSet %s: %w", key, err)
		}

		err = r.Client.Delete(ctx, &ds)
		if err != nil {
			return fmt.Errorf("failed to delete stale DaemonSet %s: %w", key, err)
		}
		return nil
	}

	level.Info(l).Log("msg", "reconciling logs daemonset", "ds", key)
	err = clientutil.CreateOrUpdateDaemonSet(ctx, r.Client, ds)
	if err != nil {
		return fmt.Errorf("failed to reconcile statefulset governing service: %w", err)
	}
	return nil
}
