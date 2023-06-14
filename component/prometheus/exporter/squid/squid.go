package squid

import (
	"errors"
	"net"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/squid_exporter"
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/prometheus/common/config"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.squid",
		Args:    Arguments{},
		Exports: exporter.Exports{},
		Build:   exporter.NewWithTargetBuilder(createExporter, "squid", customizeTarget),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return a.Convert().NewIntegration(opts.Logger)
}

func customizeTarget(baseTarget discovery.Target, args component.Arguments) []discovery.Target {
	a := args.(Arguments)
	target := baseTarget

	target["instance"] = a.SquidAddr
	return []discovery.Target{target}
}

// DefaultArguments holds the default settings for the squid exporter
var DefaultArguments = Arguments{
	SquidAddr: "localhost:3128",
}

var (
	errNoAddress  = errors.New("no address was provided")
	errNoHostname = errors.New("no hostname in provided address")
	errNoPort     = errors.New("no port in provided address")
)

// Arguments controls the squid exporter.
type Arguments struct {
	SquidAddr     string            `river:"address,attr"`
	SquidUser     string            `river:"username,attr,optional"`
	SquidPassword rivertypes.Secret `river:"password,attr,optional"`
}

// SetToDefault implements river.Defaulter.
func (a *Arguments) SetToDefault() {
	*a = DefaultArguments
}

// Validate implements river.Validator.
func (a *Arguments) Validate() error {
	if a.SquidAddr == "" {
		return errNoAddress
	}

	host, port, err := net.SplitHostPort(a.SquidAddr)
	if err != nil {
		return err
	}

	if host == "" {
		return errNoHostname
	}

	if port == "" {
		return errNoPort
	}

	return nil
}

func (a *Arguments) Convert() *squid_exporter.Config {
	return &squid_exporter.Config{
		Address:  a.SquidAddr,
		Username: a.SquidUser,
		Password: config.Secret(a.SquidPassword),
	}
}
