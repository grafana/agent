package windows_exporter //nolint:golint

import (
	"github.com/prometheus-community/windows_exporter/collector"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Populate defaults for all collector configs.
func init() {
	// Register flags from all collector configs to a fake integration and then
	// parse an empty command line to force defaults to be populated.
	ka := kingpin.New("init", "")

	configs := collector.AllConfigs()
	for _, cfg := range configs {
		cfg.RegisterFlags(ka)
	}
	_, err := ka.Parse(nil)
	if err != nil {
		panic(err)
	}

	// Map the configs with defaults applied to our default config.
	DefaultConfig.fromExporterConfig(configs)
}

// fromExporterConfig converts windows_exporter configs into the integration Config.
func (c *Config) fromExporterConfig(configs []collector.Config) {
	for _, ec := range configs {
		switch other := ec.(type) {
		case *collector.ExchangeConfig:
			c.Exchange.EnabledList = other.Enabled

		case *collector.IISConfig:
			c.IIS.SiteWhiteList = other.SiteWhitelist
			c.IIS.SiteBlackList = other.SiteBlacklist
			c.IIS.AppWhiteList = other.AppWhitelist
			c.IIS.AppBlackList = other.AppBlacklist

		case *collector.TextFileConfig:
			c.TextFile.TextFileDirectory = other.Directory

		case *collector.SMTPConfig:
			c.SMTP.WhiteList = other.ServerWhitelist
			c.SMTP.BlackList = other.ServerBlacklist

		case *collector.ServiceConfig:
			c.Service.Where = other.WhereClause

		case *collector.ProcessConfig:
			c.Process.WhiteList = other.ProcessWhitelist
			c.Process.BlackList = other.ProcessBlacklist

		case *collector.NetworkConfig:
			c.Network.WhiteList = other.NICWhitelist
			c.Network.BlackList = other.NICBlacklist

		case *collector.MSSQLConfig:
			c.MSSQL.EnabledClasses = other.EnabledCollectors

		case *collector.MSMQConfig:
			c.MSMQ.Where = other.WhereClause

		case *collector.LogicalDiskConfig:
			c.LogicalDisk.WhiteList = other.VolumeWhitelist
			c.LogicalDisk.BlackList = other.VolumeBlacklist
		}
	}
}

// toExporterConfig converts integration Configs into windows_exporter configs.
func (c *Config) toExporterConfig(configs []collector.Config) {
	for _, ec := range configs {
		switch other := ec.(type) {
		case *collector.ExchangeConfig:
			other.Enabled = c.Exchange.EnabledList

		case *collector.IISConfig:
			other.SiteWhitelist = c.IIS.SiteWhiteList
			other.SiteBlacklist = c.IIS.SiteBlackList
			other.AppWhitelist = c.IIS.AppWhiteList
			other.AppBlacklist = c.IIS.AppBlackList

		case *collector.TextFileConfig:
			other.Directory = c.TextFile.TextFileDirectory

		case *collector.SMTPConfig:
			other.ServerWhitelist = c.SMTP.WhiteList
			other.ServerBlacklist = c.SMTP.BlackList

		case *collector.ServiceConfig:
			other.WhereClause = c.Service.Where

		case *collector.ProcessConfig:
			other.ProcessWhitelist = c.Process.WhiteList
			other.ProcessBlacklist = c.Process.BlackList

		case *collector.NetworkConfig:
			other.NICWhitelist = c.Network.WhiteList
			other.NICBlacklist = c.Network.BlackList

		case *collector.MSSQLConfig:
			other.EnabledCollectors = c.MSSQL.EnabledClasses

		case *collector.MSMQConfig:
			other.WhereClause = c.MSMQ.Where

		case *collector.LogicalDiskConfig:
			other.VolumeWhitelist = c.LogicalDisk.WhiteList
			other.VolumeBlacklist = c.LogicalDisk.BlackList
		}
	}
}
