package build

import (
	"github.com/grafana/agent/component/prometheus/exporter/unix"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/integrations/node_exporter"
)

func (b *IntegrationsV1ConfigBuilder) AppendNodeExporter(config *node_exporter.Config) {
	args := ToNodeExporter(config)
	// compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, "default")
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "unix"},
		"",
		args,
	))
	// s.allTargetsExps = append(s.allTargetsExps, "discovery.azure."+compLabel+".targets")

}

func ToNodeExporter(config *node_exporter.Config) *unix.Arguments {
	return &unix.Arguments{}
}
