package node_exporter //nolint:golint

import "github.com/grafana/agent/pkg/integrations"

func init() {
	// Register macos_exporter
	integrations.RegisterIntegration(&DarwinConfig{})
}

// DarwinConfig extends the Config struct and overrides the name of
// the integration to avoid conflicts with node_exporter integration.
type DarwinConfig struct {
	Config
}

// Name returns the name of the integration that this config represents.
func (*DarwinConfig) Name() string {
	return "macos_exporter"
}
