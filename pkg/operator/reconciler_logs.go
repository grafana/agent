package operator

import (
	"context"
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	grafana_v1alpha1 "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/grafana/agent/pkg/operator/clientutil"
	"github.com/grafana/agent/pkg/operator/config"
	apps_v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

// createLogsConfigurationSecret creates the Grafana Agent logs configuration
// and stores it into a secret.
func (r *reconciler) createLogsConfigurationSecret(
	ctx context.Context,
	l log.Logger,
	h grafana_v1alpha1.Hierarchy,
) error {
	return r.createTelemetryConfigurationSecret(ctx, l, h, config.LogsType)
}

// createLogsDaemonSet creates a DaemonSet for logs.
func (r *reconciler) createLogsDaemonSet(
	ctx context.Context,
	l log.Logger,
	h grafana_v1alpha1.Hierarchy,
) error {
	name := fmt.Sprintf("%s-logs", h.Agent.Name)
	ds, err := generateLogsDaemonSet(r.config, name, h)
	if err != nil {
		return fmt.Errorf("failed to generate DaemonSet: %w", err)
	}
	key := types.NamespacedName{Namespace: ds.Namespace, Name: ds.Name}

	if len(h.Logs) == 0 {
		var ds apps_v1.DaemonSet
		return deleteManagedResource(ctx, r.Client, key, &ds)
	}

	level.Info(l).Log("msg", "reconciling logs daemonset", "ds", key)
	err = clientutil.CreateOrUpdateDaemonSet(ctx, r.Client, ds)
	if err != nil {
		return fmt.Errorf("failed to reconcile statefulset governing service: %w", err)
	}
	return nil
}
