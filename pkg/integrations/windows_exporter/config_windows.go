package windows_exporter

import "github.com/prometheus-community/windows_exporter/pkg/collector"

func (c *Config) ToWindowsExporterConfig() collector.Config {
	cfg := collector.ConfigDefaults
	cfg.Dfsr.DfsrEnabledCollectors = c.Dfsr.SourcesEnabled
	cfg.Exchange.CollectorsEnabled = c.Exchange.EnabledList

	cfg.Iis.SiteInclude = coalesceString(c.IIS.SiteInclude, c.IIS.SiteWhiteList)
	cfg.Iis.SiteExclude = coalesceString(c.IIS.SiteExclude, c.IIS.SiteBlackList)
	cfg.Iis.AppInclude = coalesceString(c.IIS.AppInclude, c.IIS.AppWhiteList)
	cfg.Iis.AppExclude = coalesceString(c.IIS.AppExclude, c.IIS.AppBlackList)

	cfg.Service.ServiceWhereClause = c.Service.Where
	cfg.Service.UseAPI = c.Service.UseApi == "true"

	cfg.Smtp.ServerInclude = coalesceString(c.SMTP.Include, c.SMTP.WhiteList)
	cfg.Smtp.ServerExclude = coalesceString(c.SMTP.Exclude, c.SMTP.BlackList)

	cfg.Textfile.TextFileDirectories = c.TextFile.TextFileDirectory

	cfg.Process.ProcessExclude = coalesceString(c.Process.Exclude, c.Process.BlackList)
	cfg.Process.ProcessInclude = coalesceString(c.Process.Include, c.Process.WhiteList)

	cfg.Net.NicExclude = coalesceString(c.Network.Exclude, c.Network.BlackList)
	cfg.Net.NicInclude = coalesceString(c.Network.Include, c.Network.WhiteList)

	cfg.Mssql.EnabledCollectors = c.MSSQL.EnabledClasses

	cfg.Msmq.QueryWhereClause = c.MSMQ.Where

	cfg.LogicalDisk.VolumeInclude = coalesceString(c.LogicalDisk.Include, c.LogicalDisk.WhiteList)
	cfg.LogicalDisk.VolumeExclude = coalesceString(c.LogicalDisk.Exclude, c.LogicalDisk.BlackList)

	cfg.ScheduledTask.TaskInclude = c.ScheduledTask.Include
	cfg.ScheduledTask.TaskExclude = c.ScheduledTask.Exclude

	return cfg
}

func coalesceString(v ...string) string {
	for _, s := range v {
		if s != "" {
			return s
		}
	}
	return ""
}
