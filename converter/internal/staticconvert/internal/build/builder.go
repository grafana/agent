package build

import (
	"fmt"
	"strings"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/agent/pkg/integrations/apache_http"
	"github.com/grafana/agent/pkg/integrations/azure_exporter"
	"github.com/grafana/agent/pkg/integrations/blackbox_exporter"
	"github.com/grafana/agent/pkg/integrations/cadvisor"
	"github.com/grafana/agent/pkg/integrations/cloudwatch_exporter"
	int_config "github.com/grafana/agent/pkg/integrations/config"
	"github.com/grafana/agent/pkg/integrations/consul_exporter"
	"github.com/grafana/agent/pkg/integrations/dnsmasq_exporter"
	"github.com/grafana/agent/pkg/integrations/elasticsearch_exporter"
	"github.com/grafana/agent/pkg/integrations/gcp_exporter"
	"github.com/grafana/agent/pkg/integrations/github_exporter"
	"github.com/grafana/agent/pkg/integrations/kafka_exporter"
	"github.com/grafana/agent/pkg/integrations/memcached_exporter"
	"github.com/grafana/agent/pkg/integrations/mongodb_exporter"
	mssql_exporter "github.com/grafana/agent/pkg/integrations/mssql"
	"github.com/grafana/agent/pkg/integrations/mysqld_exporter"
	"github.com/grafana/agent/pkg/integrations/node_exporter"
	"github.com/grafana/agent/pkg/integrations/oracledb_exporter"
	"github.com/grafana/agent/pkg/integrations/postgres_exporter"
	"github.com/grafana/agent/pkg/integrations/process_exporter"
	"github.com/grafana/agent/pkg/integrations/redis_exporter"
	"github.com/grafana/agent/pkg/integrations/snmp_exporter"
	"github.com/grafana/agent/pkg/integrations/snowflake_exporter"
	"github.com/grafana/agent/pkg/integrations/squid_exporter"
	"github.com/grafana/agent/pkg/integrations/statsd_exporter"
	"github.com/grafana/agent/pkg/integrations/windows_exporter"
	"github.com/grafana/river/token/builder"
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

func (b *IntegrationsV1ConfigBuilder) Build() {
	b.appendLogging(b.cfg.Server)
	b.appendServer(b.cfg.Server)
	b.appendIntegrations()
}

func (b *IntegrationsV1ConfigBuilder) appendIntegrations() {
	for _, integration := range b.cfg.Integrations.ConfigV1.Integrations {
		if !integration.Common.Enabled {
			continue
		}

		scrapeIntegration := b.cfg.Integrations.ConfigV1.ScrapeIntegrations
		if integration.Common.ScrapeIntegration != nil {
			scrapeIntegration = *integration.Common.ScrapeIntegration
		}

		if !scrapeIntegration {
			b.diags.Add(diag.SeverityLevelError, fmt.Sprintf("unsupported integration which is not being scraped was provided: %s.", integration.Name()))
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
		case *consul_exporter.Config:
			exports = b.appendConsulExporter(itg)
		case *dnsmasq_exporter.Config:
			exports = b.appendDnsmasqExporter(itg)
		case *elasticsearch_exporter.Config:
			exports = b.appendElasticsearchExporter(itg)
		case *gcp_exporter.Config:
			exports = b.appendGcpExporter(itg)
		case *github_exporter.Config:
			exports = b.appendGithubExporter(itg)
		case *kafka_exporter.Config:
			exports = b.appendKafkaExporter(itg)
		case *memcached_exporter.Config:
			exports = b.appendMemcachedExporter(itg)
		case *mongodb_exporter.Config:
			exports = b.appendMongodbExporter(itg)
		case *mssql_exporter.Config:
			exports = b.appendMssqlExporter(itg)
		case *mysqld_exporter.Config:
			exports = b.appendMysqldExporter(itg)
		case *oracledb_exporter.Config:
			exports = b.appendOracledbExporter(itg)
		case *postgres_exporter.Config:
			exports = b.appendPostgresExporter(itg)
		case *process_exporter.Config:
			exports = b.appendProcessExporter(itg)
		case *redis_exporter.Config:
			exports = b.appendRedisExporter(itg)
		case *snmp_exporter.Config:
			exports = b.appendSnmpExporter(itg)
		case *snowflake_exporter.Config:
			exports = b.appendSnowflakeExporter(itg)
		case *squid_exporter.Config:
			exports = b.appendSquidExporter(itg)
		case *statsd_exporter.Config:
			exports = b.appendStatsdExporter(itg)
		case *windows_exporter.Config:
			exports = b.appendWindowsExporter(itg)
		case *azure_exporter.Config:
			exports = b.appendAzureExporter(itg)
		case *cadvisor.Config:
			exports = b.appendCadvisorExporter(itg)
		}

		if len(exports.Targets) > 0 {
			b.appendExporter(&integration.Common, integration.Name(), exports.Targets)
		}
	}
}

func (b *IntegrationsV1ConfigBuilder) appendExporter(commonConfig *int_config.Common, name string, extraTargets []discovery.Target) {
	scrapeConfig := prom_config.DefaultScrapeConfig
	scrapeConfig.JobName = fmt.Sprintf("integrations/%s", name)
	scrapeConfig.RelabelConfigs = commonConfig.RelabelConfigs
	scrapeConfig.MetricRelabelConfigs = commonConfig.MetricRelabelConfigs
	scrapeConfig.HTTPClientConfig.TLSConfig = b.cfg.Integrations.ConfigV1.TLSConfig

	scrapeConfig.ScrapeInterval = model.Duration(commonConfig.ScrapeInterval)
	if commonConfig.ScrapeInterval == 0 {
		scrapeConfig.ScrapeInterval = b.cfg.Integrations.ConfigV1.PrometheusGlobalConfig.ScrapeInterval
	}

	scrapeConfig.ScrapeTimeout = model.Duration(commonConfig.ScrapeTimeout)
	if commonConfig.ScrapeTimeout == 0 {
		scrapeConfig.ScrapeTimeout = b.cfg.Integrations.ConfigV1.PrometheusGlobalConfig.ScrapeTimeout
	}

	scrapeConfigs := []*prom_config.ScrapeConfig{&scrapeConfig}

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

func splitByCommaNullOnEmpty(s string) []string {
	if s == "" {
		return nil
	}

	return strings.Split(s, ",")
}
