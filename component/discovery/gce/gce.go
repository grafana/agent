// Package gce implements the discovery.gce component.
package gce

import (
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/river"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/gce"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.gce",
		Args:    Arguments{},
		Exports: discovery.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments configures the discovery.gce component.
type Arguments struct {
	Project         string        `river:"project,attr"`
	Zone            string        `river:"zone,attr"`
	Filter          string        `river:"filter,attr,optional"`
	RefreshInterval time.Duration `river:"refresh_interval,attr,optional"`
	Port            int           `river:"port,attr,optional"`
	TagSeparator    string        `river:"tag_separator,attr,optional"`
}

// DefaultArguments holds default values for Arguments.
var DefaultArguments = Arguments{
	Port:            80,
	TagSeparator:    ",",
	RefreshInterval: time.Duration(60 * time.Second),
}

var _ river.Unmarshaler = (*Arguments)(nil)

// UnmarshalRiver implements river.Unmarshaler, applying defaults and
// validating the provided config.
func (args *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultArguments

	type arguments Arguments
	if err := f((*arguments)(args)); err != nil {
		return err
	}
	return nil
}

// Convert converts Arguments to the upstream Prometheus SD type.
func (args Arguments) Convert() gce.SDConfig {
	return gce.SDConfig{
		Project:         args.Project,
		Zone:            args.Zone,
		Filter:          args.Filter,
		RefreshInterval: model.Duration(args.RefreshInterval),
		Port:            args.Port,
		TagSeparator:    args.TagSeparator,
	}

}

// New returns a new instance of a discovery.gce component.
func New(opts component.Options, args Arguments) (component.Component, error) {
	return discovery.New(opts, args, func(args component.Arguments) (discovery.Discoverer, error) {
		conf := args.(Arguments).Convert()
		return gce.NewDiscovery(conf, opts.Logger)
	})
}
