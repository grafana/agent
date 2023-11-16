package vcenter

import (
	"testing"
	"time"

	"github.com/grafana/river"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver"
	"github.com/stretchr/testify/require"
)

func TestArguments_UnmarshalRiver(t *testing.T) {
	in := `
		endpoint = "http://localhost:1234"
		username = "user"
		password = "pass"
		collection_interval = "2m"

		output { /* no-op */ }
	`

	var args Arguments
	require.NoError(t, river.Unmarshal([]byte(in), &args))
	args.Convert()
	ext, err := args.Convert()
	require.NoError(t, err)
	otelArgs, ok := (ext).(*vcenterreceiver.Config)

	require.True(t, ok)

	require.Equal(t, "user", otelArgs.Username)
	require.Equal(t, "pass", string(otelArgs.Password))
	require.Equal(t, "http://localhost:1234", otelArgs.Endpoint)

	require.Equal(t, 2*time.Minute, otelArgs.ScraperControllerSettings.CollectionInterval)
	require.Equal(t, time.Second, otelArgs.ScraperControllerSettings.InitialDelay)
	require.Equal(t, 0*time.Second, otelArgs.ScraperControllerSettings.Timeout)

	require.Equal(t, true, otelArgs.Metrics.VcenterClusterCPUEffective.Enabled)
	require.Equal(t, false, otelArgs.Metrics.VcenterVMMemoryUtilization.Enabled)
	require.Equal(t, true, otelArgs.ResourceAttributes.VcenterClusterName.Enabled)
}
