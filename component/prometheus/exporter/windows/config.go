package windows

import (
	windows_integration "github.com/grafana/agent/pkg/integrations/windows_exporter"
	"strings"
)

// DefaultArguments holds non-zero default options for Arguments when it is
// unmarshaled from YAML.
//
// Some defaults are populated from init functions in the github.com/grafana/agent/pkg/integrations/node_exporter package.
var DefaultArguments = Arguments{
	EnabledCollectors: []string{"cpu", "cs", "logical_disk", "net", "os", "service", "system"},
	IIS:               IISConfig{AppWhiteList: ".+", SiteWhiteList: ".+"},
	TextFile:          TextFileConfig{TextFileDirectory: "C:\\Program Files\\windows_exporter\\textfile_inputs"},
	SMTP:              SMTPConfig{WhiteList: ".+"},
	Process:           ProcessConfig{WhiteList: ".*"},
	Network:           NetworkConfig{WhiteList: ".*"},
	MSSQL:             MSSQLConfig{EnabledClasses: []string{"accessmethods", "availreplica", "bufman", "databases", "dbreplica", "genstats", "locks", "memmgr", "sqlstats", "sqlerrorstransactions"}},
	LogicalDisk:       LogicalDiskConfig{WhiteList: ".+"},
}

// Arguments is used for controlling for this exporter.
type Arguments struct {
	// Collectors to mark as enabled
	EnabledCollectors []string `river:"enabled_collectors,attr,optional"`

	// Collector-specific config options
	Exchange    ExchangeConfig    `river:"exchange,block,optional"`
	IIS         IISConfig         `river:"iis,block,optional"`
	TextFile    TextFileConfig    `river:"text_file,block,optional"`
	SMTP        SMTPConfig        `river:"smtp,block,optional"`
	Service     ServiceConfig     `river:"service,block,optional"`
	Process     ProcessConfig     `river:"process,block,optional"`
	Network     NetworkConfig     `river:"network,block,optional"`
	MSSQL       MSSQLConfig       `river:"mssql,block,optional"`
	MSMQ        MSMQConfig        `river:"msmq,block,optional"`
	LogicalDisk LogicalDiskConfig `river:"logical_disk,block,optional"`
}

// UnmarshalRiver implements River unmarshalling for Config.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = DefaultArguments

	type args Arguments
	return f((*args)(a))
}

// Convert converts the component's Arguments to the integration's Config.
func (a *Arguments) Convert() *windows_integration.Config {
	return &windows_integration.Config{
		EnabledCollectors: strings.Join(a.EnabledCollectors, ","),
		Exchange:          a.Exchange.Convert(),
		IIS:               a.IIS.Convert(),
		TextFile:          a.TextFile.Convert(),
		SMTP:              a.SMTP.Convert(),
		Service:           a.Service.Convert(),
		Process:           a.Process.Convert(),
		Network:           a.Network.Convert(),
		MSSQL:             a.MSSQL.Convert(),
		MSMQ:              a.MSMQ.Convert(),
		LogicalDisk:       a.LogicalDisk.Convert(),
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
}

// Convert converts the component's IISConfig to the integration's IISConfig.
func (t IISConfig) Convert() windows_integration.IISConfig {
	return windows_integration.IISConfig{
		AppBlackList:  t.AppBlackList,
		AppWhiteList:  t.AppWhiteList,
		SiteBlackList: t.SiteBlackList,
		SiteWhiteList: t.SiteWhiteList,
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
}

// Convert converts the component's SMTPConfig to the integration's SMTPConfig.
func (t SMTPConfig) Convert() windows_integration.SMTPConfig {
	return windows_integration.SMTPConfig{
		BlackList: t.BlackList,
		WhiteList: t.WhiteList,
	}
}

// ServiceConfig handles settings for the windows_exporter service collector
type ServiceConfig struct {
	Where string `river:"where_clause,attr,optional"`
}

// Convert converts the component's ServiceConfig to the integration's ServiceConfig.
func (t ServiceConfig) Convert() windows_integration.ServiceConfig {
	return windows_integration.ServiceConfig{
		Where: t.Where,
	}
}

// ProcessConfig handles settings for the windows_exporter process collector
type ProcessConfig struct {
	BlackList string `river:"blacklist,attr,optional"`
	WhiteList string `river:"whitelist,attr,optional"`
}

// Convert converts the component's ProcessConfig to the integration's ProcessConfig.
func (t ProcessConfig) Convert() windows_integration.ProcessConfig {
	return windows_integration.ProcessConfig{
		BlackList: t.BlackList,
		WhiteList: t.WhiteList,
	}
}

// NetworkConfig handles settings for the windows_exporter network collector
type NetworkConfig struct {
	BlackList string `river:"blacklist,attr,optional"`
	WhiteList string `river:"whitelist,attr,optional"`
}

// Convert converts the component's NetworkConfig to the integration's NetworkConfig.
func (t NetworkConfig) Convert() windows_integration.NetworkConfig {
	return windows_integration.NetworkConfig{
		BlackList: t.BlackList,
		WhiteList: t.WhiteList,
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
	WhiteList string `river:"whitelist,attr,optional"`
	BlackList string `river:"blacklist,attr,optional"`
}

// Convert converts the component's LogicalDiskConfig to the integration's LogicalDiskConfig.
func (t LogicalDiskConfig) Convert() windows_integration.LogicalDiskConfig {
	return windows_integration.LogicalDiskConfig{
		WhiteList: t.WhiteList,
		BlackList: t.BlackList,
	}
}
