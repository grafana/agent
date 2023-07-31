package hetzner

import (
	"fmt"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	"github.com/prometheus/common/model"
	prom_discovery "github.com/prometheus/prometheus/discovery/hetzner"
)

// TODO(marctc): hetzner role constants are not exported, we need to manual create them
// Remove once this is merged: https://github.com/prometheus/prometheus/pull/12620
const (
	hetznerRoleRobot  string = "robot"
	hetznerRoleHcloud string = "hcloud"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.hetzner",
		Args:    Arguments{},
		Exports: discovery.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
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
	case hetznerRoleRobot, hetznerRoleHcloud:
	default:
		return fmt.Errorf("unknown role %s, must be one of robot or hcloud", args.Role)
	}
	return args.HTTPClientConfig.Validate()
}

func (args *Arguments) Convert() *prom_discovery.SDConfig {
	httpClient := &args.HTTPClientConfig

	cfg := &prom_discovery.SDConfig{
		RefreshInterval:  model.Duration(args.RefreshInterval),
		Port:             args.Port,
		HTTPClientConfig: *httpClient.Convert(),
	}
	// TODO(marctc): hetzner.role is not exported, we need to manual convert it
	// Remove once this is merged: https://github.com/prometheus/prometheus/pull/12620
	if args.Role == hetznerRoleRobot {
		cfg.Role = "robot"
	} else if args.Role == hetznerRoleHcloud {
		cfg.Role = "hcloud"
	}
	return cfg
}

// New returns a new instance of a discovery.hetzner component.
func New(opts component.Options, args Arguments) (*discovery.Component, error) {
	return discovery.New(opts, args, func(args component.Arguments) (discovery.Discoverer, error) {
		newArgs := args.(Arguments)
		return prom_discovery.NewDiscovery(newArgs.Convert(), opts.Logger)
	})
}
