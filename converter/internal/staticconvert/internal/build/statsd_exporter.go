package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/statsd"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/statsd_exporter"
)

func (b *IntegrationsV1ConfigBuilder) appendStatsdExporter(config *statsd_exporter.Config) discovery.Exports {
	args := toStatsdExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "statsd"},
		compLabel,
		args,
	))

	if config.MappingConfig != nil {
		b.diags.Add(diag.SeverityLevelError, "mapping_config is not supported in statsd_exporter integrations config")
	}

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.statsd.%s.targets", compLabel))
}

func toStatsdExporter(config *statsd_exporter.Config) *statsd.Arguments {
	return &statsd.Arguments{
		ListenUDP:           config.ListenUDP,
		ListenTCP:           config.ListenTCP,
		ListenUnixgram:      config.ListenUnixgram,
		UnixSocketMode:      config.UnixSocketMode,
		MappingConfig:       "",
		ReadBuffer:          config.ReadBuffer,
		CacheSize:           config.CacheSize,
		CacheType:           config.CacheType,
		EventQueueSize:      config.EventQueueSize,
		EventFlushThreshold: config.EventFlushThreshold,
		EventFlushInterval:  config.EventFlushInterval,
		ParseDogStatsd:      config.ParseDogStatsd,
		ParseInfluxDB:       config.ParseInfluxDB,
		ParseLibrato:        config.ParseLibrato,
		ParseSignalFX:       config.ParseSignalFX,
		RelayAddr:           config.RelayAddr,
		RelayPacketLength:   config.RelayPacketLength,
	}
}
