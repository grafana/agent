package memcached

import (
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/memcached_exporter"
)

func init() {
	component.Register(component.Registration{
		Name:    "memcached_exporter",
		Args:    Arguments{},
		Exports: exporter.Exports{},
		Build:   exporter.New(createExporter, "memcached"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	cfg := args.(Arguments)
	return cfg.Convert().NewIntegration(opts.Logger)
}

// DefaultArguments holds the default arguments for the prometheus.exporter.memcached component.
var DefaultArguments = Arguments{
	MemcachedAddress: "localhost:11211",
	Timeout:          time.Second,
}

// Arguments configures the prometheus.exporter.memcached component.
type Arguments struct {
	// MemcachedAddress is the address of the memcached server to connect to.
	MemcachedAddress string `river:"memcached_address,attr,optional"`

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
		MemcachedAddress: a.MemcachedAddress,
		Timeout:          a.Timeout,
	}
}
