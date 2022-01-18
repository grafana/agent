package v1

import (
	"github.com/grafana/agent/pkg/integrations/shared"
	"github.com/grafana/agent/pkg/integrations/v1/agent"
	"github.com/grafana/agent/pkg/integrations/v1/cadvisor"
	"github.com/grafana/agent/pkg/integrations/v1/consul_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/dnsmasq_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/elasticsearch_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/github_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/kafka_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/memcached_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/mongodb_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/mysqld_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/node_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/postgres_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/process_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/redis_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/statsd_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/windows_exporter"
)

// Configs is a registry of all integration configurations
var Configs = make([]ConfigurationTemplate, 0)

func init() {
	addIntegrationConfig(agent.Config{}, nil, shared.TypeSingleton)
	addIntegrationConfig(cadvisor.Config{}, cadvisor.DefaultConfig, shared.TypeSingleton)
	addIntegrationConfig(consul_exporter.Config{}, consul_exporter.DefaultConfig, shared.TypeMultiplex)
	addIntegrationConfig(dnsmasq_exporter.Config{}, dnsmasq_exporter.DefaultConfig, shared.TypeMultiplex)
	addIntegrationConfig(elasticsearch_exporter.Config{}, elasticsearch_exporter.DefaultConfig, shared.TypeMultiplex)
	addIntegrationConfig(github_exporter.Config{}, github_exporter.DefaultConfig, shared.TypeMultiplex)
	addIntegrationConfig(kafka_exporter.Config{}, kafka_exporter.DefaultConfig, shared.TypeMultiplex)
	addIntegrationConfig(memcached_exporter.Config{}, memcached_exporter.DefaultConfig, shared.TypeMultiplex)
	addIntegrationConfig(mongodb_exporter.Config{}, nil, shared.TypeMultiplex)
	addIntegrationConfig(mysqld_exporter.Config{}, mysqld_exporter.DefaultConfig, shared.TypeMultiplex)
	addIntegrationConfig(node_exporter.Config{}, node_exporter.DefaultConfig, shared.TypeSingleton)
	addIntegrationConfig(postgres_exporter.Config{}, nil, shared.TypeMultiplex)
	addIntegrationConfig(process_exporter.Config{}, process_exporter.DefaultConfig, shared.TypeSingleton)
	addIntegrationConfig(redis_exporter.Config{}, redis_exporter.DefaultConfig, shared.TypeMultiplex)
	addIntegrationConfig(statsd_exporter.Config{}, statsd_exporter.DefaultConfig, shared.TypeSingleton)
	addIntegrationConfig(windows_exporter.Config{}, windows_exporter.DefaultConfig, shared.TypeSingleton)
}

func addIntegrationConfig(config interface{}, defaultConfig interface{}, t shared.Type) {
	Configs = append(Configs, ConfigurationTemplate{
		Config:        config,
		DefaultConfig: defaultConfig,
		Type:          t,
	})

}

// ConfigurationTemplate is used for the code generator to generate the config
type ConfigurationTemplate struct {
	Config        interface{}
	DefaultConfig interface{}
	Type          shared.Type
}
