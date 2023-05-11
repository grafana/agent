package oracledb

import (
	"github.com/grafana/agent/component"
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
		Build:   exporter.New(createExporter, "oracledb"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return a.Convert().NewIntegration(opts.Logger)
}

// DefaultArguments holds the default settings for the oracledb exporter
var DefaultArguments = Arguments{
	MaxIdleConns: 0,
	MaxOpenConns: 10,
	QueryTimeout: 5,
}

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
	return f((*args)(a))
}

func (a *Arguments) Convert() *oracledb_exporter.Config {
	return &oracledb_exporter.Config{
		ConnectionString: config_util.Secret(a.ConnectionString),
		MaxIdleConns:     a.MaxIdleConns,
		MaxOpenConns:     a.MaxOpenConns,
		QueryTimeout:     a.QueryTimeout,
	}
}
