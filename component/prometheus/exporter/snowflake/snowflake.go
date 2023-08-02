package snowflake

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/snowflake_exporter"
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/grafana/agent/service/http"
	config_util "github.com/prometheus/common/config"
)

func init() {
	component.Register(component.Registration{
		Name:          "prometheus.exporter.snowflake",
		Args:          Arguments{},
		Exports:       exporter.Exports{},
		NeedsServices: []string{http.ServiceName},
		Build:         exporter.NewWithTargetBuilder(createExporter, "snowflake", customizeTarget),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return a.Convert().NewIntegration(opts.Logger)
}

func customizeTarget(baseTarget discovery.Target, args component.Arguments) []discovery.Target {
	a := args.(Arguments)
	target := baseTarget

	target["instance"] = a.AccountName
	return []discovery.Target{target}
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

// SetToDefault implements river.Defaulter.
func (a *Arguments) SetToDefault() {
	*a = DefaultArguments
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
