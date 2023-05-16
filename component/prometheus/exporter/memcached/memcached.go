package memcached

import (
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/memcached_exporter"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.memcached",
		Args:    Arguments{},
		Exports: exporter.Exports{},
		Build:   exporter.NewWithTargetBuilder(createExporter, "memcached", customizeTarget),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return a.Convert().NewIntegration(opts.Logger)
}

func customizeTarget(baseTarget discovery.Target, args component.Arguments) []discovery.Target {
	a := args.(Arguments)
	target := baseTarget

	target["instance"] = a.Address
	return []discovery.Target{target}
}

// DefaultArguments holds the default arguments for the prometheus.exporter.memcached component.
var DefaultArguments = Arguments{
	Address: "localhost:11211",
	Timeout: time.Second,
}

// Arguments configures the prometheus.exporter.memcached component.
type Arguments struct {
	// Address is the address of the memcached server to connect to (host:port).
	Address string `river:"address,attr,optional"`

	// Timeout is the timeout for the memcached exporter to use when connecting to the
	// memcached server.
	Timeout time.Duration `river:"timeout,attr,optional"`
}

// UnmarshalRiver implements River unmarshalling for Arguments.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = DefaultArguments

	type args Arguments
	return f((*args)(a))
}

func (a Arguments) Convert() *memcached_exporter.Config {
	return &memcached_exporter.Config{
		MemcachedAddress: a.Address,
		Timeout:          a.Timeout,
	}
}
