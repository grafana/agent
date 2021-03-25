package windows_exporter //nolint:golint
import (
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/prometheus-community/windows_exporter/collector"
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

func (c *ExchangeConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.ExchangeConfig)
	if ok {
		setStringIfNotEmpty(c.EnabledList, &other.Enabled)
	}
	return ok
}

// IISConfig handles settings for the windows_exporter IIS collector
type IISConfig struct {
	SiteWhiteList string `yaml:"site_whitelist"`
	SiteBlackList string `yaml:"site_blacklist"`
	AppWhiteList  string `yaml:"app_whitelist"`
	AppBlackList  string `yaml:"app_blacklist"`
}

func (c *IISConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.IISConfig)
	if ok {
		setStringIfNotEmpty(c.SiteWhiteList, &other.SiteWhiteList)
		setStringIfNotEmpty(c.SiteBlackList, &other.SiteBlackList)
		setStringIfNotEmpty(c.AppWhiteList, &other.AppWhiteList)
		setStringIfNotEmpty(c.AppBlackList, &other.AppBlackList)
	}
	return ok
}

// TextFileConfig handles settings for the windows_exporter Text File collector
type TextFileConfig struct {
	TextFileDirectory string `yaml:"text_file_directory"`
}

func (c *TextFileConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.TextFileConfig)
	if ok {
		setStringIfNotEmpty(c.TextFileDirectory, &other.TextFileDirectory)
	}
	return ok
}

// SMTPConfig handles settings for the windows_exporter SMTP collector
type SMTPConfig struct {
	WhiteList string `yaml:"whitelist"`
	BlackList string `yaml:"blacklist"`
}

func (c *SMTPConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.SMTPConfig)
	if ok {
		setStringIfNotEmpty(c.WhiteList, &other.ServerWhiteList)
		setStringIfNotEmpty(c.BlackList, &other.ServerBlackList)
	}
	return ok
}

// ServiceConfig handles settings for the windows_exporter service collector
type ServiceConfig struct {
	Where string `yaml:"where_clause"`
}

func (c *ServiceConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.ServiceConfig)
	if ok {
		setStringIfNotEmpty(c.Where, &other.ServiceWhereClause)
	}
	return ok
}

// ProcessConfig handles settings for the windows_exporter process collector
type ProcessConfig struct {
	WhiteList string `yaml:"whitelist"`
	BlackList string `yaml:"blacklist"`
}

func (c *ProcessConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.ProcessConfig)
	if ok {
		setStringIfNotEmpty(c.WhiteList, &other.ProcessWhiteList)
		setStringIfNotEmpty(c.BlackList, &other.ProcessBlackList)
	}
	return ok
}

// NetworkConfig handles settings for the windows_exporter network collector
type NetworkConfig struct {
	WhiteList string `yaml:"whitelist"`
	BlackList string `yaml:"blacklist"`
}

func (c *NetworkConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.NetworkConfig)
	if ok {
		setStringIfNotEmpty(c.WhiteList, &other.NICWhiteList)
		setStringIfNotEmpty(c.BlackList, &other.NICBlackList)
	}
	return ok
}

// MSSQLConfig handles settings for the windows_exporter SQL server collector
type MSSQLConfig struct {
	EnabledClasses string `yaml:"enabled_classes"`
}

func (c *MSSQLConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.MSSQLConfig)
	if ok {
		setStringIfNotEmpty(c.EnabledClasses, &other.MSSQLEnabledCollectors)
	}
	return ok
}

// MSMQConfig handles settings for the windows_exporter MSMQ collector
type MSMQConfig struct {
	Where string `yaml:"where_clause"`
}

func (c *MSMQConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.MSMQConfig)
	if ok {
		setStringIfNotEmpty(c.Where, &other.MSMQWhereClause)
	}
	return ok
}

// LogicalDiskConfig handles settings for the windows_exporter logical disk collector
type LogicalDiskConfig struct {
	WhiteList string `yaml:"whitelist"`
	BlackList string `yaml:"blacklist"`
}

func (c *LogicalDiskConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.LogicalDiskConfig)
	if ok {
		setStringIfNotEmpty(c.WhiteList, &other.VolumeWhiteList)
		setStringIfNotEmpty(c.BlackList, &other.VolumeBlackList)
	}
	return ok
}

type translatableConfig interface {
	sync(v interface{}) bool
}

// This only works because "" is not a reasonable valid choice for any configurable option currently in windows_exporter
func setStringIfNotEmpty(source string, destination *string) {
	if source == "" {
		return
	}
	*destination = source
}
