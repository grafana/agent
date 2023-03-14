package jaeger_remote_sampling_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/grafana/agent/component/otelcol/extension/jaeger_remote_sampling"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
)

// Test performs a basic integration test which runs the otelcol.extension.jaeger_remote_sampling
// component and ensures that it can be used for authentication.
func Test(t *testing.T) {
	// write remote sampling config to a temp file
	remoteSamplingConfig := `
	{
		"default_strategy": {
		  "type": "probabilistic",
		  "param": 0.5
		}
	}
	`
	expectedRemoteSamplingConfig := `
	{
		"strategyType": "PROBABILISTIC",
		"probabilisticSampling": {
			"samplingRate": 0.5
		}
	}
	`

	remoteSamplingConfigFile := filepath.Join(t.TempDir(), "remote.json")
	err := os.WriteFile(remoteSamplingConfigFile, []byte(remoteSamplingConfig), 0644)
	require.NoError(t, err)

	ctx := componenttest.TestContext(t)
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	l := util.TestLogger(t)

	// Create and run our component
	ctrl, err := componenttest.NewControllerFromID(l, "otelcol.extension.jaeger_remote_sampling")
	require.NoError(t, err)

	listenAddr := getFreeAddr(t)
	cfg := fmt.Sprintf(`
	    http {
			endpoint = "%s"
	    }
		grpc {
			endpoint = "%s"
		}
		source {
			file = "%s"
		}
	`, listenAddr, getFreeAddr(t), remoteSamplingConfigFile)
	var args jaeger_remote_sampling.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	go func() {
		err := ctrl.Run(ctx, args)
		require.NoError(t, err)
	}()

	require.NoError(t, ctrl.WaitRunning(time.Second), "component never started")
	// the wrapped jaeger remote sampler starts its http server async: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/extension/jaegerremotesampling/v0.63.0/extension/jaegerremotesampling/internal/http.go#L85
	// and reports errors back through ReportFatalError. Since we can't wait on this server directly just pause for a bit here while it starts up
	time.Sleep(time.Second)

	// request the remote sampling config above
	require.NoError(t, err)
	resp, err := http.Get("http://" + listenAddr + "/sampling?service=foo")
	require.NoError(t, err, "HTTP request failed")
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.JSONEq(t, string(b), expectedRemoteSamplingConfig)
}

func getFreeAddr(t *testing.T) string {
	t.Helper()

	portNumber, err := freeport.GetFreePort()
	require.NoError(t, err)

	return fmt.Sprintf("localhost:%d", portNumber)
}
