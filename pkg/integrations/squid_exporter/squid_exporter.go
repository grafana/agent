package squid_exporter

import (
	"errors"
	"net"
	"strconv"

	se "github.com/boynux/squid-exporter/collector"
	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/prometheus/client_golang/prometheus"

	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
)

// DefaultConfig is the default config for the squid integration
var DefaultConfig = Config{
	Address:  "localhost:3128",
	Login:    "",
	Password: "",
}

var (
	// 	errNoConnectionString = errors.New("no connection string was provided")
	errNoHostname = errors.New("no hostname in provided address")
)

// Config is the configuration for the squid integration
type Config struct {
	Address  string `yaml:"address"`
	Login    string `yaml:"username"`
	Password string `yaml:"password"`
}

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

// InstanceKey returns the addr of the squid instance.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return c.Address, nil
}

// NewIntegration returns the Squid Exporter Integration
func (c *Config) NewIntegration(_ log.Logger) (integrations.Integration, error) {
	return New(c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
	integrations_v2.RegisterLegacy(&Config{}, integrations_v2.TypeMultiplex, metricsutils.NewNamedShim("squid"))
}

// New creates a new squid integration. The integration scrapes metrics
// from an Squid exporter running with the https://github.com/boynux/squid-exporter
func New(c *Config) (integrations.Integration, error) {
	cfg := c
	se.ExtractServiceTimes = true

	host, scheme, err := net.SplitHostPort(cfg.Address)
	if err != nil {
		return nil, err
	}
	port, _ := strconv.Atoi(scheme)

	seExporter := se.New(&se.CollectorConfig{
		Hostname: host,
		Port:     port,
		Login:    cfg.Login,
		Password: cfg.Password,
	})

	prometheus.MustRegister(seExporter)

	return integrations.NewCollectorIntegration(c.Name(), integrations.WithCollectors(seExporter)), nil
}
