package squid_exporter

import (
	"fmt"
	"net/url"

	se "github.com/boynux/squid-exporter/collector"
	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations"

	// required driver for integration
	_ "github.com/sijms/go-ora/v2"

	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
)

// DefaultConfig is the default config for the oracledb v2 integration
var DefaultConfig = Config{
	Hostname: "localhost",
	Port:     3128,
	Login:    "",
	Password: "",
	Headers:  []string{},
}

// var (
// 	errNoConnectionString = errors.New("no connection string was provided")
// 	errNoHostname         = errors.New("no hostname in connection string")
// )

// Config is the configuration for the oracledb v2 integration
type Config struct {
	Hostname string   `yaml:"hostname"`
	Port     int      `yaml:"int"`
	Login    string   `yaml:"username"`
	Password string   `yaml:"password"`
	Headers  []string `yaml:"headers"`
}

// ValidateConnString attempts to ensure the connection string supplied is valid
// to connect to an OracleDB instance
// func validateConnString(connStr string) error {
// 	if connStr == "" {
// 		return errNoConnectionString
// 	}
// 	u, err := url.Parse(connStr)
// 	if err != nil {
// 		return fmt.Errorf("unable to parse connection string: %w", err)
// 	}

// 	if u.Scheme != "oracle" {
// 		return fmt.Errorf("unexpected scheme of type '%s'. Was expecting 'oracle': %w", u.Scheme, err)
// 	}

// 	// hostname is required for identification
// 	if u.Hostname() == "" {
// 		return errNoHostname
// 	}
// 	return nil
// }

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the integration name this config is associated with.
func (c *Config) Name() string {
	return "squid"
}

// InstanceKey returns the addr of the oracle instance.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	u, err := url.Parse("")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", u.Hostname(), u.Port()), nil
}

// NewIntegration returns the Squid Exporter Integration
func (c *Config) NewIntegration(logger log.Logger) (integrations.Integration, error) {
	return New(logger, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
	integrations_v2.RegisterLegacy(&Config{}, integrations_v2.TypeMultiplex, metricsutils.NewNamedShim("squid"))
}

// New creates a new squid integration. The integration scrapes metrics
// from an Squid exporter running with the https://github.com/boynux/squid-exporter
func New(logger log.Logger, c *Config) (integrations.Integration, error) {
	seExporter := se.New(&se.CollectorConfig{
		Hostname: c.Hostname,
		Port:     c.Port,
		Login:    c.Login,
		Password: c.Password,
		Headers:  c.Headers,
	})

	return integrations.NewCollectorIntegration(c.Name(), integrations.WithCollectors(seExporter)), nil
}
