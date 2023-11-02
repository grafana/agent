package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/stretchr/testify/require"
)

func TestMetrcsGeneration(t *testing.T) {
	ctlr, err := componenttest.NewControllerFromID(nil, "prometheus.test.metrics")
	require.NoError(t, err)
	require.NotNil(t, ctlr)
	ctx := context.Background()
	ctx, cncl := context.WithTimeout(ctx, 10*time.Second)
	ctlr.Run(ctx,Arguments{

	}
}
