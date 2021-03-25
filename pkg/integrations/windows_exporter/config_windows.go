package windows_exporter //nolint:golint

import (
	"reflect"

	"github.com/prometheus-community/windows_exporter/collector"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
)

func init() {
	integrations.RegisterIntegration(&Config{})
}

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

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Config
	return unmarshal((*plain)(c))
}

func (c *Config) Name() string {
	return "windows_exporter"
}

func (c *Config) CommonConfig() config.Common {
	return c.Common
}

func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	return New(l, c)
}

func (c *Config) ApplyConfig(exporterConfigs map[string]collector.Config) {
	agentConfigs := []translatableConfig{
		&c.Exchange,
		&c.IIS,
		&c.LogicalDisk,
		&c.MSMQ,
		&c.MSSQL,
		&c.Network,
		&c.Process,
		&c.Service,
		&c.SMTP,
		&c.TextFile,
	}
	// Brute force the syncing, its a bounded set and reduces the code footprint
	for _, ac := range agentConfigs {
		if ac == nil || reflect.ValueOf(ac).IsNil() {
			continue
		}
		for _, ec := range exporterConfigs {
			// Sync will return true if it can handle the exporter config
			// which means we can break early
			if ac.Sync(ec) {
				break
			}
		}
	}
}

type ExchangeConfig struct {
	EnabledList string `yaml:"enabled_list"`
}

func (c *ExchangeConfig) Sync(v interface{}) bool {
	other, ok := v.(*collector.ExchangeConfig)
	if ok {
		setStringIfNotEmpty(c.EnabledList, &other.Enabled)
	}
	return ok
}

type IISConfig struct {
	SiteWhiteList string `yaml:"site_whitelist"`
	SiteBlackList string `yaml:"site_blacklist"`
	AppWhiteList  string `yaml:"app_whitelist"`
	AppBlackList  string `yaml:"app_blacklist"`
}

func (c *IISConfig) Sync(v interface{}) bool {
	other, ok := v.(*collector.IISConfig)
	if ok {
		setStringIfNotEmpty(c.SiteWhiteList, &other.SiteWhiteList)
		setStringIfNotEmpty(c.SiteBlackList, &other.SiteBlackList)
		setStringIfNotEmpty(c.AppWhiteList, &other.AppWhiteList)
		setStringIfNotEmpty(c.AppBlackList, &other.AppBlackList)
	}
	return ok
}

type TextFileConfig struct {
	TextFileDirectory string `yaml:"text_file_directory"`
}

func (c *TextFileConfig) Sync(v interface{}) bool {
	other, ok := v.(*collector.TextFileConfig)
	if ok {
		setStringIfNotEmpty(c.TextFileDirectory, &other.TextFileDirectory)
	}
	return ok
}

type SMTPConfig struct {
	WhiteList string `yaml:"whitelist"`
	BlackList string `yaml:"blacklist"`
}

func (c *SMTPConfig) Sync(v interface{}) bool {
	other, ok := v.(*collector.SMTPConfig)
	if ok {
		setStringIfNotEmpty(c.WhiteList, &other.ServerWhiteList)
		setStringIfNotEmpty(c.BlackList, &other.ServerBlackList)
	}
	return ok
}

type ServiceConfig struct {
	Where string `yaml:"where_clause"`
}

func (c *ServiceConfig) Sync(v interface{}) bool {
	other, ok := v.(*collector.ServiceConfig)
	if ok {
		setStringIfNotEmpty(c.Where, &other.ServiceWhereClause)
	}
	return ok
}

type ProcessConfig struct {
	WhiteList string `yaml:"whitelist"`
	BlackList string `yaml:"blacklist"`
}

func (c *ProcessConfig) Sync(v interface{}) bool {
	other, ok := v.(*collector.ProcessConfig)
	if ok {
		setStringIfNotEmpty(c.WhiteList, &other.ProcessWhiteList)
		setStringIfNotEmpty(c.BlackList, &other.ProcessBlackList)
	}
	return ok
}

type NetworkConfig struct {
	WhiteList string `yaml:"whitelist"`
	BlackList string `yaml:"blacklist"`
}

func (c *NetworkConfig) Sync(v interface{}) bool {
	other, ok := v.(*collector.NetworkConfig)
	if ok {
		setStringIfNotEmpty(c.WhiteList, &other.NICWhiteList)
		setStringIfNotEmpty(c.BlackList, &other.NICBlackList)
	}
	return ok
}

type MSSQLConfig struct {
	EnabledClasses string `yaml:"enabled_classes"`
}

func (c *MSSQLConfig) Sync(v interface{}) bool {
	other, ok := v.(*collector.MSSQLConfig)
	if ok {
		setStringIfNotEmpty(c.EnabledClasses, &other.MSSQLEnabledCollectors)
	}
	return ok
}

type MSMQConfig struct {
	Where string `yaml:"where_clause"`
}

func (c *MSMQConfig) Sync(v interface{}) bool {
	other, ok := v.(*collector.MSMQConfig)
	if ok {
		setStringIfNotEmpty(c.Where, &other.MSMQWhereClause)
	}
	return ok
}

type LogicalDiskConfig struct {
	WhiteList string `yaml:"whitelist"`
	BlackList string `yaml:"blacklist"`
}

func (c *LogicalDiskConfig) Sync(v interface{}) bool {
	other, ok := v.(*collector.LogicalDiskConfig)
	if ok {
		setStringIfNotEmpty(c.WhiteList, &other.VolumeWhiteList)
		setStringIfNotEmpty(c.BlackList, &other.VolumeBlackList)
	}
	return ok
}

type translatableConfig interface {
	Sync(v interface{}) bool
}

// This only works because "" is not a reasonable valid choice for any configurable option currently in windows_exporter
func setStringIfNotEmpty(source string, destination *string) {
	if source == "" {
		return
	}
	*destination = source
}
