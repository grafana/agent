// Package integrations exposes the integrations subsystem. It will select
// between v1 and v2 based on a field.
package integrations

import (
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	v1 "github.com/grafana/agent/pkg/integrations"
	v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/metrics"
	"github.com/weaveworks/common/server"
	"gopkg.in/yaml.v2"
)

type Version int

const (
	VersionDefault Version = 0

	Version1 Version = iota
	Version2
)

// Config abstracts the subsystem configs for integrations v1 and v2.
type Config struct {
	Version Version

	configV1 *v1.ManagerConfig
	configV2 *v2.SubsystemOptions
}

// init will initialize the inner config based on the set version.
func (c *Config) init() {
	switch c.Version {
	case VersionDefault, Version1:
		if c.configV1 == nil {
			val := v1.DefaultManagerConfig
			c.configV1 = &val
		}
	case Version2:
		if c.configV2 == nil {
			val := v2.DefaultSubsystemOptions
			c.configV2 = &val
		}
	}
}

var (
	_ yaml.Unmarshaler = (*Config)(nil)
	_ yaml.Marshaler   = (*Config)(nil)
)

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.init()

	if c.Version != Version2 {
		return unmarshal(&c.configV1)
	}
	return unmarshal(&c.configV2)
}

// MarshalYAML implements yaml.Marshaler.
func (c Config) MarshalYAML() (interface{}, error) {
	c.init()

	if c.Version != Version2 {
		return c.configV1, nil
	}
	return c.configV2, nil
}

// ApplyDefaults applies defaults to the subsystem based on globals.
func (c *Config) ApplyDefaults(scfg *server.Config, mcfg *metrics.Config) error {
	c.init()

	if c.Version != Version2 {
		return c.configV1.ApplyDefaults(scfg, mcfg)
	}

	if len(c.configV2.PrometheusRemoteWrite) == 0 {
		c.configV2.PrometheusRemoteWrite = mcfg.Global.RemoteWrite
	}

	return nil
}

type Globals = v2.Globals

// Subsystem is an abstraction over both the v1 and v2 systems.
type Subsystem interface {
	ApplyConfig(*Config, Globals) error
	WireAPI(*mux.Router)
	Stop()
}

// NewSubsystem creates a new subsystem. globals should be provided regardless
// of useV2. globals.SubsystemOptions will be automatically set if cfg.Version
// is set to Version2.
func NewSubsystem(logger log.Logger, cfg *Config, globals Globals) (Subsystem, error) {
	cfg.init()

	if cfg.Version != Version2 {
		instance, err := v1.NewManager(*cfg.configV1, logger, globals.Metrics.InstanceManager(), globals.Metrics.Validate)
		if err != nil {
			return nil, err
		}
		return &v1Subsystem{Manager: instance}, nil
	}

	globals.SubsystemOpts = *cfg.configV2
	instance, err := v2.NewSubsystem(logger, globals)
	if err != nil {
		return nil, err
	}
	return &v2Subsystem{Subsystem: instance}, nil
}

type v1Subsystem struct{ *v1.Manager }

func (s *v1Subsystem) ApplyConfig(cfg *Config, globals Globals) error {
	return s.Manager.ApplyConfig(*cfg.configV1)
}

type v2Subsystem struct{ *v2.Subsystem }

func (s *v2Subsystem) ApplyConfig(cfg *Config, globals Globals) error {
	globals.SubsystemOpts = *cfg.configV2
	return s.Subsystem.ApplyConfig(globals)
}
