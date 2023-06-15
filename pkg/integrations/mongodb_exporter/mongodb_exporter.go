package mongodb_exporter //nolint:golint

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-kit/log"
	config_util "github.com/prometheus/common/config"

	"github.com/grafana/agent/pkg/integrations"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
)

// Config controls mongodb_exporter
type Config struct {
	// MongoDB connection URI. example:mongodb://user:pass@127.0.0.1:27017/admin?ssl=true"
	URI config_util.Secret `yaml:"mongodb_uri"`
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

// InstanceKey returns the address:port of the mongodb server being queried.
func (c *Config) InstanceKey(_ string) (string, error) {
	u, err := url.Parse(string(c.URI))
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
	integrations_v2.RegisterLegacy(&Config{}, integrations_v2.TypeMultiplex, metricsutils.NewNamedShim("mongodb"))
}

// New creates a new mongodb_exporter integration.
func New(logger log.Logger, c *Config) (integrations.Integration, error) {
	// TODO(jcreixell): Re-implement conflicting bits and briong the exporter back
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "mongdb exporter has been disabled for the time being due to licensing issues", http.StatusBadRequest)
	})

	return integrations.NewHandlerIntegration(c.Name(), handler), nil
}
