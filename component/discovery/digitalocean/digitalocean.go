package digitalocean

import (
	"fmt"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	"github.com/prometheus/common/model"
	prom_discovery "github.com/prometheus/prometheus/discovery/digitalocean"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.digitalocean",
		Args:    Arguments{},
		Exports: discovery.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
	RefreshInterval  time.Duration           `river:"refresh_interval,attr,optional"`
	Port             int                     `river:"port,attr,optional"`
	HTTPClientConfig config.HTTPClientConfig `river:",squash"`
}

var DefaultArguments = Arguments{
	Port:             80,
	RefreshInterval:  time.Minute,
	HTTPClientConfig: config.DefaultHTTPClientConfig,
}

func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = DefaultArguments
	type arguments Arguments
	err := f((*arguments)(a))
	if err != nil {
		return err
	}
	return a.Validate()
}

func (a *Arguments) Validate() error {
	httpClientConfig := a.HTTPClientConfig
	if httpClientConfig.BearerToken == "" && httpClientConfig.BearerTokenFile == "" {
		return fmt.Errorf("digitalocean uses bearer tokens to authenticate with the API, bearer token or bearer token file must be specified")
	}
	return httpClientConfig.Validate()
}

func (a *Arguments) Convert() *prom_discovery.SDConfig {
	return &prom_discovery.SDConfig{
		RefreshInterval:  model.Duration(a.RefreshInterval),
		Port:             a.Port,
		HTTPClientConfig: *a.HTTPClientConfig.Convert(),
	}
}

func New(opts component.Options, args Arguments) (component.Component, error) {
	return discovery.New(opts, args, func(args component.Arguments) (discovery.Discoverer, error) {
		newArgs := args.(Arguments)
		return prom_discovery.NewDiscovery(newArgs.Convert(), opts.Logger)
	})
}
