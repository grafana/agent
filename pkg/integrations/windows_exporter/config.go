package windows_exporter //nolint:golint
import (
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
)

// Config controls the windows_exporter integration.
// All of these and their child fields are pointers so we can determine if the value was set or not.
type Config struct {
	Common config.Common `yaml:",inline"`

	EnabledCollectors string `yaml:"enabled_collectors"`

	Exchange    ExchangeConfig    `yaml:"exchange"`
	IIS         IISConfig         `yaml:"iis"`
	TextFile    TextFileConfig    `yaml:"text_file"`
	SMTP        SMTPConfig        `yaml:"smtp"`
	Service     ServiceConfig     `yaml:"service"`
	Process     ProcessConfig     `yaml:"process"`
	Network     NetworkConfig     `yaml:"network"`
	MSSQL       MSSQLConfig       `yaml:"mssql"`
	MSMQ        MSMQConfig        `yaml:"msmq"`
	LogicalDisk LogicalDiskConfig `yaml:"logical_disk"`
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// Name returns the name used, "windows_explorer"
func (c *Config) Name() string {
	return "windows_exporter"
}

// CommonConfig returns the common fields that all integrations have
func (c *Config) CommonConfig() config.Common {
	return c.Common
}

// NewIntegration creates an integration based on the given configuration
func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	return New(l, c)
}

// ExchangeConfig handles settings for the windows_exporter Exchange collector
type ExchangeConfig struct {
	EnabledList string `yaml:"enabled_list"`
}

// IISConfig handles settings for the windows_exporter IIS collector
type IISConfig struct {
	SiteWhiteList string `yaml:"site_whitelist"`
	SiteBlackList string `yaml:"site_blacklist"`
	AppWhiteList  string `yaml:"app_whitelist"`
	AppBlackList  string `yaml:"app_blacklist"`
}

// TextFileConfig handles settings for the windows_exporter Text File collector
type TextFileConfig struct {
	TextFileDirectory string `yaml:"text_file_directory"`
}

// SMTPConfig handles settings for the windows_exporter SMTP collector
type SMTPConfig struct {
	WhiteList string `yaml:"whitelist"`
	BlackList string `yaml:"blacklist"`
}

// ServiceConfig handles settings for the windows_exporter service collector
type ServiceConfig struct {
	Where string `yaml:"where_clause"`
}

// ProcessConfig handles settings for the windows_exporter process collector
type ProcessConfig struct {
	WhiteList string `yaml:"whitelist"`
	BlackList string `yaml:"blacklist"`
}

// NetworkConfig handles settings for the windows_exporter network collector
type NetworkConfig struct {
	WhiteList string `yaml:"whitelist"`
	BlackList string `yaml:"blacklist"`
}

// MSSQLConfig handles settings for the windows_exporter SQL server collector
type MSSQLConfig struct {
	EnabledClasses string `yaml:"enabled_classes"`
}

// MSMQConfig handles settings for the windows_exporter MSMQ collector
type MSMQConfig struct {
	Where string `yaml:"where_clause"`
}

// LogicalDiskConfig handles settings for the windows_exporter logical disk collector
type LogicalDiskConfig struct {
	WhiteList string `yaml:"whitelist"`
	BlackList string `yaml:"blacklist"`
}
