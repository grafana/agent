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

var Configs = make([]ConfigurationTemplate, 0)

func init() {
	AddIntegrationConfig(agent.Config{}, nil, shared.TypeSingleton)
	AddIntegrationConfig(cadvisor.Config{}, cadvisor.DefaultConfig, shared.TypeSingleton)
	AddIntegrationConfig(consul_exporter.Config{}, consul_exporter.DefaultConfig, shared.TypeMultiplex)
	AddIntegrationConfig(dnsmasq_exporter.Config{}, dnsmasq_exporter.DefaultConfig, shared.TypeMultiplex)
	AddIntegrationConfig(elasticsearch_exporter.Config{}, elasticsearch_exporter.DefaultConfig, shared.TypeMultiplex)
	AddIntegrationConfig(github_exporter.Config{}, github_exporter.DefaultConfig, shared.TypeMultiplex)
	AddIntegrationConfig(kafka_exporter.Config{}, kafka_exporter.DefaultConfig, shared.TypeMultiplex)
	AddIntegrationConfig(memcached_exporter.Config{}, memcached_exporter.DefaultConfig, shared.TypeMultiplex)
	AddIntegrationConfig(mongodb_exporter.Config{}, nil, shared.TypeMultiplex)
	AddIntegrationConfig(mysqld_exporter.Config{}, mysqld_exporter.DefaultConfig, shared.TypeMultiplex)
	AddIntegrationConfig(node_exporter.Config{}, node_exporter.DefaultConfig, shared.TypeSingleton)
	AddIntegrationConfig(postgres_exporter.Config{}, nil, shared.TypeMultiplex)
	AddIntegrationConfig(process_exporter.Config{}, process_exporter.DefaultConfig, shared.TypeSingleton)
	AddIntegrationConfig(redis_exporter.Config{}, redis_exporter.DefaultConfig, shared.TypeMultiplex)
	AddIntegrationConfig(statsd_exporter.Config{}, statsd_exporter.DefaultConfig, shared.TypeSingleton)
	AddIntegrationConfig(windows_exporter.Config{}, windows_exporter.DefaultConfig, shared.TypeSingleton)
}

func AddIntegrationConfig(config interface{}, defaultConfig interface{}, t shared.Type) {
	Configs = append(Configs, ConfigurationTemplate{
		Config:        config,
		DefaultConfig: defaultConfig,
		Type:          t,
	})

}

type ConfigurationTemplate struct {
	Config        interface{}
	DefaultConfig interface{}
	Type          shared.Type
}
