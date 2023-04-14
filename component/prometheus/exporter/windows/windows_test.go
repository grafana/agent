package windows

import (
	"testing"
	"time"

	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

var (
	exampleRiverConfig = `
		enabled_collectors = "textfile"
		
		exchange {
			enabled_list = "example"
		}
		
		iis {
			site_whitelist = ".+"
			site_blacklist = ""
			app_whitelist = ".+"
			app_blacklist = ""
		}
		
		text_file {
			text_file_directory = "C:"
		}
		
		smtp {
			whitelist = ".+"
			blacklist = ""
		}

        service {
            where_clause = "where"
        }
		
		process {
			whitelist = ".+"
			blacklist = ""
		}
		
		network {
			whitelist = ".+"
			blacklist = ""
		}
		
		mssql {
			enabled_classes = "accessmethods"
		}
		
		msmq {
            where_clause = "where"
		}
		
		logical_disk {
			whitelist = ".+"
			blacklist = ""
		}
		`
	duration1m, _ = time.ParseDuration("1m")
)

func TestRiverUnmarshal(t *testing.T) {
	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)

	require.Equal(t, "textfile", args.EnabledCollectors)
	require.Equal(t, "example", args.Exchange.EnabledList)
	require.Equal(t, "", args.IIS.SiteBlackList)
	require.Equal(t, ".+", args.IIS.SiteWhiteList)
	require.Equal(t, "", args.IIS.AppBlackList)
	require.Equal(t, ".+", args.IIS.AppWhiteList)
	require.Equal(t, "C:", args.TextFile.TextFileDirectory)
	require.Equal(t, "", args.SMTP.BlackList)
	require.Equal(t, ".+", args.SMTP.WhiteList)
	require.Equal(t, "where", args.Service.Where)
	require.Equal(t, "", args.Process.BlackList)
	require.Equal(t, ".+", args.Process.WhiteList)
	require.Equal(t, "", args.Network.BlackList)
	require.Equal(t, ".+", args.Network.WhiteList)
	require.Equal(t, "accessmethods", args.MSSQL.EnabledClasses)
	require.Equal(t, "where", args.MSMQ.Where)
	require.Equal(t, "", args.LogicalDisk.BlackList)
	require.Equal(t, ".+", args.LogicalDisk.WhiteList)
}

func TestConvert(t *testing.T) {
	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)

	conf := args.Convert()

	require.Equal(t, "textfile", conf.EnabledCollectors)
	require.Equal(t, "example", conf.Exchange.EnabledList)
	require.Equal(t, "", conf.IIS.SiteBlackList)
	require.Equal(t, ".+", conf.IIS.SiteWhiteList)
	require.Equal(t, "", conf.IIS.AppBlackList)
	require.Equal(t, ".+", conf.IIS.AppWhiteList)
	require.Equal(t, "C:", conf.TextFile.TextFileDirectory)
	require.Equal(t, "", conf.SMTP.BlackList)
	require.Equal(t, ".+", conf.SMTP.WhiteList)
	require.Equal(t, "where", conf.Service.Where)
	require.Equal(t, "", conf.Process.BlackList)
	require.Equal(t, ".+", conf.Process.WhiteList)
	require.Equal(t, "", conf.Network.BlackList)
	require.Equal(t, ".+", conf.Network.WhiteList)
	require.Equal(t, "accessmethods", conf.MSSQL.EnabledClasses)
	require.Equal(t, "where", conf.MSMQ.Where)
	require.Equal(t, "", conf.LogicalDisk.BlackList)
	require.Equal(t, ".+", conf.LogicalDisk.WhiteList)
}
