package vcenter

import (
	"fmt"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/river"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
)

// Test ensures that otelcol.receiver.vcenter can start successfully.
func Test(t *testing.T) {
	httpAddr := getFreeAddr(t)

	ctx := componenttest.TestContext(t)
	l := util.TestLogger(t)

	ctrl, err := componenttest.NewControllerFromID(l, "otelcol.receiver.vcenter")
	require.NoError(t, err)

	cfg := fmt.Sprintf(`
		endpoint = "%s"
		username = "user"
		password = "pass"

		output { /* no-op */ }
	`, httpAddr)

	var args Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	go func() {
		err := ctrl.Run(ctx, args)
		require.NoError(t, err)
	}()

	require.NoError(t, ctrl.WaitRunning(time.Second))
}

func TestArguments_UnmarshalRiver(t *testing.T) {
	httpAddr := getFreeAddr(t)
	in := fmt.Sprintf(`
		endpoint = "%s"
		username = "user"
		password = "pass"
		collection_interval = "2m"

		output { /* no-op */ }
	`, httpAddr)

	var args Arguments
	require.NoError(t, river.Unmarshal([]byte(in), &args))
	args.Convert()
	ext, err := args.Convert()
	require.NoError(t, err)
	otelArgs, ok := (ext).(*vcenterreceiver.Config)

	require.True(t, ok)

	require.Equal(t, "user", otelArgs.Username)
	require.Equal(t, "pass", string(otelArgs.Password))
	require.Equal(t, httpAddr, otelArgs.Endpoint)

	require.Equal(t, 2*time.Minute, otelArgs.ScraperControllerSettings.CollectionInterval)
	require.Equal(t, time.Second, otelArgs.ScraperControllerSettings.InitialDelay)
	require.Equal(t, 0*time.Second, otelArgs.ScraperControllerSettings.Timeout)

}

func getFreeAddr(t *testing.T) string {
	t.Helper()

	portNumber, err := freeport.GetFreePort()
	require.NoError(t, err)

	return fmt.Sprintf("http://localhost:%d", portNumber)
}
