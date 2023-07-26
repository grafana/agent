package docker

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	// Use host that works on all platforms (including Windows).
	var cfg = `
		host       = "tcp://127.0.0.1:9375"
		targets    = []
		forward_to = []
	`

	var args Arguments
	err := river.Unmarshal([]byte(cfg), &args)
	require.NoError(t, err)

	ctrl, err := componenttest.NewControllerFromID(util.TestLogger(t), "loki.source.docker")
	require.NoError(t, err)

	go func() {
		err := ctrl.Run(context.Background(), args)
		require.NoError(t, err)
	}()

	require.NoError(t, ctrl.WaitRunning(time.Minute))
}
