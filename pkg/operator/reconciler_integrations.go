package operator

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/operator/assets"
	"github.com/grafana/agent/pkg/operator/config"
)

// createIntegrationsConfigurationSecrets creates the Grafana Agent logs
// configuration and stores it into a secret.
func (r *reconciler) createIntegrationsConfigurationSecrets(
	ctx context.Context,
	l log.Logger,
	d config.Deployment,
	s assets.SecretStore,
) error {
	return r.createTelemetryConfigurationSecret(ctx, l, d, s, config.LogsType)
}
