package dnsmasq

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/dnsmasq_exporter"
)

func init() {
	component.Register(component.Registration{
		Name:          "prometheus.exporter.dnsmasq",
		Args:          Arguments{},
		Exports:       exporter.Exports{},
		NeedsServices: exporter.RequiredServices(),
		Build:         exporter.New(createExporter, "dnsmasq"),
	})
}

func createExporter(opts component.Options, args component.Arguments, defaultInstanceKey string) (integrations.Integration, string, error) {
	a := args.(Arguments)
	return integrations.NewIntegrationWithInstanceKey(opts.Logger, a.Convert(), defaultInstanceKey)
}

// DefaultArguments holds the default arguments for the prometheus.exporter.dnsmasq component.
var DefaultArguments = Arguments{
	Address:      "localhost:53",
	LeasesFile:   "/var/lib/misc/dnsmasq.leases",
	ExposeLeases: false,
}

// Arguments configures the prometheus.exporter.dnsmasq component.
type Arguments struct {
	// Address is the address of the dnsmasq server to connect to (host:port).
	Address string `river:"address,attr,optional"`

	// LeasesFile is the path to the dnsmasq leases file.
	LeasesFile string `river:"leases_file,attr,optional"`

	// ExposeLeases controls whether expose dnsmasq leases as metrics (high cardinality).
	ExposeLeases bool `river:"expose_leases,attr,optional"`
}

// SetToDefault implements river.Defaulter.
func (a *Arguments) SetToDefault() {
	*a = DefaultArguments
}

func (a Arguments) Convert() *dnsmasq_exporter.Config {
	return &dnsmasq_exporter.Config{
		DnsmasqAddress: a.Address,
		LeasesPath:     a.LeasesFile,
		ExposeLeases:   a.ExposeLeases,
	}
}
