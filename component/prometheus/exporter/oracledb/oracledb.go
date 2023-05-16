package oracledb

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/oracledb_exporter"
	"github.com/grafana/agent/pkg/river/rivertypes"
	config_util "github.com/prometheus/common/config"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.oracledb",
		Args:    Arguments{},
		Exports: exporter.Exports{},
		Build:   exporter.NewWithTargetBuilder(createExporter, "oracledb", customizeTarget),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return a.Convert().NewIntegration(opts.Logger)
}

func customizeTarget(baseTarget discovery.Target, args component.Arguments) []discovery.Target {
	a := args.(Arguments)
	target := baseTarget

	url, err := url.Parse(string(a.ConnectionString))
	if err != nil {
		return []discovery.Target{target}
	}

	target["instance"] = url.Host
	return []discovery.Target{target}
}

// DefaultArguments holds the default settings for the oracledb exporter
var DefaultArguments = Arguments{
	MaxIdleConns: 0,
	MaxOpenConns: 10,
	QueryTimeout: 5,
}

var (
	errNoConnectionString = errors.New("no connection string was provided")
	errNoHostname         = errors.New("no hostname in connection string")
)

// Arguments controls the oracledb exporter.
type Arguments struct {
	ConnectionString rivertypes.Secret `river:"connection_string,attr"`
	MaxIdleConns     int               `river:"max_idle_conns,attr,optional"`
	MaxOpenConns     int               `river:"max_open_conns,attr,optional"`
	QueryTimeout     int               `river:"query_timeout,attr,optional"`
}

// UnmarshalRiver implements River unmarshalling for Arguments.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = DefaultArguments

	type args Arguments
	if err := f((*args)(a)); err != nil {
		return err
	}
	return a.Validate()
}

func (a *Arguments) Validate() error {
	if a.ConnectionString == "" {
		return errNoConnectionString
	}
	u, err := url.Parse(string(a.ConnectionString))
	if err != nil {
		return fmt.Errorf("unable to parse connection string: %w", err)
	}

	if u.Scheme != "oracle" {
		return fmt.Errorf("unexpected scheme of type '%s'. Was expecting 'oracle': %w", u.Scheme, err)
	}

	// hostname is required for identification
	if u.Hostname() == "" {
		return errNoHostname
	}
	return nil
}

func (a *Arguments) Convert() *oracledb_exporter.Config {
	return &oracledb_exporter.Config{
		ConnectionString: config_util.Secret(a.ConnectionString),
		MaxIdleConns:     a.MaxIdleConns,
		MaxOpenConns:     a.MaxOpenConns,
		QueryTimeout:     a.QueryTimeout,
	}
}
