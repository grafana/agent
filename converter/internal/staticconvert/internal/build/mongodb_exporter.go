package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/mongodb"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/mongodb_exporter"
	"github.com/grafana/river/rivertypes"
)

func (b *IntegrationsV1ConfigBuilder) appendMongodbExporter(config *mongodb_exporter.Config) discovery.Exports {
	args := toMongodbExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "mongodb"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.mongodb.%s.targets", compLabel))
}

func toMongodbExporter(config *mongodb_exporter.Config) *mongodb.Arguments {
	return &mongodb.Arguments{
		URI:                    rivertypes.Secret(config.URI),
		DirectConnect:          config.DirectConnect,
		DiscoveringMode:        config.DiscoveringMode,
		TLSBasicAuthConfigPath: config.TLSBasicAuthConfigPath,
	}
}
