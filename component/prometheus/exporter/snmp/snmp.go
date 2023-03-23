package snmp

import (
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/flow/rivertypes"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/snmp_exporter"
	snmp_config "github.com/prometheus/snmp_exporter/config"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.snmp",
		Args:    Config{},
		Exports: exporter.Exports{},
		Build:   exporter.NewMultiTarget(createExporter, "snmp", buildSNMPTargets),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	cfg := args.(Config)
	return cfg.Convert().NewIntegration(opts.Logger)
}

// buildSNMPTargets creates the exporter's discovery targets based on the defined SNMP targets.
func buildSNMPTargets(baseTarget discovery.Target, args component.Arguments) []discovery.Target {
	var targets []discovery.Target

	cfg := args.(Config)
	for _, tgt := range cfg.Targets {
		target := make(discovery.Target)
		for k, v := range baseTarget {
			target[k] = v
		}

		target["job"] = target["job"] + "/" + tgt.Name
		target["__param_target"] = tgt.Target
		if tgt.Module != "" {
			target["__param_module"] = tgt.Module
		}
		if tgt.WalkParams != "" {
			target["__param_walk_params"] = tgt.WalkParams
		}

		targets = append(targets, target)
	}

	return targets
}

// DefaultConfig holds non-zero default options for the Config when it is
// unmarshaled from river.
// var DefaultConfig = Config{
// 	ProbeTimeoutOffset: 500 * time.Millisecond,
// }

// SNMPTarget defines a target to be used by the exporter.
type SNMPTarget struct {
	Name       string `river:",label"`
	Target     string `river:"address,attr"`
	Module     string `river:"module,attr,optional"`
	WalkParams string `river:"walk_params,attr,optional"`
}

type TargetBlock []SNMPTarget

// Convert converts the component's TargetBlock to a slice of integration's SNMPTarget.
func (t TargetBlock) Convert() []snmp_exporter.SNMPTarget {
	targets := make([]snmp_exporter.SNMPTarget, 0, len(t))
	for _, target := range t {
		targets = append(targets, snmp_exporter.SNMPTarget{
			Name:       target.Name,
			Target:     target.Target,
			Module:     target.Module,
			WalkParams: target.WalkParams,
		})
	}
	return targets
}

type Auth struct {
	Community     rivertypes.Secret `river:"community,attr,optional"`
	SecurityLevel string            `river:"security_level,attr,optional"`
	Username      string            `river:"username,attr,optional"`
	Password      rivertypes.Secret `river:"password,attr,optional"`
	AuthProtocol  string            `river:"auth_protocol,attr,optional"`
	PrivProtocol  string            `river:"priv_protocol,attr,optional"`
	PrivPassword  rivertypes.Secret `river:"priv_password,attr,optional"`
	ContextName   string            `river:"context_name,attr,optional"`
}

// Convert converts the component's Auth to the integration's Auth.
func (a Auth) Convert() snmp_config.Auth {
	return snmp_config.Auth{
		Community:     snmp_config.Secret(a.Community),
		SecurityLevel: a.SecurityLevel,
		Username:      a.Username,
		Password:      snmp_config.Secret(a.Password),
		AuthProtocol:  a.AuthProtocol,
		PrivProtocol:  a.PrivProtocol,
		PrivPassword:  snmp_config.Secret(a.PrivPassword),
		ContextName:   a.ContextName,
	}
}

type WalkParam struct {
	Name                    string        `river:",label"`
	Version                 int           `river:"version,attr,optional"`
	MaxRepetitions          uint32        `river:"max_repetitions,attr,optional"`
	Retries                 int           `river:"retries,attr,optional"`
	Timeout                 time.Duration `river:"timeout,attr,optional"`
	Auth                    Auth          `river:"auth,block,optional"`
	UseUnconnectedUDPSocket bool          `river:"use_unconnected_udp_socket,attr,optional"`
}

type WalkParams []WalkParam

// Convert converts the component's WalkParams to the integration's WalkParams.
func (w WalkParams) Convert() map[string]snmp_config.WalkParams {
	walkParams := make(map[string]snmp_config.WalkParams)
	for _, walkParam := range w {
		walkParams[walkParam.Name] = snmp_config.WalkParams{
			Version:                 walkParam.Version,
			MaxRepetitions:          walkParam.MaxRepetitions,
			Retries:                 walkParam.Retries,
			Timeout:                 walkParam.Timeout,
			Auth:                    walkParam.Auth.Convert(),
			UseUnconnectedUDPSocket: walkParam.UseUnconnectedUDPSocket,
		}
	}
	return walkParams
}

type Config struct {
	ConfigFile string      `river:"config_file,attr"`
	Targets    TargetBlock `river:"target,block"`
	WalkParams WalkParams  `river:"walk_param,block,optional"`
}

// UnmarshalRiver implements River unmarshalling for Config.
func (c *Config) UnmarshalRiver(f func(interface{}) error) error {
	//*c = DefaultConfig

	type cfg Config
	return f((*cfg)(c))
}

// Convert converts the component's Config to the integration's Config.
func (c *Config) Convert() *snmp_exporter.Config {
	return &snmp_exporter.Config{
		SnmpConfigFile: c.ConfigFile,
		SnmpTargets:    c.Targets.Convert(),
		WalkParams:     c.WalkParams.Convert(),
	}
}
