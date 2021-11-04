package install

import (
	"testing"

	"github.com/grafana/agent/pkg/integrations"
)

// TestRegisteredIntegrations runs the integration tests suite for all
// registered integrations.
func TestRegisteredIntegrations(t *testing.T) {
	for _, cfg := range integrations.RegisteredIntegrations() {
		t.Run(cfg.Name(), func(t *testing.T) {
			integrations.TestIntegration(t, cfg)
		})
	}
}
