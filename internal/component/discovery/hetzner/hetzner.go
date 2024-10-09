package hetzner

import (
	"fmt"
	"time"

	"github.com/grafana/agent/internal/component"
	"github.com/grafana/agent/internal/component/common/config"
	"github.com/grafana/agent/internal/component/discovery"
	"github.com/grafana/agent/internal/featuregate"
	"github.com/prometheus/common/model"
	prom_discovery "github.com/prometheus/prometheus/discovery/hetzner"
)

func init() {
	component.Register(component.Registration{
		Name:      "discovery.hetzner",
		Stability: featuregate.StabilityStable,
		Args:      Arguments{},
		Exports:   discovery.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return discovery.NewFromConvertibleConfig(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
	Role             string                  `river:"role,attr"`
	RefreshInterval  time.Duration           `river:"refresh_interval,attr,optional"`
	Port             int                     `river:"port,attr,optional"`
	HTTPClientConfig config.HTTPClientConfig `river:",squash"`
}

var DefaultArguments = Arguments{
	Port:             80,
	RefreshInterval:  60 * time.Second,
	HTTPClientConfig: config.DefaultHTTPClientConfig,
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Validate implements river.Validator.
func (args *Arguments) Validate() error {
	switch args.Role {
	case string(prom_discovery.HetznerRoleRobot), string(prom_discovery.HetznerRoleHcloud):
	default:
		return fmt.Errorf("unknown role %s, must be one of robot or hcloud", args.Role)
	}
	return args.HTTPClientConfig.Validate()
}

func (args Arguments) Convert() discovery.DiscovererConfig {
	httpClient := &args.HTTPClientConfig

	cfg := &prom_discovery.SDConfig{
		RefreshInterval:  model.Duration(args.RefreshInterval),
		Port:             args.Port,
		HTTPClientConfig: *httpClient.Convert(),
		Role:             prom_discovery.Role(args.Role),
	}
	return cfg
}
