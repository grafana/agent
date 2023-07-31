package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/apache"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/apache_http"
	"github.com/grafana/agent/pkg/integrations/config"
)

func (b *IntegrationsV1ConfigBuilder) AppendApacheExporter(config *apache_http.Config, commonConfig *config.Common) discovery.Exports {
	args := ToApacheExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "apache"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoverExports(fmt.Sprintf("prometheus.exporter.apache.%s.targets", compLabel))
}

func ToApacheExporter(config *apache_http.Config) *apache.Arguments {
	return &apache.Arguments{
		ApacheAddr:         config.ApacheAddr,
		ApacheHostOverride: config.ApacheHostOverride,
		ApacheInsecure:     config.ApacheInsecure,
	}
}
