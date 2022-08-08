package generator

import (
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestGenerator(t *testing.T) {
	g := &Generator{
		opts: component.Options{
			ID: "test",
			OnStateChange: func(e component.Exports) {
			},
			Registerer: prometheus.NewRegistry(),
		},

		args: Arguments{
			LabelCount:     1,
			LabelLength:    1,
			SeriesCount:    1,
			MetricCount:    1,
			ScrapeInterval: 1 * time.Minute,
			ForwardTo:      []*metrics.Receiver{},
		},
		health: component.Health{},
	}
	require.NotNil(t, g)
	metricArr := g.generate()
	require.Len(t, metricArr, 1)
	for _, m := range metricArr {
		require.True(t, m.Labels.Has("__name__"))
		// Labels should always have a __name__ + the 1 specified
		require.Len(t, m.Labels, 2)
		// we go with the second item here because __name__ is the first
		require.Len(t, m.Labels[1].Value, 1)
	}

}
