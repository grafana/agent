package windows_exporter //nolint:golint

import (
	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
)

// DefaultConfig holds the default settings for the windows_exporter integration.
var DefaultConfig = Config{
	// NOTE(rfratto): there is an init function in config_windows.go that
	// populates defaults for collectors based on the exporter defaults.
	EnabledCollectors: "cpu,cs,logical_disk,net,os,service,system",

	Dfsr: DfsrConfig{
		SourcesEnabled: "",
	},
	Exchange: ExchangeConfig{
		EnabledList: "",
	},
	IIS: IISConfig{
		AppBlackList:  "",
		AppWhiteList:  "",
		SiteBlackList: "",
		SiteWhiteList: "",
		AppInclude:    "",
		AppExclude:    "",
		SiteInclude:   "",
		SiteExclude:   "",
	},
	LogicalDisk: LogicalDiskConfig{
		BlackList: "",
		WhiteList: "",
		Include:   "",
		Exclude:   "",
	},
	MSMQ: MSMQConfig{
		Where: "",
	},
	MSSQL: MSSQLConfig{
		EnabledClasses: "",
	},
	Network: NetworkConfig{
		BlackList: "",
		WhiteList: "",
		Include:   "",
		Exclude:   "",
	},
	Process: ProcessConfig{
		BlackList: "",
		WhiteList: "",
		Include:   "",
		Exclude:   "",
	},
	ScheduledTask: ScheduledTaskConfig{
		Include: "",
		Exclude: "",
	},
	Service: ServiceConfig{
		UseApi: "",
		Where:  "",
	},
	SMTP: SMTPConfig{
		BlackList: "",
		WhiteList: "",
		Include:   "",
		Exclude:   "",
	},
	TextFile: TextFileConfig{
		TextFileDirectory: "",
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
