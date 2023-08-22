package build

import (
	"fmt"
	"strings"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/redis"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/redis_exporter"
	"github.com/grafana/agent/pkg/river/rivertypes"
)

func (b *IntegrationsV1ConfigBuilder) appendRedisExporter(config *redis_exporter.Config) discovery.Exports {
	args := toRedisExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "redis"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoverExports(fmt.Sprintf("prometheus.exporter.redis.%s.targets", compLabel))
}

func toRedisExporter(config *redis_exporter.Config) *redis.Arguments {
	return &redis.Arguments{
		IncludeExporterMetrics:  config.IncludeExporterMetrics,
		RedisAddr:               config.RedisAddr,
		RedisUser:               config.RedisUser,
		RedisPassword:           rivertypes.Secret(config.RedisPassword),
		RedisPasswordFile:       config.RedisPasswordFile,
		RedisPasswordMapFile:    config.RedisPasswordMapFile,
		Namespace:               config.Namespace,
		ConfigCommand:           config.ConfigCommand,
		CheckKeys:               splitNullOnEmpty(config.CheckKeys, ","),
		CheckKeyGroups:          splitNullOnEmpty(config.CheckKeyGroups, ","),
		CheckKeyGroupsBatchSize: config.CheckKeyGroupsBatchSize,
		MaxDistinctKeyGroups:    config.MaxDistinctKeyGroups,
		CheckSingleKeys:         splitNullOnEmpty(config.CheckSingleKeys, ","),
		CheckStreams:            splitNullOnEmpty(config.CheckStreams, ","),
		CheckSingleStreams:      splitNullOnEmpty(config.CheckSingleStreams, ","),
		CountKeys:               splitNullOnEmpty(config.CountKeys, ","),
		ScriptPath:              config.ScriptPath,
		ScriptPaths:             nil,
		ConnectionTimeout:       config.ConnectionTimeout,
		TLSClientKeyFile:        config.TLSClientKeyFile,
		TLSClientCertFile:       config.TLSClientCertFile,
		TLSCaCertFile:           config.TLSCaCertFile,
		SetClientName:           config.SetClientName,
		IsTile38:                config.IsTile38,
		IsCluster:               config.IsCluster,
		ExportClientList:        config.ExportClientList,
		ExportClientPort:        config.ExportClientPort,
		RedisMetricsOnly:        config.RedisMetricsOnly,
		PingOnConnect:           config.PingOnConnect,
		InclSystemMetrics:       config.InclSystemMetrics,
		SkipTLSVerification:     config.SkipTLSVerification,
	}
}

func splitNullOnEmpty(s string, sep string) []string {
	if s == "" {
		return nil
	}

	return strings.Split(s, sep)
}
