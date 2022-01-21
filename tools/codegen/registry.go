package main

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
	agentv2 "github.com/grafana/agent/pkg/integrations/v2/agent"
)

var v1Configs = []ConfigurationTemplate{
	newIntegrationConfig(agent.Config{}, nil, shared.TypeSingleton),
	newIntegrationConfig(cadvisor.Config{}, cadvisor.DefaultConfig, shared.TypeSingleton),
	newIntegrationConfig(consul_exporter.Config{}, consul_exporter.DefaultConfig, shared.TypeMultiplex),
	newIntegrationConfig(dnsmasq_exporter.Config{}, dnsmasq_exporter.DefaultConfig, shared.TypeMultiplex),
	newIntegrationConfig(elasticsearch_exporter.Config{}, elasticsearch_exporter.DefaultConfig, shared.TypeMultiplex),
	newIntegrationConfig(github_exporter.Config{}, github_exporter.DefaultConfig, shared.TypeMultiplex),
	newIntegrationConfig(kafka_exporter.Config{}, kafka_exporter.DefaultConfig, shared.TypeMultiplex),
	newIntegrationConfig(memcached_exporter.Config{}, memcached_exporter.DefaultConfig, shared.TypeMultiplex),
	newIntegrationConfig(mongodb_exporter.Config{}, nil, shared.TypeMultiplex),
	newIntegrationConfig(mysqld_exporter.Config{}, mysqld_exporter.DefaultConfig, shared.TypeMultiplex),
	newIntegrationConfig(node_exporter.Config{}, node_exporter.DefaultConfig, shared.TypeSingleton),
	newIntegrationConfig(postgres_exporter.Config{}, nil, shared.TypeMultiplex),
	newIntegrationConfig(process_exporter.Config{}, process_exporter.DefaultConfig, shared.TypeSingleton),
	newIntegrationConfig(redis_exporter.Config{}, redis_exporter.DefaultConfig, shared.TypeMultiplex),
	newIntegrationConfig(statsd_exporter.Config{}, statsd_exporter.DefaultConfig, shared.TypeSingleton),
	newIntegrationConfig(windows_exporter.Config{}, windows_exporter.DefaultConfig, shared.TypeSingleton),
}

var v2Configs = []ConfigurationTemplate{
	newV2IntegrationConfig(&agentv2.Config{}, nil, shared.TypeSingleton),
	newIntegrationConfig(cadvisor.Config{}, cadvisor.DefaultConfig, shared.TypeSingleton),
	newIntegrationConfig(consul_exporter.Config{}, consul_exporter.DefaultConfig, shared.TypeMultiplex),
	newIntegrationConfig(dnsmasq_exporter.Config{}, dnsmasq_exporter.DefaultConfig, shared.TypeMultiplex),
	newIntegrationConfig(elasticsearch_exporter.Config{}, elasticsearch_exporter.DefaultConfig, shared.TypeMultiplex),
	newIntegrationConfig(github_exporter.Config{}, github_exporter.DefaultConfig, shared.TypeMultiplex),
	newIntegrationConfig(kafka_exporter.Config{}, kafka_exporter.DefaultConfig, shared.TypeMultiplex),
	newIntegrationConfig(memcached_exporter.Config{}, memcached_exporter.DefaultConfig, shared.TypeMultiplex),
	newIntegrationConfig(mongodb_exporter.Config{}, nil, shared.TypeMultiplex),
	newIntegrationConfig(mysqld_exporter.Config{}, mysqld_exporter.DefaultConfig, shared.TypeMultiplex),
	newIntegrationConfig(node_exporter.Config{}, node_exporter.DefaultConfig, shared.TypeSingleton),
	newIntegrationConfig(postgres_exporter.Config{}, nil, shared.TypeMultiplex),
	newIntegrationConfig(process_exporter.Config{}, process_exporter.DefaultConfig, shared.TypeSingleton),
	newIntegrationConfig(redis_exporter.Config{}, redis_exporter.DefaultConfig, shared.TypeMultiplex),
	newIntegrationConfig(statsd_exporter.Config{}, statsd_exporter.DefaultConfig, shared.TypeSingleton),
	newIntegrationConfig(windows_exporter.Config{}, windows_exporter.DefaultConfig, shared.TypeSingleton),
}

func newIntegrationConfig(config interface{}, defaultConfig interface{}, t shared.Type) ConfigurationTemplate {
	return ConfigurationTemplate{
		Config:        config,
		DefaultConfig: defaultConfig,
		Type:          t,
		IsV1:          true,
	}
}

func newV2IntegrationConfig(config v2Config, defaultConfig interface{}, t shared.Type) ConfigurationTemplate {
	return ConfigurationTemplate{
		Config:        config,
		DefaultConfig: defaultConfig,
		Type:          t,
		IsV1:          false,
	}
}

type v2Config interface {
	Name() string
	ApplyDefaults(shared.Globals) error
	Identifier(shared.Globals) (string, error)
}
