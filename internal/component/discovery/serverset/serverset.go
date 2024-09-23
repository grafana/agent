package serverset

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/grafana/agent/internal/component"
	"github.com/grafana/agent/internal/component/discovery"
	"github.com/grafana/agent/internal/featuregate"
	"github.com/prometheus/common/model"
	prom_discovery "github.com/prometheus/prometheus/discovery/zookeeper"
)

func init() {
	component.Register(component.Registration{
		Name:      "discovery.serverset",
		Stability: featuregate.StabilityStable,
		Args:      Arguments{},
		Exports:   discovery.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return discovery.NewFromConvertibleConfig(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
	Servers []string      `river:"servers,attr"`
	Paths   []string      `river:"paths,attr"`
	Timeout time.Duration `river:"timeout,attr,optional"`
}

var DefaultArguments = Arguments{
	Timeout: 10 * time.Second,
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Validate implements river.Validator.
func (args *Arguments) Validate() error {
	if len(args.Servers) == 0 {
		return errors.New("discovery.serverset config must contain at least one Zookeeper server")
	}
	if len(args.Paths) == 0 {
		return errors.New("discovery.serverset config must contain at least one path")
	}
	for _, path := range args.Paths {
		if !strings.HasPrefix(path, "/") {
			return fmt.Errorf("discovery.serverset config paths must begin with '/': %s", path)
		}
	}
	return nil
}

func (args Arguments) Convert() discovery.DiscovererConfig {
	return &prom_discovery.ServersetSDConfig{
		Servers: args.Servers,
		Paths:   args.Paths,
		Timeout: model.Duration(args.Timeout),
	}
}
