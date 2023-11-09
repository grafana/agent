package build

import (
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/squid"
	"github.com/grafana/agent/pkg/integrations/squid_exporter"
	"github.com/grafana/river/rivertypes"
)

func (b *IntegrationsConfigBuilder) appendSquidExporter(config *squid_exporter.Config, instanceKey *string) discovery.Exports {
	args := toSquidExporter(config)
	return b.appendExporterBlock(args, config.Name(), instanceKey, "squid")
}

func toSquidExporter(config *squid_exporter.Config) *squid.Arguments {
	return &squid.Arguments{
		SquidAddr:     config.Address,
		SquidUser:     config.Username,
		SquidPassword: rivertypes.Secret(config.Password),
	}
}
