package windows_exporter //nolint:golint

import (
	"reflect"

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

	Exchange    *ExchangeConfig    `yaml:"exchange"`
	IIS         *IISConfig         `yaml:"iis"`
	TextFile    *TextFileConfig    `yaml:"text_file"`
	SMTP        *SMTPConfig        `yaml:"smtp"`
	Service     *ServiceConfig     `yaml:"service"`
	Process     *ProcessConfig     `yaml:"process"`
	Network     *NetworkConfig     `yaml:"network"`
	MSSQL       *MSSQLConfig       `yaml:"mssql"`
	MSMQ        *MSMQConfig        `yaml:"msmq"`
	LogicalDisk *LogicalDiskConfig `yaml:"logical_disk"`
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

// The Windows Collector takes a map of configuration to set, so we need to convert from agent config to a key value
// using the windows_exporter key name 'collector.iis.site-whitelist' for example.
func (c *Config) ConvertToMap() map[string]string {
	configMap := make(map[string]string)
	translateConfig(c.Exchange, configMap)
	translateConfig(c.IIS, configMap)
	translateConfig(c.LogicalDisk, configMap)
	translateConfig(c.MSMQ, configMap)
	translateConfig(c.MSSQL, configMap)
	translateConfig(c.Network, configMap)
	translateConfig(c.Process, configMap)
	translateConfig(c.Service, configMap)
	translateConfig(c.SMTP, configMap)
	translateConfig(c.TextFile, configMap)
	return configMap
}

type translatableConfig interface {
	translate(cm map[string]string)
}

func translateConfig(c translatableConfig, cm map[string]string) {
	if c == nil || reflect.ValueOf(c).IsNil() {
		return
	}
	c.translate(cm)
}

type ExchangeConfig struct {
	EnabledList *string `yaml:"enabled_list"`
}

func (c *ExchangeConfig) translate(cm map[string]string) {
	setIfNotNil(cm, "collectors.exchange.enabled", c.EnabledList)
}

type IISConfig struct {
	SiteWhiteList *string `yaml:"site_whitelist"`
	SiteBlackList *string `yaml:"site_blacklist"`
	AppWhiteList  *string `yaml:"app_whitelist"`
	AppBlackList  *string `yaml:"app_blacklist"`
}

func (c *IISConfig) translate(cm map[string]string) {
	setIfNotNil(cm, "collector.iis.site-whitelist", c.SiteWhiteList)
	setIfNotNil(cm, "collector.iis.site-blacklist", c.SiteBlackList)
	setIfNotNil(cm, "collector.iis.app-whitelist", c.AppWhiteList)
	setIfNotNil(cm, "collector.iis.app-blacklist", c.AppBlackList)
}

type TextFileConfig struct {
	TextFileDirectory *string `yaml:"text_file_directory"`
}

func (c *TextFileConfig) translate(cm map[string]string) {
	setIfNotNil(cm, "collector.textfile.directory", c.TextFileDirectory)
}

type SMTPConfig struct {
	WhiteList *string `yaml:"whitelist"`
	BlackList *string `yaml:"blacklist"`
}

func (c *SMTPConfig) translate(cm map[string]string) {
	setIfNotNil(cm, "collector.smtp.server-whitelist", c.WhiteList)
	setIfNotNil(cm, "collector.smtp.server-blacklist", c.BlackList)
}

type ServiceConfig struct {
	Where *string `yaml:"where_clause"`
}

func (c *ServiceConfig) translate(cm map[string]string) {
	setIfNotNil(cm, "collector.service.services-where", c.Where)
}

type ProcessConfig struct {
	WhiteList *string `yaml:"whitelist"`
	BlackList *string `yaml:"blacklist"`
}

func (c *ProcessConfig) translate(cm map[string]string) {
	setIfNotNil(cm, "collector.process.whitelist", c.WhiteList)
	setIfNotNil(cm, "collector.process.blacklist", c.BlackList)
}

type NetworkConfig struct {
	WhiteList *string `yaml:"whitelist"`
	BlackList *string `yaml:"blacklist"`
}

func (c *NetworkConfig) translate(cm map[string]string) {
	setIfNotNil(cm, "collector.net.nic-whitelist", c.WhiteList)
	setIfNotNil(cm, "collector.net.nic-blacklist", c.BlackList)
}

type MSSQLConfig struct {
	EnabledClasses *string `yaml:"enabled_classes"`
}

func (c *MSSQLConfig) translate(cm map[string]string) {
	setIfNotNil(cm, "collectors.mssql.classes-enabled", c.EnabledClasses)
}

type MSMQConfig struct {
	Where *string `yaml:"where_clause"`
}

func (c *MSMQConfig) translate(cm map[string]string) {
	setIfNotNil(cm, "collector.msmq.msmq-where", c.Where)
}

type LogicalDiskConfig struct {
	WhiteList *string `yaml:"whitelist"`
	BlackList *string `yaml:"blacklist"`
}

func (c *LogicalDiskConfig) translate(cm map[string]string) {
	setIfNotNil(cm, "collector.logical_disk.volume-whitelist", c.WhiteList)
	setIfNotNil(cm, "collector.logical_disk.volume-blacklist", c.BlackList)

}

func setIfNotNil(cm map[string]string, key string, value *string) {
	if value == nil {
		return
	}
	cm[key] = *value
}
