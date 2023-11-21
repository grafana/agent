package windows_exporter //nolint:golint

import (
	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
	col "github.com/prometheus-community/windows_exporter/pkg/collector"
)

// DefaultConfig holds the default settings for the windows_exporter integration.
var DefaultConfig = Config{
	EnabledCollectors: "cpu,cs,logical_disk,net,os,service,system",
	Dfsr: DfsrConfig{
		SourcesEnabled: col.ConfigDefaults.Dfsr.DfsrEnabledCollectors,
	},
	Exchange: ExchangeConfig{
		EnabledList: col.ConfigDefaults.Exchange.CollectorsEnabled,
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
		EnabledClasses: col.ConfigDefaults.Mssql.EnabledCollectors,
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

func init() {
	integrations.RegisterIntegration(&Config{})
	integrations_v2.RegisterLegacy(&Config{}, integrations_v2.TypeSingleton, metricsutils.NewNamedShim("windows"))
}

// Config controls the windows_exporter integration.
// All of these and their child fields are pointers, so we can determine if the value was set or not.
type Config struct {
	EnabledCollectors string `yaml:"enabled_collectors"`

	Dfsr          DfsrConfig          `yaml:"dfsr,omitempty"`
	Exchange      ExchangeConfig      `yaml:"exchange,omitempty"`
	IIS           IISConfig           `yaml:"iis,omitempty"`
	TextFile      TextFileConfig      `yaml:"text_file,omitempty"`
	SMTP          SMTPConfig          `yaml:"smtp,omitempty"`
	Service       ServiceConfig       `yaml:"service,omitempty"`
	Process       ProcessConfig       `yaml:"process,omitempty"`
	Network       NetworkConfig       `yaml:"network,omitempty"`
	MSSQL         MSSQLConfig         `yaml:"mssql,omitempty"`
	MSMQ          MSMQConfig          `yaml:"msmq,omitempty"`
	LogicalDisk   LogicalDiskConfig   `yaml:"logical_disk,omitempty"`
	ScheduledTask ScheduledTaskConfig `yaml:"scheduled_task,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name used, "windows_explorer"
func (c *Config) Name() string {
	return "windows_exporter"
}

// InstanceKey returns the hostname:port of the agent.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return agentKey, nil
}

// NewIntegration creates an integration based on the given configuration
func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	return New(l, c)
}

// DfsrConfig handles settings for the windows_exporter dfsr collector
type DfsrConfig struct {
	SourcesEnabled string `yaml:"sources_enabled,omitempty"`
}

// ExchangeConfig handles settings for the windows_exporter Exchange collector
type ExchangeConfig struct {
	EnabledList string `yaml:"enabled_list,omitempty"`
}

// IISConfig handles settings for the windows_exporter IIS collector
type IISConfig struct {
	SiteWhiteList string `yaml:"site_whitelist,omitempty"`
	SiteBlackList string `yaml:"site_blacklist,omitempty"`
	AppWhiteList  string `yaml:"app_whitelist,omitempty"`
	AppBlackList  string `yaml:"app_blacklist,omitempty"`
	SiteInclude   string `yaml:"site_include,omitempty"`
	SiteExclude   string `yaml:"site_exclude,omitempty"`
	AppInclude    string `yaml:"app_include,omitempty"`
	AppExclude    string `yaml:"app_exclude,omitempty"`
}

// TextFileConfig handles settings for the windows_exporter Text File collector
type TextFileConfig struct {
	TextFileDirectory string `yaml:"text_file_directory,omitempty"`
}

// SMTPConfig handles settings for the windows_exporter SMTP collector
type SMTPConfig struct {
	BlackList string `yaml:"blacklist,omitempty"`
	WhiteList string `yaml:"whitelist,omitempty"`
	Include   string `yaml:"include,omitempty"`
	Exclude   string `yaml:"exclude,omitempty"`
}

// ServiceConfig handles settings for the windows_exporter service collector
type ServiceConfig struct {
	UseApi string `yaml:"use_api,omitempty"`
	Where  string `yaml:"where_clause,omitempty"`
}

// ProcessConfig handles settings for the windows_exporter process collector
type ProcessConfig struct {
	BlackList string `yaml:"blacklist,omitempty"`
	WhiteList string `yaml:"whitelist,omitempty"`
	Include   string `yaml:"include,omitempty"`
	Exclude   string `yaml:"exclude,omitempty"`
}

// NetworkConfig handles settings for the windows_exporter network collector
type NetworkConfig struct {
	BlackList string `yaml:"blacklist,omitempty"`
	WhiteList string `yaml:"whitelist,omitempty"`
	Include   string `yaml:"include,omitempty"`
	Exclude   string `yaml:"exclude,omitempty"`
}

// MSSQLConfig handles settings for the windows_exporter SQL server collector
type MSSQLConfig struct {
	EnabledClasses string `yaml:"enabled_classes,omitempty"`
}

// MSMQConfig handles settings for the windows_exporter MSMQ collector
type MSMQConfig struct {
	Where string `yaml:"where_clause,omitempty"`
}

// LogicalDiskConfig handles settings for the windows_exporter logical disk collector
type LogicalDiskConfig struct {
	BlackList string `yaml:"blacklist,omitempty"`
	WhiteList string `yaml:"whitelist,omitempty"`
	Include   string `yaml:"include,omitempty"`
	Exclude   string `yaml:"exclude,omitempty"`
}

// ScheduledTaskConfig handles settings for the windows_exporter scheduled_task collector
type ScheduledTaskConfig struct {
	Include string `yaml:"include,omitempty"`
	Exclude string `yaml:"exclude,omitempty"`
}
