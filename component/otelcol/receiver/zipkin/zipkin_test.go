package zipkin_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/grafana/agent/component/otelcol/receiver/zipkin"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zipkinreceiver"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	httpAddr := getFreeAddr(t)

	ctx := componenttest.TestContext(t)
	l := util.TestLogger(t)

	ctrl, err := componenttest.NewControllerFromID(l, "otelcol.receiver.zipkin")
	require.NoError(t, err)

	cfg := fmt.Sprintf(`
		endpoint = "%s"

		output { /* no-op */ }
	`, httpAddr)

	var args zipkin.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	go func() {
		err := ctrl.Run(ctx, args)
		require.NoError(t, err)
	}()

	require.NoError(t, ctrl.WaitRunning(time.Second))
}

func TestArguments_UnmarshalRiver(t *testing.T) {
	t.Run("grpc", func(t *testing.T) {
		httpAddr := getFreeAddr(t)
		in := fmt.Sprintf(`
		endpoint = "%s"
		cors {
			allowed_origins = ["https://*.test.com", "https://test.com"]
		}

		parse_string_tags = true

		output { /* no-op */ }
		`, httpAddr)

		var args zipkin.Arguments
		require.NoError(t, river.Unmarshal([]byte(in), &args))
		ext, err := args.Convert()
		require.NoError(t, err)
		otelArgs, ok := (ext).(*zipkinreceiver.Config)

		require.True(t, ok)

		// Check the arguments
		require.Equal(t, otelArgs.HTTPServerSettings.Endpoint, httpAddr)
		require.Equal(t, len(otelArgs.HTTPServerSettings.CORS.AllowedOrigins), 2)
		require.Equal(t, otelArgs.HTTPServerSettings.CORS.AllowedOrigins[0], "https://*.test.com")
		require.Equal(t, otelArgs.HTTPServerSettings.CORS.AllowedOrigins[1], "https://test.com")
		require.Equal(t, otelArgs.ParseStringTags, true)
	})
}

func getFreeAddr(t *testing.T) string {
	t.Helper()

	portNumber, err := freeport.GetFreePort()
	require.NoError(t, err)

	return fmt.Sprintf("localhost:%d", portNumber)
}
