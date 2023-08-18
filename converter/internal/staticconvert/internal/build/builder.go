package build

import (
	"fmt"
	"strings"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/agent/pkg/integrations/apache_http"
	"github.com/grafana/agent/pkg/integrations/blackbox_exporter"
	"github.com/grafana/agent/pkg/integrations/cloudwatch_exporter"
	int_config "github.com/grafana/agent/pkg/integrations/config"
	"github.com/grafana/agent/pkg/integrations/node_exporter"
	"github.com/grafana/agent/pkg/river/token/builder"
	"github.com/prometheus/common/model"
	prom_config "github.com/prometheus/prometheus/config"
)

type IntegrationsV1ConfigBuilder struct {
	f         *builder.File
	diags     *diag.Diagnostics
	cfg       *config.Config
	globalCtx *GlobalContext
}

func NewIntegrationsV1ConfigBuilder(f *builder.File, diags *diag.Diagnostics, cfg *config.Config, globalCtx *GlobalContext) *IntegrationsV1ConfigBuilder {
	return &IntegrationsV1ConfigBuilder{
		f:         f,
		diags:     diags,
		cfg:       cfg,
		globalCtx: globalCtx,
	}
}

func (b *IntegrationsV1ConfigBuilder) AppendIntegrations() {
	for _, integration := range b.cfg.Integrations.ConfigV1.Integrations {
		if !integration.Common.Enabled {
			continue
		}

		var exports discovery.Exports
		switch itg := integration.Config.(type) {
		case *apache_http.Config:
			exports = b.appendApacheExporter(itg)
		case *node_exporter.Config:
			exports = b.appendNodeExporter(itg)
		case *blackbox_exporter.Config:
			exports = b.appendBlackboxExporter(itg)
		case *cloudwatch_exporter.Config:
			exports = b.appendCloudwatchExporter(itg)
		}

		if len(exports.Targets) > 0 {
			b.appendExporter(&integration.Common, integration.Name(), exports.Targets)
		}
	}
}

func (b *IntegrationsV1ConfigBuilder) appendExporter(commonConfig *int_config.Common, name string, extraTargets []discovery.Target) {
	scrapeConfigs := []*prom_config.ScrapeConfig{}
	if b.cfg.Integrations.ConfigV1.ScrapeIntegrations {
		scrapeConfig := prom_config.DefaultScrapeConfig
		scrapeConfig.JobName = fmt.Sprintf("integrations/%s", name)
		scrapeConfig.RelabelConfigs = commonConfig.RelabelConfigs
		scrapeConfig.MetricRelabelConfigs = commonConfig.MetricRelabelConfigs
		// TODO: Add support for scrapeConfig.HTTPClientConfig

		scrapeConfig.ScrapeInterval = model.Duration(commonConfig.ScrapeInterval)
		if commonConfig.ScrapeInterval == 0 {
			scrapeConfig.ScrapeInterval = b.cfg.Integrations.ConfigV1.PrometheusGlobalConfig.ScrapeInterval
		}

		scrapeConfig.ScrapeTimeout = model.Duration(commonConfig.ScrapeTimeout)
		if commonConfig.ScrapeTimeout == 0 {
			scrapeConfig.ScrapeTimeout = b.cfg.Integrations.ConfigV1.PrometheusGlobalConfig.ScrapeTimeout
		}

		scrapeConfigs = []*prom_config.ScrapeConfig{&scrapeConfig}
	}

	promConfig := &prom_config.Config{
		GlobalConfig:       b.cfg.Integrations.ConfigV1.PrometheusGlobalConfig,
		ScrapeConfigs:      scrapeConfigs,
		RemoteWriteConfigs: b.cfg.Integrations.ConfigV1.PrometheusRemoteWrite,
	}

	jobNameToCompLabelsFunc := func(jobName string) string {
		labelSuffix := strings.TrimPrefix(jobName, "integrations/")
		if labelSuffix == "" {
			return b.globalCtx.LabelPrefix
		}

		return fmt.Sprintf("%s_%s", b.globalCtx.LabelPrefix, labelSuffix)
	}

	b.diags.AddAll(prometheusconvert.AppendAllNested(b.f, promConfig, jobNameToCompLabelsFunc, extraTargets, b.globalCtx.RemoteWriteExports))
	b.globalCtx.InitializeRemoteWriteExports()
}
