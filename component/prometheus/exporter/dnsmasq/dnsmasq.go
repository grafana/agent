package dnsmasq

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/dnsmasq_exporter"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.dnsmasq",
		Args:    Arguments{},
		Exports: exporter.Exports{},
		Build:   exporter.New(createExporter, "dnsmasq"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return a.Convert().NewIntegration(opts.Logger)
}

// DefaultArguments holds the default arguments for the prometheus.exporter.dnsmasq component.
var DefaultArguments = Arguments{
	Address:    "localhost:53",
	LeasesFile: "/var/lib/misc/dnsmasq.leases",
}

// Arguments configures the prometheus.exporter.dnsmasq component.
type Arguments struct {
	// Address is the address of the dnsmasq server to connect to (host:port).
	Address string `river:"address,attr,optional"`

	// LeasesFile is the path to the dnsmasq leases file.
	LeasesFile string `river:"leases_file,attr,optional"`
}

// UnmarshalRiver implements River unmarshalling for Arguments.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = DefaultArguments

	type args Arguments
	return f((*args)(a))
}

func (a Arguments) Convert() *dnsmasq_exporter.Config {
	return &dnsmasq_exporter.Config{
		DnsmasqAddress: a.Address,
		LeasesPath:     a.LeasesFile,
	}
}
