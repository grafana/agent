package windows_exporter //nolint:golint
import (
	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
)

// DefaultConfig holds the default settings for the windows_exporter integration.
var DefaultConfig = Config{
	EnabledCollectors: "cpu,cs,logical_disk,net,os,service,system",

	// NOTE(rfratto): there is an init function in config_windows.go that
	// populates defaults for collectors based on the exporter defaults.
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// Config controls the windows_exporter integration.
// All of these and their child fields are pointers so we can determine if the value was set or not.
type Config struct {
	Common config.Common `yaml:",inline"`

	EnabledCollectors string `yaml:"enabled_collectors"`

	Exchange    ExchangeConfig    `yaml:"exchange,omitempty"`
	IIS         IISConfig         `yaml:"iis,omitempty"`
	TextFile    TextFileConfig    `yaml:"text_file,omitempty"`
	SMTP        SMTPConfig        `yaml:"smtp,omitempty"`
	Service     ServiceConfig     `yaml:"service,omitempty"`
	Process     ProcessConfig     `yaml:"process,omitempty"`
	Network     NetworkConfig     `yaml:"network,omitempty"`
	MSSQL       MSSQLConfig       `yaml:"mssql,omitempty"`
	MSMQ        MSMQConfig        `yaml:"msmq,omitempty"`
	LogicalDisk LogicalDiskConfig `yaml:"logical_disk,omitempty"`
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

// CommonConfig returns the common fields that all integrations have
func (c *Config) CommonConfig() config.Common {
	return c.Common
}

// InstanceKey returns the hostname:port of the agent.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return agentKey, nil
}

// NewIntegration creates an integration based on the given configuration
func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	return New(l, c)
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
}

// TextFileConfig handles settings for the windows_exporter Text File collector
type TextFileConfig struct {
	TextFileDirectory string `yaml:"text_file_directory,omitempty"`
}

// SMTPConfig handles settings for the windows_exporter SMTP collector
type SMTPConfig struct {
	WhiteList string `yaml:"whitelist,omitempty"`
	BlackList string `yaml:"blacklist,omitempty"`
}

// ServiceConfig handles settings for the windows_exporter service collector
type ServiceConfig struct {
	Where string `yaml:"where_clause,omitempty"`
}

// ProcessConfig handles settings for the windows_exporter process collector
type ProcessConfig struct {
	WhiteList string `yaml:"whitelist,omitempty"`
	BlackList string `yaml:"blacklist,omitempty"`
}

// NetworkConfig handles settings for the windows_exporter network collector
type NetworkConfig struct {
	WhiteList string `yaml:"whitelist,omitempty"`
	BlackList string `yaml:"blacklist,omitempty"`
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
	WhiteList string `yaml:"whitelist,omitempty"`
	BlackList string `yaml:"blacklist,omitempty"`
}
