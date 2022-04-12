// Package snmp_exporter embeds https://github.com/prometheus/snmp_exporter
package snmp_exporter

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	snmp_common "github.com/grafana/agent/pkg/integrations/snmp_exporter/common"
	snmp_config "github.com/prometheus/snmp_exporter/config"
)

// DefaultConfig holds the default settings for the snmp_exporter integration.
var DefaultConfig = Config{
	WalkParams:     make(map[string]snmp_config.WalkParams),
	SnmpConfigFile: "",
	SnmpTargets:    make([]SNMPTarget, 0),
}

// SNMPTarget defines a target device to be used by the integration.
type SNMPTarget struct {
	Name       string `yaml:"name"`
	Target     string `yaml:"address"`
	Module     string `yaml:"module"`
	WalkParams string `yaml:"walk_params,omitempty"`
}

// Config configures the SNMP integration.
type Config struct {
	WalkParams     map[string]snmp_config.WalkParams `yaml:"walk_params,omitempty"`
	SnmpConfigFile string                            `yaml:"config_file,omitempty"`
	SnmpTargets    []SNMPTarget                      `yaml:"snmp_targets"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration.
func (c *Config) Name() string {
	return "snmp"
}

// InstanceKey returns the hostname:port of the agent.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return agentKey, nil
}

// NewIntegration creates a new SNMP integration.
func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	return New(l, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// New creates a new snmp_exporter integration
func New(log log.Logger, c *Config) (integrations.Integration, error) {
	var modules *snmp_config.Config
	var err error
	if c.SnmpConfigFile != "" {
		modules, err = snmp_config.LoadFile(c.SnmpConfigFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load snmp config from file %v: %w", c.SnmpConfigFile, err)
		}
	} else {
		modules, err = snmp_common.LoadEmbeddedConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load embedded snmp config: %w", err)
		}
	}

	// The `name` and `address` fields are mandatory for the SNMP targets are mandatory.
	// Enforce this check and fail the creation of the integration if they're missing.
	for _, target := range c.SnmpTargets {
		if target.Name == "" || target.Target == "" {
			return nil, fmt.Errorf("failed to load snmp_targets; the `name` and `address` fields are mandatory")
		}
	}

	sh := &snmpHandler{
		cfg:     c,
		modules: modules,
		log:     log,
	}
	integration := &Integration{
		sh: sh,
	}

	return integration, nil
}

// Integration is the node_exporter integration. The integration scrapes metrics
// from the host Linux-based system.
type Integration struct {
	sh *snmpHandler
}

// MetricsHandler implements Integration.
func (i *Integration) MetricsHandler() (http.Handler, error) {
	return i.sh, nil
}

// Run satisfies Integration.Run.
func (i *Integration) Run(ctx context.Context) error {
	// We don't need to do anything here, so we can just wait for the context to
	// finish.
	<-ctx.Done()
	return ctx.Err()
}

// ScrapeConfigs satisfies Integration.ScrapeConfigs.
func (i *Integration) ScrapeConfigs() []config.ScrapeConfig {
	var res []config.ScrapeConfig
	for _, target := range i.sh.cfg.SnmpTargets {
		queryParams := url.Values{}
		queryParams.Add("target", target.Target)
		res = append(res, config.ScrapeConfig{
			JobName:     i.sh.cfg.Name() + "/" + target.Name,
			MetricsPath: "/metrics",
			QueryParams: queryParams,
		})
	}
	return res
}
