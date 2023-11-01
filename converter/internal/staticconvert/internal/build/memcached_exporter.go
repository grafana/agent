package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/memcached"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/integrations/memcached_exporter"
)

func (b *IntegrationsConfigBuilder) appendMemcachedExporter(config *memcached_exporter.Config) discovery.Exports {
	args := toMemcachedExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "memcached"},
		compLabel,
		args,
	))

	return common.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.memcached.%s.targets", compLabel))
}

func toMemcachedExporter(config *memcached_exporter.Config) *memcached.Arguments {
	return &memcached.Arguments{
		Address:   config.MemcachedAddress,
		Timeout:   config.Timeout,
		TLSConfig: common.ToTLSConfig(config.TLSConfig),
	}
}
