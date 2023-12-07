package build

import (
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/consul"
	"github.com/grafana/agent/pkg/integrations/consul_exporter"
)

func (b *IntegrationsConfigBuilder) appendConsulExporter(config *consul_exporter.Config, instanceKey *string) discovery.Exports {
	args := toConsulExporter(config)
	return b.appendExporterBlock(args, config.Name(), instanceKey, "consul")
}

func toConsulExporter(config *consul_exporter.Config) *consul.Arguments {
	return &consul.Arguments{
		Server:             config.Server,
		CAFile:             config.CAFile,
		CertFile:           config.CertFile,
		KeyFile:            config.KeyFile,
		ServerName:         config.ServerName,
		Timeout:            config.Timeout,
		InsecureSkipVerify: config.InsecureSkipVerify,
		RequestLimit:       config.RequestLimit,
		AllowStale:         config.AllowStale,
		RequireConsistent:  config.RequireConsistent,
		KVPrefix:           config.KVPrefix,
		KVFilter:           config.KVFilter,
		HealthSummary:      config.HealthSummary,
	}
}
