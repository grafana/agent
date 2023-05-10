package snowflake

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/snowflake_exporter"
	"github.com/grafana/agent/pkg/river/rivertypes"
	config_util "github.com/prometheus/common/config"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.snowflake",
		Args:    Arguments{},
		Exports: exporter.Exports{},
		Build:   exporter.New(createExporter, "snowflake"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return a.Convert().NewIntegration(opts.Logger)
}

// DefaultArguments holds the default settings for the snowflake exporter
var DefaultArguments = Arguments{
	Role: "ACCOUNTADMIN",
}

// Arguments controls the snowflake exporter.
type Arguments struct {
	AccountName string            `river:"account_name,attr"`
	Username    string            `river:"username,attr"`
	Password    rivertypes.Secret `river:"password,attr"`
	Role        string            `river:"role,attr,optional"`
	Warehouse   string            `river:"warehouse,attr"`
}

// UnmarshalRiver implements River unmarshalling for Arguments.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = DefaultArguments

	type args Arguments
	return f((*args)(a))
}

func (a *Arguments) Convert() *snowflake_exporter.Config {
	return &snowflake_exporter.Config{
		AccountName: a.AccountName,
		Username:    a.Username,
		Password:    config_util.Secret(a.Password),
		Role:        a.Role,
		Warehouse:   a.Warehouse,
	}
}
