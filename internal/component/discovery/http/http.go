package http

import (
	"time"

	"github.com/grafana/agent/internal/component"
	"github.com/grafana/agent/internal/component/common/config"
	"github.com/grafana/agent/internal/component/discovery"
	"github.com/grafana/agent/internal/featuregate"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/http"
)

func init() {
	component.Register(component.Registration{
		Name:      "discovery.http",
		Stability: featuregate.StabilityStable,
		Args:      Arguments{},
		Exports:   discovery.Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return discovery.NewFromConvertibleConfig(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
	HTTPClientConfig config.HTTPClientConfig `river:",squash"`
	RefreshInterval  time.Duration           `river:"refresh_interval,attr,optional"`
	URL              config.URL              `river:"url,attr"`
}

var DefaultArguments = Arguments{
	RefreshInterval:  60 * time.Second,
	HTTPClientConfig: config.DefaultHTTPClientConfig,
}

func (args *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultArguments

	type arguments Arguments
	if err := f((*arguments)(args)); err != nil {
		return err
	}

	return nil
}

func (args Arguments) Convert() discovery.DiscovererConfig {
	cfg := &http.SDConfig{
		HTTPClientConfig: *args.HTTPClientConfig.Convert(),
		URL:              args.URL.String(),
		RefreshInterval:  model.Duration(args.RefreshInterval),
	}
	return cfg
}
