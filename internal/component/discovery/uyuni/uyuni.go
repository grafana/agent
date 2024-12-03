package uyuni

import (
	"fmt"
	"net/url"
	"time"

	"github.com/grafana/agent/internal/component"
	"github.com/grafana/agent/internal/component/common/config"
	"github.com/grafana/agent/internal/component/discovery"
	"github.com/grafana/agent/internal/featuregate"
	"github.com/grafana/river/rivertypes"
	promcfg "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	prom_discovery "github.com/prometheus/prometheus/discovery/uyuni"
)

func init() {
	component.Register(component.Registration{
		Name:      "discovery.uyuni",
		Stability: featuregate.StabilityStable,
		Args:      Arguments{},
		Exports:   discovery.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return discovery.NewFromConvertibleConfig(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
	Server          string              `river:"server,attr"`
	Username        string              `river:"username,attr"`
	Password        rivertypes.Secret   `river:"password,attr"`
	Entitlement     string              `river:"entitlement,attr,optional"`
	Separator       string              `river:"separator,attr,optional"`
	RefreshInterval time.Duration       `river:"refresh_interval,attr,optional"`
	ProxyConfig     *config.ProxyConfig `river:",squash"`
	TLSConfig       config.TLSConfig    `river:"tls_config,block,optional"`
	FollowRedirects bool                `river:"follow_redirects,attr,optional"`
	EnableHTTP2     bool                `river:"enable_http2,attr,optional"`
}

var DefaultArguments = Arguments{
	Entitlement:     "monitoring_entitled",
	Separator:       ",",
	RefreshInterval: 1 * time.Minute,

	EnableHTTP2:     config.DefaultHTTPClientConfig.EnableHTTP2,
	FollowRedirects: config.DefaultHTTPClientConfig.FollowRedirects,
}

// SetToDefault implements river.Defaulter.
func (a *Arguments) SetToDefault() {
	*a = DefaultArguments
}

// Validate implements river.Validator.
func (a *Arguments) Validate() error {
	_, err := url.Parse(a.Server)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	if err = a.TLSConfig.Validate(); err != nil {
		return err
	}

	return a.ProxyConfig.Validate()
}

func (a Arguments) Convert() discovery.DiscovererConfig {
	return &prom_discovery.SDConfig{
		Server:          a.Server,
		Username:        a.Username,
		Password:        promcfg.Secret(a.Password),
		Entitlement:     a.Entitlement,
		Separator:       a.Separator,
		RefreshInterval: model.Duration(a.RefreshInterval),

		HTTPClientConfig: promcfg.HTTPClientConfig{
			ProxyConfig:     a.ProxyConfig.Convert(),
			TLSConfig:       *a.TLSConfig.Convert(),
			FollowRedirects: a.FollowRedirects,
			EnableHTTP2:     a.EnableHTTP2,
		},
	}
}
