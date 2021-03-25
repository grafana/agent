package windows_exporter //nolint:golint

import (
	"reflect"

	"github.com/prometheus-community/windows_exporter/collector"
)

func (c *Config) applyConfig(exporterConfigs map[string]collector.Config) {
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
			if ac.sync(ec) {
				break
			}
		}
	}
}

// The sync functions are specifically not with their types since they contain windows specific code

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

func (c *ExchangeConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.ExchangeConfig)
	if ok {
		setStringIfNotEmpty(c.EnabledList, &other.Enabled)
	}
	return ok
}

func (c *TextFileConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.TextFileConfig)
	if ok {
		setStringIfNotEmpty(c.TextFileDirectory, &other.TextFileDirectory)
	}
	return ok
}

func (c *SMTPConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.SMTPConfig)
	if ok {
		setStringIfNotEmpty(c.WhiteList, &other.ServerWhiteList)
		setStringIfNotEmpty(c.BlackList, &other.ServerBlackList)
	}
	return ok
}

func (c *ServiceConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.ServiceConfig)
	if ok {
		setStringIfNotEmpty(c.Where, &other.ServiceWhereClause)
	}
	return ok
}

func (c *ProcessConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.ProcessConfig)
	if ok {
		setStringIfNotEmpty(c.WhiteList, &other.ProcessWhiteList)
		setStringIfNotEmpty(c.BlackList, &other.ProcessBlackList)
	}
	return ok
}

func (c *NetworkConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.NetworkConfig)
	if ok {
		setStringIfNotEmpty(c.WhiteList, &other.NICWhiteList)
		setStringIfNotEmpty(c.BlackList, &other.NICBlackList)
	}
	return ok
}

func (c *MSSQLConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.MSSQLConfig)
	if ok {
		setStringIfNotEmpty(c.EnabledClasses, &other.MSSQLEnabledCollectors)
	}
	return ok
}

func (c *MSMQConfig) sync(v interface{}) bool {
	other, ok := v.(*collector.MSMQConfig)
	if ok {
		setStringIfNotEmpty(c.Where, &other.MSMQWhereClause)
	}
	return ok
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
