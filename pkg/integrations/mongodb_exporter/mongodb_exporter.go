package mongodb_exporter //nolint:golint

import (
	"fmt"
	"net/url"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/percona/mongodb_exporter/exporter"
)

// Config controls mongodb_exporter
type Config struct {
	config.Common `yaml:",inline"`

	// MongoDB connection URI. example:mongodb://user:pass@127.0.0.1:27017/admin?ssl=true"
	URI string `yaml:"mongodb_uri"`
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

// InstanceKey returns the address:port of the mongodb server being queried.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	u, err := url.Parse(c.URI)
	if err != nil {
		return "", fmt.Errorf("could not parse url: %w", err)
	}
	return u.Host, nil
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

		// NOTE(rfratto): CompatibleMode configures the exporter to use old metric
		// names from mongodb_exporter <v0.20.0. Many existing dashboards rely on
		// the old names, so we hard-code it to true now. We may wish to make this
		// configurable in the future.
		CompatibleMode: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create mongodb_exporter: %w", err)
	}

	return integrations.NewHandlerIntegration(c.Name(), exp.Handler()), nil
}
