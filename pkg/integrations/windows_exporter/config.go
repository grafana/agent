package windows_exporter //nolint:golint

import (
	"reflect"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
)

// Config controls the node_exporter integration.
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

func (c *Config) ConvertToMap() map[string]string {
	configMap := make(map[string]string)
	mapToConfig(c, configMap)
	return configMap
}

// This is to create a map from the config using the exporter struct tag as the key. This is to be ONLY used on pointer configs
// and pointer strings. This is because we need to know if a value was set so that we don't overwrite defaults.
func mapToConfig(config interface{}, cm map[string]string) {
	if config == nil || (reflect.ValueOf(config).Kind() == reflect.Ptr && reflect.ValueOf(config).IsNil()) {
		return
	}
	t := reflect.TypeOf(config)
	v := reflect.ValueOf(config)
	if t.Kind() != reflect.Ptr {
		return
	}
	fieldCount := t.Elem().NumField()
	for i := 0; i < fieldCount; i++ {
		iv := v.Elem().Field(i)
		f := t.Elem().Field(i)
		en := f.Tag.Get("exporter")
		if en != "" && iv.Kind() == reflect.Ptr && iv.Elem().Kind() == reflect.String {
			cm[en] = iv.Elem().String()
			continue
		}
		if iv.Kind() == reflect.Ptr {
			mapToConfig(iv.Interface(), cm)
		}
	}
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

type ExchangeConfig struct {
	EnabledList *string `yaml:"enabled_list" exporter:"collectors.exchange.enabled"`
}

type IISConfig struct {
	SiteWhiteList *string `yaml:"site_whitelist" exporter:"collector.iis.site-whitelist"`
	SiteBlackList *string `yaml:"site_blacklist" exporter:"collector.iis.site-blacklist"`
	AppWhiteList  *string `yaml:"app_whitelist" exporter:"collector.iis.app-whitelist"`
	AppBlackList  *string `yaml:"app_blacklist" exporter:"collector.iis.app-blacklist"`
}

type TextFileConfig struct {
	TextFileDirectory *string `yaml:"text_file_directory" exporter:"collector.textfile.directory"`
}

type SMTPConfig struct {
	WhiteList *string `yaml:"whitelist" exporter:"collector.smtp.server-whitelist"`
	BlackList *string `yaml:"blacklist" exporter:"collector.smtp.server-blacklist"`
}

type ServiceConfig struct {
	Where *string `yaml:"where_clause" exporter:"collector.service.services-where"`
}

type ProcessConfig struct {
	WhiteList *string `yaml:"whitelist" exporter:"collector.process.whitelist"`
	BlackList *string `yaml:"blacklist" exporter:"collector.process.blacklist"`
}

type NetworkConfig struct {
	WhiteList *string `yaml:"whitelist" exporter:"collector.net.nic-whitelist"`
	BlackList *string `yaml:"blacklist" exporter:"collector.net.nic-blacklist"`
}

type MSSQLConfig struct {
	EnabledClasses *string `yaml:"enabled_classes" exporter:"collectors.mssql.classes-enabled"`
}

type MSMQConfig struct {
	Where *string `yaml:"where_clause" exporter:"collector.msmq.msmq-where"`
}

type LogicalDiskConfig struct {
	WhiteList *string `yaml:"whitelist" exporter:"collector.logical_disk.volume-whitelist"`
	BlackList *string `yaml:"blacklist" exporter:"collector.logical_disk.volume-blacklist"`
}
