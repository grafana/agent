package windows

import (
	"strings"
	"testing"

	windows_integration "github.com/grafana/agent/pkg/integrations/windows_exporter"
	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestRiverUnmarshalWithDefaultConfig(t *testing.T) {
	var args Arguments
	err := river.Unmarshal([]byte(""), &args)
	require.NoError(t, err)

	require.Equal(t, strings.Split(windows_integration.DefaultConfig.EnabledCollectors, ","), args.EnabledCollectors)
	require.Equal(t, strings.Split(windows_integration.DefaultConfig.Dfsr.SourcesEnabled, ","), args.Dfsr.SourcesEnabled)
	require.Equal(t, strings.Split(windows_integration.DefaultConfig.Exchange.EnabledList, ","), args.Exchange.EnabledList)
	require.Equal(t, windows_integration.DefaultConfig.IIS.AppExclude, args.IIS.AppExclude)
	require.Equal(t, windows_integration.DefaultConfig.IIS.AppInclude, args.IIS.AppInclude)
	require.Equal(t, windows_integration.DefaultConfig.IIS.SiteExclude, args.IIS.SiteExclude)
	require.Equal(t, windows_integration.DefaultConfig.IIS.SiteInclude, args.IIS.SiteInclude)
	require.Equal(t, windows_integration.DefaultConfig.LogicalDisk.Exclude, args.LogicalDisk.Exclude)
	require.Equal(t, windows_integration.DefaultConfig.LogicalDisk.Include, args.LogicalDisk.Include)
	require.Equal(t, windows_integration.DefaultConfig.MSMQ.Where, args.MSMQ.Where)
	require.Equal(t, strings.Split(windows_integration.DefaultConfig.MSSQL.EnabledClasses, ","), args.MSSQL.EnabledClasses)
	require.Equal(t, windows_integration.DefaultConfig.Network.Exclude, args.Network.Exclude)
	require.Equal(t, windows_integration.DefaultConfig.Network.Include, args.Network.Include)
	require.Equal(t, windows_integration.DefaultConfig.Process.Exclude, args.Process.Exclude)
	require.Equal(t, windows_integration.DefaultConfig.Process.Include, args.Process.Include)
	require.Equal(t, windows_integration.DefaultConfig.ScheduledTask.Exclude, args.ScheduledTask.Exclude)
	require.Equal(t, windows_integration.DefaultConfig.ScheduledTask.Include, args.ScheduledTask.Include)
	require.Equal(t, windows_integration.DefaultConfig.Service.UseApi, args.Service.UseApi)
	require.Equal(t, windows_integration.DefaultConfig.Service.Where, args.Service.Where)
	require.Equal(t, windows_integration.DefaultConfig.SMTP.Exclude, args.SMTP.Exclude)
	require.Equal(t, windows_integration.DefaultConfig.SMTP.Include, args.SMTP.Include)
	require.Equal(t, windows_integration.DefaultConfig.TextFile.TextFileDirectory, args.TextFile.TextFileDirectory)
}
