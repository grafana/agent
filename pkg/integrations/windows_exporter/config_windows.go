package windows_exporter //nolint:golint

import (
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/collector"
	"strconv"
)

// Populate defaults for all collector configs.
func init() {
	// TODO (@mattdurham) we should look at removing this init. I think it can become
	// a function call now.
	// Register flags from all collector configs to a fake integration and then
	// parse an empty command line to force defaults to be populated.
	app := kingpin.New("", "")

	// Register all flags from collector
	collectors := collector.CreateCollectors()
	collector.RegisterCollectorsFlags(collectors, app)

	_, err := app.Parse([]string{})
	if err != nil {
		panic(err)
	}

	// Map the configs with defaults applied to our default config.
	DefaultConfig.fromExporterConfig(app)
}

func getDefault(app *kingpin.Application, name string) string {
	for _, f := range app.Model().Flags {
		if f.Name != name {
			continue
		}
		return f.String()
	}
	return ""
}

// fromExporterConfig converts windows_exporter configs into the integration Config.
func (c *Config) fromExporterConfig(app *kingpin.Application) {
	c.Dfsr.SourcesEnabled = getDefault(app, collector.FlagDfsrEnabledCollectors)
	c.Exchange.EnabledList = getDefault(app, collector.FlagExchangeCollectorsEnabled)
	c.IIS.SiteBlackList = getDefault(app, collector.FlagIISSiteOldExclude)
	c.IIS.SiteWhiteList = getDefault(app, collector.FlagIISSiteOldInclude)
	c.IIS.AppBlackList = getDefault(app, collector.FlagIISAppOldExclude)
	c.IIS.AppWhiteList = getDefault(app, collector.FlagIISAppOldInclude)
	c.IIS.SiteExclude = getDefault(app, collector.FlagIISSiteExclude)
	c.IIS.SiteInclude = getDefault(app, collector.FlagIISSiteInclude)
	c.IIS.AppExclude = getDefault(app, collector.FlagIISAppExclude)
	c.IIS.AppInclude = getDefault(app, collector.FlagIISAppInclude)
	c.LogicalDisk.BlackList = getDefault(app, collector.FlagLogicalDiskVolumeOldExclude)
	c.LogicalDisk.WhiteList = getDefault(app, collector.FlagLogicalDiskVolumeOldInclude)
	c.LogicalDisk.Exclude = getDefault(app, collector.FlagLogicalDiskVolumeExclude)
	c.LogicalDisk.Include = getDefault(app, collector.FlagLogicalDiskVolumeInclude)
	c.MSMQ.Where = getDefault(app, collector.FlagMsmqWhereClause)
	c.MSSQL.EnabledClasses = getDefault(app, collector.FlagMssqlEnabledCollectors)
	c.Network.BlackList = getDefault(app, collector.FlagNicOldExclude)
	c.Network.WhiteList = getDefault(app, collector.FlagNicOldInclude)
	c.Network.Exclude = getDefault(app, collector.FlagNicExclude)
	c.Network.Include = getDefault(app, collector.FlagNicInclude)
	c.Process.BlackList = getDefault(app, collector.FlagProcessOldExclude)
	c.Process.WhiteList = getDefault(app, collector.FlagProcessOldInclude)
	c.Process.Exclude = getDefault(app, collector.FlagProcessExclude)
	c.Process.Include = getDefault(app, collector.FlagProcessInclude)
	c.ScheduledTask.Exclude = getDefault(app, collector.FlagScheduledTaskExclude)
	c.ScheduledTask.Include = getDefault(app, collector.FlagScheduledTaskInclude)
	c.Service.Where = getDefault(app, collector.FlagServiceWhereClause)
	useApi, _ := strconv.ParseBool(getDefault(app, collector.FlagServiceUseAPI))
	c.Service.UseApi = useApi
	c.SMTP.BlackList = getDefault(app, collector.FlagSmtpServerOldExclude)
	c.SMTP.WhiteList = getDefault(app, collector.FlagSmtpServerOldInclude)
	c.SMTP.Exclude = getDefault(app, collector.FlagSmtpServerExclude)
	c.SMTP.Include = getDefault(app, collector.FlagSmtpServerInclude)
	c.TextFile.TextFileDirectory = getDefault(app, collector.FlagTextFileDirectory)
}

// toExporterConfig converts integration Configs into windows_exporter configs.
func (c *Config) toExporterConfig(collectors map[string]*collector.CollectorInit) error {
	for _, v := range collectors {
		if v.Settings == nil {
			continue
		}
		switch t := v.Settings.(type) {
		case *collector.DFRSSettings:
			t.DFRSEnabledCollectors = &c.Dfsr.SourcesEnabled
		case *collector.DiskSettings:
			t.VolumeInclude = &c.LogicalDisk.Include
			t.VolumeExclude = &c.LogicalDisk.Exclude
		case *collector.ExchangeSettings:
			t.ArgExchangeCollectorsEnabled = &c.Exchange.EnabledList
		case *collector.IISSettings:
			t.AppInclude = &c.IIS.AppInclude
			t.AppExclude = &c.IIS.AppExclude
			t.OldAppExclude = &c.IIS.AppBlackList
			t.OldAppInclude = &c.IIS.AppWhiteList
			t.SiteExclude = &c.IIS.SiteExclude
			t.SiteInclude = &c.IIS.SiteInclude
			t.OldSiteExclude = &c.IIS.SiteBlackList
			t.OldSiteInclude = &c.IIS.SiteWhiteList
		case *collector.MSMQSettings:
			t.MSMQWhereClause = &c.MSMQ.Where
		case *collector.MSSqlSettings:
			t.ClassesEnabled = &c.MSSQL.EnabledClasses
		case *collector.NetSettings:
			t.NicInclude = &c.Network.Include
			t.NicExclude = &c.Network.Exclude
			t.NicOldExclude = &c.Network.BlackList
			t.NicOldInclude = &c.Network.WhiteList
		case *collector.ProcessSettings:
			t.ProcessExclude = &c.Process.Exclude
			t.ProcessInclude = &c.Process.Include
			t.ProcessOldExclude = &c.Process.BlackList
			t.ProcessOldInclude = &c.Process.WhiteList
		case *collector.ServiceSettings:
			t.UseAPI = &c.Service.UseApi
			t.ServiceWhereClause = &c.Service.Where
		case *collector.SMTPSettings:
			t.ServerInclude = &c.SMTP.Include
			t.ServerExclude = &c.SMTP.Exclude
			t.ServerOldExclude = &c.SMTP.BlackList
			t.ServerOldInclude = &c.SMTP.WhiteList
		case *collector.TaskSettings:
			t.TaskInclude = &c.ScheduledTask.Include
			t.TaskExclude = &c.ScheduledTask.Exclude
		case *collector.TextSettings:
			t.TextFileDirectory = &c.TextFile.TextFileDirectory
		default:
			return fmt.Errorf("unknown windows exporter type %t", t)
		}
	}
	return nil
}

var _ kingpin.Value = (*SValue)(nil)

type SValue struct {
	val string
}

func (S *SValue) String() string {
	return S.val
}

func (S *SValue) Set(s string) error {
	S.val = s
	return nil
}
