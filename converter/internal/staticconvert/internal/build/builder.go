package build

import (
	"fmt"
	"strings"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/remotewrite"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/config"
	agent_exporter "github.com/grafana/agent/pkg/integrations/agent"
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
	agent_exporter_v2 "github.com/grafana/agent/pkg/integrations/v2/agent"
	common_v2 "github.com/grafana/agent/pkg/integrations/v2/common"
	"github.com/grafana/agent/pkg/integrations/windows_exporter"
	"github.com/grafana/river/scanner"
	"github.com/grafana/river/token/builder"
	"github.com/prometheus/common/model"
	prom_config "github.com/prometheus/prometheus/config"
)

type IntegrationsConfigBuilder struct {
	f         *builder.File
	diags     *diag.Diagnostics
	cfg       *config.Config
	globalCtx *GlobalContext
}

func NewIntegrationsConfigBuilder(f *builder.File, diags *diag.Diagnostics, cfg *config.Config, globalCtx *GlobalContext) *IntegrationsConfigBuilder {
	return &IntegrationsConfigBuilder{
		f:         f,
		diags:     diags,
		cfg:       cfg,
		globalCtx: globalCtx,
	}
}

func (b *IntegrationsConfigBuilder) Build() {
	b.appendLogging(b.cfg.Server)
	b.appendServer(b.cfg.Server)
	b.appendIntegrations()
}

func (b *IntegrationsConfigBuilder) appendIntegrations() {
	switch b.cfg.Integrations.Version {
	case config.IntegrationsVersion1:
		b.appendV1Integrations()
	case config.IntegrationsVersion2:
		b.appendV2Integrations()
	default:
		panic(fmt.Sprintf("unknown integrations version %d", b.cfg.Integrations.Version))
	}
}

func (b *IntegrationsConfigBuilder) appendV1Integrations() {
	for _, integration := range b.cfg.Integrations.ConfigV1.Integrations {
		if !integration.Common.Enabled {
			continue
		}

		scrapeIntegration := b.cfg.Integrations.ConfigV1.ScrapeIntegrations
		if integration.Common.ScrapeIntegration != nil {
			scrapeIntegration = *integration.Common.ScrapeIntegration
		}

		if !scrapeIntegration {
			b.diags.Add(diag.SeverityLevelError, fmt.Sprintf("The converter does not support handling integrations which are not being scraped: %s.", integration.Name()))
			continue
		}

		var exports discovery.Exports
		switch itg := integration.Config.(type) {
		case *agent_exporter.Config:
			exports = b.appendAgentExporter(itg)
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

func (b *IntegrationsConfigBuilder) appendExporter(commonConfig *int_config.Common, name string, extraTargets []discovery.Target) {
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

func (b *IntegrationsConfigBuilder) appendV2Integrations() {
	for _, integration := range b.cfg.Integrations.ConfigV2.Configs {
		var exports discovery.Exports
		var commonConfig common_v2.MetricsConfig

		switch itg := integration.(type) {
		case *agent_exporter_v2.Config:
			exports = b.appendAgentExporter(itg)
			commonConfig = itg.Common
		}

		if len(exports.Targets) > 0 {
			b.appendExporterV2(&commonConfig, integration.Name(), exports.Targets)
		}
	}
}

func (b *IntegrationsConfigBuilder) appendExporterV2(commonConfig *common_v2.MetricsConfig, name string, extraTargets []discovery.Target) {
	scrapeConfig := prom_config.DefaultScrapeConfig
	scrapeConfig.JobName = fmt.Sprintf("integrations/%s", name)
	scrapeConfig.RelabelConfigs = commonConfig.Autoscrape.RelabelConfigs
	scrapeConfig.MetricRelabelConfigs = commonConfig.Autoscrape.MetricRelabelConfigs
	// TODO extra labels - discovery.relabel to add the labels

	commonConfig.ApplyDefaults(b.cfg.Integrations.ConfigV2.Metrics.Autoscrape)
	scrapeConfig.ScrapeInterval = commonConfig.Autoscrape.ScrapeInterval
	scrapeConfig.ScrapeTimeout = commonConfig.Autoscrape.ScrapeTimeout

	scrapeConfigs := []*prom_config.ScrapeConfig{&scrapeConfig}

	var remoteWriteExports *remotewrite.Exports
	for _, metrics := range b.cfg.Metrics.Configs {
		if metrics.Name == commonConfig.Autoscrape.MetricsInstance {
			// This must match the name of the existing remote write config in the metrics config:
			label, err := scanner.SanitizeIdentifier("metrics_" + metrics.Name)
			if err != nil {
				b.diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("failed to sanitize job name: %s", err))
			}

			remoteWriteExports = &remotewrite.Exports{
				Receiver: common.ConvertAppendable{Expr: "prometheus.remote_write." + label + ".receiver"},
			}
			break
		}
	}

	if remoteWriteExports == nil {
		b.diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("integration %s is looking for an undefined metrics config: %s", name, commonConfig.Autoscrape.MetricsInstance))
	}

	promConfig := &prom_config.Config{
		GlobalConfig:  b.cfg.Metrics.Global.Prometheus,
		ScrapeConfigs: scrapeConfigs,
	}

	jobNameToCompLabelsFunc := func(jobName string) string {
		labelSuffix := strings.TrimPrefix(jobName, "integrations/")
		if labelSuffix == "" {
			return b.globalCtx.LabelPrefix
		}

		return fmt.Sprintf("%s_%s", b.globalCtx.LabelPrefix, labelSuffix)
	}

	// Need to pass in the remote write reference from the metrics config here:
	b.diags.AddAll(prometheusconvert.AppendAllNested(b.f, promConfig, jobNameToCompLabelsFunc, extraTargets, remoteWriteExports))
}

func splitByCommaNullOnEmpty(s string) []string {
	if s == "" {
		return nil
	}

	return strings.Split(s, ",")
}
