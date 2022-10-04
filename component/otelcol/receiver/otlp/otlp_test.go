package otlp_test

import (
	"testing"
	"time"

	"github.com/grafana/agent/component/otelcol/receiver/otlp"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
)

// Test performs a basic test which ensures the component can run.
func Test(t *testing.T) {
	ctrl, err := componenttest.NewControllerFromID(util.TestLogger(t), "otelcol.receiver.otlp")
	require.NoError(t, err)

	cfg := `
		grpc {}
		http {}

		output {
			metrics = []
			logs    = []
			traces  = []
		}
	`
	var args otlp.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	go func() {
		err := ctrl.Run(componenttest.TestContext(t), args)
		require.NoError(t, err)
	}()

	require.NoError(t, ctrl.WaitRunning(time.Second))
}
