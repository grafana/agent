package config

import (
	"reflect"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	v1 "github.com/grafana/agent/pkg/integrations"
	v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/metrics"
	"github.com/weaveworks/common/server"
	"gopkg.in/yaml.v2"
)

type integrationsVersion int

const (
	integrationsVersionDefault integrationsVersion = 0

	integrationsVersion1 integrationsVersion = iota
	integrationsVersion2
)

// VersionedIntegrations abstracts the subsystem configs for integrations v1 and v2.
type VersionedIntegrations struct {
	version integrationsVersion

	configV1 *v1.ManagerConfig
	configV2 *v2.SubsystemOptions
}

// init will initialize the inner config based on the set version.
func (c *VersionedIntegrations) init() {
	switch c.version {
	case integrationsVersionDefault, integrationsVersion1:
		if c.configV1 == nil {
			val := v1.DefaultManagerConfig
			c.configV1 = &val
		}
	case integrationsVersion2:
		if c.configV2 == nil {
			val := v2.DefaultSubsystemOptions
			c.configV2 = &val
		}
	}
}

var (
	_ yaml.Unmarshaler = (*VersionedIntegrations)(nil)
	_ yaml.Marshaler   = (*VersionedIntegrations)(nil)
)

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *VersionedIntegrations) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.init()

	if c.version != integrationsVersion2 {
		return unmarshal(&c.configV1)
	}
	return unmarshal(&c.configV2)
}

// MarshalYAML implements yaml.Marshaler.
func (c VersionedIntegrations) MarshalYAML() (interface{}, error) {
	c.init()

	if c.version != integrationsVersion2 {
		return c.configV1, nil
	}
	return c.configV2, nil
}

// IsZero implements yaml.IsZeroer.
func (c VersionedIntegrations) IsZero() bool {
	switch {
	case c.configV1 != nil:
		return reflect.ValueOf(*c.configV1).IsZero()
	case c.configV2 != nil:
		return reflect.ValueOf(*c.configV2).IsZero()
	default:
		return true
	}
}

// ApplyDefaults applies defaults to the subsystem based on globals.
func (c *VersionedIntegrations) ApplyDefaults(scfg *server.Config, mcfg *metrics.Config) error {
	c.init()

	if c.version != integrationsVersion2 {
		return c.configV1.ApplyDefaults(scfg, mcfg)
	}

	if len(c.configV2.PrometheusRemoteWrite) == 0 {
		c.configV2.PrometheusRemoteWrite = mcfg.Global.RemoteWrite
	}

	return nil
}

// IntegrationsGlobals is a global struct shared across integrations.
type IntegrationsGlobals = v2.Globals

// Integrations is an abstraction over both the v1 and v2 systems.
type Integrations interface {
	ApplyConfig(*VersionedIntegrations, IntegrationsGlobals) error
	WireAPI(*mux.Router)
	Stop()
}

// NewIntegrations creates a new subsystem. globals should be provided regardless
// of useV2. globals.SubsystemOptions will be automatically set if cfg.Version
// is set to IntegrationsVersion2.
func NewIntegrations(logger log.Logger, cfg *VersionedIntegrations, globals IntegrationsGlobals) (Integrations, error) {
	cfg.init()

	if cfg.version != integrationsVersion2 {
		instance, err := v1.NewManager(*cfg.configV1, logger, globals.Metrics.InstanceManager(), globals.Metrics.Validate)
		if err != nil {
			return nil, err
		}
		return &v1Integrations{Manager: instance}, nil
	}

	globals.SubsystemOpts = *cfg.configV2
	instance, err := v2.NewSubsystem(logger, globals)
	if err != nil {
		return nil, err
	}
	return &v2Integrations{Subsystem: instance}, nil
}

type v1Integrations struct{ *v1.Manager }

func (s *v1Integrations) ApplyConfig(cfg *VersionedIntegrations, globals IntegrationsGlobals) error {
	return s.Manager.ApplyConfig(*cfg.configV1)
}

type v2Integrations struct{ *v2.Subsystem }

func (s *v2Integrations) ApplyConfig(cfg *VersionedIntegrations, globals IntegrationsGlobals) error {
	globals.SubsystemOpts = *cfg.configV2
	return s.Subsystem.ApplyConfig(globals)
}
