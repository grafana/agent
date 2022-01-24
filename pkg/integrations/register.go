package integrations

import (
	"github.com/grafana/agent/pkg/integrations/config"
)

// Configs is a list of UnmarshaledConfig. Configs for integrations which are
// unmarshaled from YAML are combined with common settings.
type Configs []UnmarshaledConfig

// UnmarshaledConfig combines an integration's config with common settings.
type UnmarshaledConfig struct {
	Config
	Common config.Common
}
