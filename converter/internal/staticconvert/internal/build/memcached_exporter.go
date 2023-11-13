package build

import (
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/memcached"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/integrations/memcached_exporter"
)

func (b *IntegrationsConfigBuilder) appendMemcachedExporter(config *memcached_exporter.Config, instanceKey *string) discovery.Exports {
	args := toMemcachedExporter(config)
	return b.appendExporterBlock(args, config.Name(), instanceKey, "memcached")
}

func toMemcachedExporter(config *memcached_exporter.Config) *memcached.Arguments {
	return &memcached.Arguments{
		Address:   config.MemcachedAddress,
		Timeout:   config.Timeout,
		TLSConfig: common.ToTLSConfig(config.TLSConfig),
	}
}
