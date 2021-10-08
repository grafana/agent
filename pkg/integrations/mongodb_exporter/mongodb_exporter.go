package mongodb_exporter //nolint:golint

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/percona/mongodb_exporter/exporter"
)

// Config controls mongodb_exporter
type Config struct {
	Common config.Common `yaml:",inline"`

	// MongoDB connection URI. example:mongodb://user:pass@127.0.0.1:27017/admin?ssl=true"
	URI string `yaml:"mongodb_uri"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string {
	return "mongodb_exporter"
}

// CommonConfig returns the common settings shared across all configs for
// integrations.
func (c *Config) CommonConfig() config.Common {
	return c.Common
}

// NewIntegration creates a new mongodb_exporter
func (c *Config) NewIntegration(logger log.Logger) (integrations.Integration, error) {
	return New(logger, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// New creates a new mongodb_exporter integration.
func New(logger log.Logger, c *Config) (integrations.Integration, error) {
	logrusLogger := NewLogger(logger)

	exp, err := exporter.New(&exporter.Opts{
		URI:                    c.URI,
		Logger:                 logrusLogger,
		DisableDefaultRegistry: true,
		CompatibleMode: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create mongodb_exporter: %w", err)
	}

	return integrations.NewHandlerIntegration(c.Name(), exp.Handler()), nil
}
