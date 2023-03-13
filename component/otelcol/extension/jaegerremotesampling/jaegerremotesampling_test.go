package jaegerremotesampling_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/grafana/agent/component/otelcol/extension"
	"github.com/grafana/agent/component/otelcol/extension/jaegerremotesampling"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
)

// Test performs a basic integration test which runs the otelcol.extension.jaegerremotesampling
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
	err := ioutil.WriteFile(remoteSamplingConfigFile, []byte(remoteSamplingConfig), 0644)
	require.NoError(t, err)

	ctx := componenttest.TestContext(t)
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	l := util.TestLogger(t)

	// Create and run our component
	ctrl, err := componenttest.NewControllerFromID(l, "otelcol.extension.jaegerremotesampling")
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
	var args jaegerremotesampling.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	go func() {
		err := ctrl.Run(ctx, args)
		require.NoError(t, err)
	}()

	require.NoError(t, ctrl.WaitRunning(time.Second), "component never started")
	require.NoError(t, ctrl.WaitExports(time.Second), "component never exported anything")
	// the wrapped jaeger remote sampler starts its http server async: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/extension/jaegerremotesampling/v0.63.0/extension/jaegerremotesampling/internal/http.go#L85
	// and reports errors back through ReportFatalError. Since we can't wait on this server directly just pause for a bit here while it starts up
	time.Sleep(time.Second)

	// Get the authentication extension from our component and use it to make a
	// request to our test server.
	exports := ctrl.Exports().(extension.Exports)
	require.NotNil(t, exports.Handler.Extension, "handler extension is nil")

	// request the remote sampling config above
	require.NoError(t, err)
	resp, err := http.Get("http://" + listenAddr + "/sampling?service=foo")
	require.NoError(t, err, "HTTP request failed")
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.JSONEq(t, string(b), expectedRemoteSamplingConfig)
}

func getFreeAddr(t *testing.T) string {
	t.Helper()

	portNumber, err := freeport.GetFreePort()
	require.NoError(t, err)

	return fmt.Sprintf("localhost:%d", portNumber)
}
