package windows

import (
	"strings"

	windows_integration "github.com/grafana/agent/pkg/integrations/windows_exporter"
	col "github.com/prometheus-community/windows_exporter/pkg/collector"
)

// DefaultArguments holds non-zero default options for Arguments when it is
// unmarshaled from YAML.
//
// Some defaults are populated from init functions in the github.com/grafana/agent/pkg/integrations/windows_exporter package.

var DefaultArguments = Arguments{
	EnabledCollectors: strings.Split(windows_integration.DefaultConfig.EnabledCollectors, ","),
	Dfsr: DfsrConfig{
		SourcesEnabled: strings.Split(col.ConfigDefaults.Dfsr.DfsrEnabledCollectors, ","),
	},
	Exchange: ExchangeConfig{
		EnabledList: strings.Split(col.ConfigDefaults.Exchange.CollectorsEnabled, ","),
	},
	IIS: IISConfig{
		AppBlackList:  col.ConfigDefaults.Iis.AppExclude,
		AppWhiteList:  col.ConfigDefaults.Iis.AppInclude,
		SiteBlackList: col.ConfigDefaults.Iis.SiteExclude,
		SiteWhiteList: col.ConfigDefaults.Iis.SiteInclude,
		AppInclude:    col.ConfigDefaults.Iis.AppInclude,
		AppExclude:    col.ConfigDefaults.Iis.AppExclude,
		SiteInclude:   col.ConfigDefaults.Iis.SiteInclude,
		SiteExclude:   col.ConfigDefaults.Iis.SiteExclude,
	},
	LogicalDisk: LogicalDiskConfig{
		BlackList: col.ConfigDefaults.LogicalDisk.VolumeExclude,
		WhiteList: col.ConfigDefaults.LogicalDisk.VolumeInclude,
		Include:   col.ConfigDefaults.LogicalDisk.VolumeInclude,
		Exclude:   col.ConfigDefaults.LogicalDisk.VolumeExclude,
	},
	MSMQ: MSMQConfig{
		Where: col.ConfigDefaults.Msmq.QueryWhereClause,
	},
	MSSQL: MSSQLConfig{
		EnabledClasses: strings.Split(col.ConfigDefaults.Mssql.EnabledCollectors, ","),
	},
	Network: NetworkConfig{
		BlackList: col.ConfigDefaults.Net.NicExclude,
		WhiteList: col.ConfigDefaults.Net.NicInclude,
		Include:   col.ConfigDefaults.Net.NicInclude,
		Exclude:   col.ConfigDefaults.Net.NicExclude,
	},
	Process: ProcessConfig{
		BlackList: col.ConfigDefaults.Process.ProcessExclude,
		WhiteList: col.ConfigDefaults.Process.ProcessInclude,
		Include:   col.ConfigDefaults.Process.ProcessInclude,
		Exclude:   col.ConfigDefaults.Process.ProcessExclude,
	},
	ScheduledTask: ScheduledTaskConfig{
		Include: col.ConfigDefaults.ScheduledTask.TaskInclude,
		Exclude: col.ConfigDefaults.ScheduledTask.TaskExclude,
	},
	Service: ServiceConfig{
		UseApi: "false",
		Where:  col.ConfigDefaults.Service.ServiceWhereClause,
	},
	SMTP: SMTPConfig{
		BlackList: col.ConfigDefaults.Smtp.ServerExclude,
		WhiteList: col.ConfigDefaults.Smtp.ServerInclude,
		Include:   col.ConfigDefaults.Smtp.ServerInclude,
		Exclude:   col.ConfigDefaults.Smtp.ServerExclude,
	},
	TextFile: TextFileConfig{
		TextFileDirectory: col.ConfigDefaults.Textfile.TextFileDirectories,
	},
}

// Arguments is used for controlling for this exporter.
type Arguments struct {
	// Collectors to mark as enabled
	EnabledCollectors []string `river:"enabled_collectors,attr,optional"`

	// Collector-specific config options
	Dfsr          DfsrConfig          `river:"dfsr,block,optional"`
	Exchange      ExchangeConfig      `river:"exchange,block,optional"`
	IIS           IISConfig           `river:"iis,block,optional"`
	LogicalDisk   LogicalDiskConfig   `river:"logical_disk,block,optional"`
	MSMQ          MSMQConfig          `river:"msmq,block,optional"`
	MSSQL         MSSQLConfig         `river:"mssql,block,optional"`
	Network       NetworkConfig       `river:"network,block,optional"`
	Process       ProcessConfig       `river:"process,block,optional"`
	ScheduledTask ScheduledTaskConfig `river:"scheduled_task,block,optional"`
	Service       ServiceConfig       `river:"service,block,optional"`
	SMTP          SMTPConfig          `river:"smtp,block,optional"`
	TextFile      TextFileConfig      `river:"text_file,block,optional"`
}

// SetToDefault implements river.Defaulter.
func (a *Arguments) SetToDefault() {
	*a = DefaultArguments
}

// Convert converts the component's Arguments to the integration's Config.
func (a *Arguments) Convert() *windows_integration.Config {
	return &windows_integration.Config{
		EnabledCollectors: strings.Join(a.EnabledCollectors, ","),
		Dfsr:              a.Dfsr.Convert(),
		Exchange:          a.Exchange.Convert(),
		IIS:               a.IIS.Convert(),
		LogicalDisk:       a.LogicalDisk.Convert(),
		MSMQ:              a.MSMQ.Convert(),
		MSSQL:             a.MSSQL.Convert(),
		Network:           a.Network.Convert(),
		Process:           a.Process.Convert(),
		ScheduledTask:     a.ScheduledTask.Convert(),
		Service:           a.Service.Convert(),
		SMTP:              a.SMTP.Convert(),
		TextFile:          a.TextFile.Convert(),
	}
}

// DfsrConfig handles settings for the windows_exporter Exchange collector
type DfsrConfig struct {
	SourcesEnabled []string `river:"sources_enabled,attr,optional"`
}

// Convert converts the component's DfsrConfig to the integration's ExchangeConfig.
func (t DfsrConfig) Convert() windows_integration.DfsrConfig {
	return windows_integration.DfsrConfig{
		SourcesEnabled: strings.Join(t.SourcesEnabled, ","),
	}
}

// ExchangeConfig handles settings for the windows_exporter Exchange collector
type ExchangeConfig struct {
	EnabledList []string `river:"enabled_list,attr,optional"`
}

// Convert converts the component's ExchangeConfig to the integration's ExchangeConfig.
func (t ExchangeConfig) Convert() windows_integration.ExchangeConfig {
	return windows_integration.ExchangeConfig{
		EnabledList: strings.Join(t.EnabledList, ","),
	}
}

// IISConfig handles settings for the windows_exporter IIS collector
type IISConfig struct {
	AppBlackList  string `river:"app_blacklist,attr,optional"`
	AppWhiteList  string `river:"app_whitelist,attr,optional"`
	SiteBlackList string `river:"site_blacklist,attr,optional"`
	SiteWhiteList string `river:"site_whitelist,attr,optional"`
	AppExclude    string `river:"app_exclude,attr,optional"`
	AppInclude    string `river:"app_include,attr,optional"`
	SiteExclude   string `river:"site_exclude,attr,optional"`
	SiteInclude   string `river:"site_include,attr,optional"`
}

// Convert converts the component's IISConfig to the integration's IISConfig.
func (t IISConfig) Convert() windows_integration.IISConfig {
	return windows_integration.IISConfig{
		AppBlackList:  t.AppBlackList,
		AppWhiteList:  t.AppWhiteList,
		SiteBlackList: t.SiteBlackList,
		SiteWhiteList: t.SiteWhiteList,
		AppExclude:    t.AppExclude,
		AppInclude:    t.AppInclude,
		SiteExclude:   t.SiteExclude,
		SiteInclude:   t.SiteInclude,
	}
}

// TextFileConfig handles settings for the windows_exporter Text File collector
type TextFileConfig struct {
	TextFileDirectory string `river:"text_file_directory,attr,optional"`
}

// Convert converts the component's TextFileConfig to the integration's TextFileConfig.
func (t TextFileConfig) Convert() windows_integration.TextFileConfig {
	return windows_integration.TextFileConfig{
		TextFileDirectory: t.TextFileDirectory,
	}
}

// SMTPConfig handles settings for the windows_exporter SMTP collector
type SMTPConfig struct {
	BlackList string `river:"blacklist,attr,optional"`
	WhiteList string `river:"whitelist,attr,optional"`
	Exclude   string `river:"exclude,attr,optional"`
	Include   string `river:"include,attr,optional"`
}

// Convert converts the component's SMTPConfig to the integration's SMTPConfig.
func (t SMTPConfig) Convert() windows_integration.SMTPConfig {
	return windows_integration.SMTPConfig{
		BlackList: t.BlackList,
		WhiteList: t.WhiteList,
		Exclude:   t.Exclude,
		Include:   t.Include,
	}
}

// ServiceConfig handles settings for the windows_exporter service collector
type ServiceConfig struct {
	UseApi string `river:"use_api,attr,optional"`
	Where  string `river:"where_clause,attr,optional"`
}

// Convert converts the component's ServiceConfig to the integration's ServiceConfig.
func (t ServiceConfig) Convert() windows_integration.ServiceConfig {
	return windows_integration.ServiceConfig{
		UseApi: t.UseApi,
		Where:  t.Where,
	}
}

// ProcessConfig handles settings for the windows_exporter process collector
type ProcessConfig struct {
	BlackList string `river:"blacklist,attr,optional"`
	WhiteList string `river:"whitelist,attr,optional"`
	Exclude   string `river:"exclude,attr,optional"`
	Include   string `river:"include,attr,optional"`
}

// Convert converts the component's ProcessConfig to the integration's ProcessConfig.
func (t ProcessConfig) Convert() windows_integration.ProcessConfig {
	return windows_integration.ProcessConfig{
		BlackList: t.BlackList,
		WhiteList: t.WhiteList,
		Exclude:   t.Exclude,
		Include:   t.Include,
	}
}

// ScheduledTaskConfig handles settings for the windows_exporter process collector
type ScheduledTaskConfig struct {
	Exclude string `river:"exclude,attr,optional"`
	Include string `river:"include,attr,optional"`
}

// Convert converts the component's ScheduledTaskConfig to the integration's ScheduledTaskConfig.
func (t ScheduledTaskConfig) Convert() windows_integration.ScheduledTaskConfig {
	return windows_integration.ScheduledTaskConfig{
		Exclude: t.Exclude,
		Include: t.Include,
	}
}

// NetworkConfig handles settings for the windows_exporter network collector
type NetworkConfig struct {
	BlackList string `river:"blacklist,attr,optional"`
	WhiteList string `river:"whitelist,attr,optional"`
	Exclude   string `river:"exclude,attr,optional"`
	Include   string `river:"include,attr,optional"`
}

// Convert converts the component's NetworkConfig to the integration's NetworkConfig.
func (t NetworkConfig) Convert() windows_integration.NetworkConfig {
	return windows_integration.NetworkConfig{
		BlackList: t.BlackList,
		WhiteList: t.WhiteList,
		Exclude:   t.Exclude,
		Include:   t.Include,
	}
}

// MSSQLConfig handles settings for the windows_exporter SQL server collector
type MSSQLConfig struct {
	EnabledClasses []string `river:"enabled_classes,attr,optional"`
}

// Convert converts the component's MSSQLConfig to the integration's MSSQLConfig.
func (t MSSQLConfig) Convert() windows_integration.MSSQLConfig {
	return windows_integration.MSSQLConfig{
		EnabledClasses: strings.Join(t.EnabledClasses, ","),
	}
}

// MSMQConfig handles settings for the windows_exporter MSMQ collector
type MSMQConfig struct {
	Where string `river:"where_clause,attr,optional"`
}

// Convert converts the component's MSMQConfig to the integration's MSMQConfig.
func (t MSMQConfig) Convert() windows_integration.MSMQConfig {
	return windows_integration.MSMQConfig{
		Where: t.Where,
	}
}

// LogicalDiskConfig handles settings for the windows_exporter logical disk collector
type LogicalDiskConfig struct {
	BlackList string `river:"blacklist,attr,optional"`
	WhiteList string `river:"whitelist,attr,optional"`
	Include   string `river:"include,attr,optional"`
	Exclude   string `river:"exclude,attr,optional"`
}

// Convert converts the component's LogicalDiskConfig to the integration's LogicalDiskConfig.
func (t LogicalDiskConfig) Convert() windows_integration.LogicalDiskConfig {
	return windows_integration.LogicalDiskConfig{
		BlackList: t.BlackList,
		WhiteList: t.WhiteList,
		Include:   t.Include,
		Exclude:   t.Exclude,
	}
}
