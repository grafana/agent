package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/windows"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/windows_exporter"
)

func (b *IntegrationsV1ConfigBuilder) appendWindowsExporter(config *windows_exporter.Config) discovery.Exports {
	args := toWindowsExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "windows"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.windows.%s.targets", compLabel))
}

func toWindowsExporter(config *windows_exporter.Config) *windows.Arguments {
	return &windows.Arguments{
		EnabledCollectors: splitByCommaNullOnEmpty(config.EnabledCollectors),
		Dfsr: windows.DfsrConfig{
			SourcesEnabled: splitByCommaNullOnEmpty(config.Dfsr.SourcesEnabled),
		},
		Exchange: windows.ExchangeConfig{
			EnabledList: splitByCommaNullOnEmpty(config.Exchange.EnabledList),
		},
		IIS: windows.IISConfig{
			AppBlackList:  config.IIS.AppBlackList,
			AppWhiteList:  config.IIS.AppWhiteList,
			SiteBlackList: config.IIS.SiteBlackList,
			SiteWhiteList: config.IIS.SiteWhiteList,
			AppExclude:    config.IIS.AppExclude,
			AppInclude:    config.IIS.AppInclude,
			SiteExclude:   config.IIS.SiteExclude,
			SiteInclude:   config.IIS.SiteInclude,
		},
		LogicalDisk: windows.LogicalDiskConfig{
			BlackList: config.LogicalDisk.BlackList,
			WhiteList: config.LogicalDisk.WhiteList,
			Include:   config.LogicalDisk.Include,
			Exclude:   config.LogicalDisk.Exclude,
		},
		MSMQ: windows.MSMQConfig{
			Where: config.MSMQ.Where,
		},
		MSSQL: windows.MSSQLConfig{
			EnabledClasses: splitByCommaNullOnEmpty(config.MSSQL.EnabledClasses),
		},
		Network: windows.NetworkConfig{
			BlackList: config.Network.BlackList,
			WhiteList: config.Network.WhiteList,
			Exclude:   config.Network.Include,
			Include:   config.Network.Exclude,
		},
		Process: windows.ProcessConfig{
			BlackList: config.Process.BlackList,
			WhiteList: config.Process.WhiteList,
			Exclude:   config.Process.Exclude,
			Include:   config.Process.Include,
		},
		ScheduledTask: windows.ScheduledTaskConfig{
			Exclude: config.ScheduledTask.Exclude,
			Include: config.ScheduledTask.Include,
		},
		Service: windows.ServiceConfig{
			UseApi: config.Service.UseApi,
			Where:  config.Service.Where,
		},
		SMTP: windows.SMTPConfig{
			BlackList: config.SMTP.BlackList,
			WhiteList: config.SMTP.WhiteList,
			Exclude:   config.SMTP.Exclude,
			Include:   config.SMTP.Include,
		},
		TextFile: windows.TextFileConfig{
			TextFileDirectory: config.TextFile.TextFileDirectory,
		},
	}
}
