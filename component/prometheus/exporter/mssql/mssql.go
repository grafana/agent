package mssql

import (
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/mssql"
	"github.com/grafana/agent/pkg/river/rivertypes"
	config_util "github.com/prometheus/common/config"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.mssql",
		Args:    Arguments{},
		Exports: exporter.Exports{},
		Build:   exporter.New(createExporter, "mssql"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return a.Convert().NewIntegration(opts.Logger)
}

// DefaultArguments holds the default settings for the mssql exporter
var DefaultArguments = Arguments{
	MaxIdleConnections: 3,
	MaxOpenConnections: 3,
	Timeout:            10 * time.Second,
}

// Arguments controls the mssql exporter.
type Arguments struct {
	ConnectionString   rivertypes.Secret `river:"connection_string,attr"`
	MaxIdleConnections int               `river:"max_idle_connections,attr,optional"`
	MaxOpenConnections int               `river:"max_open_connections,attr,optional"`
	Timeout            time.Duration     `river:"timeout,attr,optional"`
}

// UnmarshalRiver implements River unmarshalling for Arguments.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = DefaultArguments

	type args Arguments
	return f((*args)(a))
}

func (a *Arguments) Convert() *mssql.Config {
	return &mssql.Config{
		ConnectionString:   config_util.Secret(a.ConnectionString),
		MaxIdleConnections: a.MaxIdleConnections,
		MaxOpenConnections: a.MaxOpenConnections,
		Timeout:            a.Timeout,
	}
}
