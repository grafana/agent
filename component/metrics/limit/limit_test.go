package limit

import (
	"testing"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
)

func TestLengthLimit(t *testing.T) {
	passed := 0
	metricsRev := metrics.Receiver{
		Receive: func(_ int64, _ []*metrics.FlowMetric) {
			passed++
		},
	}
	l, err := New(component.Options{
		ID: "t1",
		OnStateChange: func(e component.Exports) {
		},
		Registerer: prometheus.NewRegistry(),
	}, Arguments{
		LabelLimit:  1,
		LengthLimit: 1,
		ForwardTo:   []*metrics.Receiver{&metricsRev},
	})

	require.NoError(t, err)
	require.NotNil(t, l)

	metricArr := []*metrics.FlowMetric{{
		GlobalRefID: 1,
		Labels: []labels.Label{{
			Name:  "t1",
			Value: "toolong",
		}},
		Value: 32,
	}}
	l.receiver.Receive(0, metricArr)
	require.True(t, passed == 0)
}

func TestAllowSome(t *testing.T) {
	passed := 0
	metricsRev := metrics.Receiver{
		Receive: func(_ int64, _ []*metrics.FlowMetric) {
			passed++
		},
	}
	l, err := New(component.Options{
		ID: "t1",
		OnStateChange: func(e component.Exports) {
		},
		Registerer: prometheus.NewRegistry(),
	}, Arguments{
		LabelLimit:  1,
		LengthLimit: 2,
		ForwardTo:   []*metrics.Receiver{&metricsRev},
	})

	require.NoError(t, err)
	require.NotNil(t, l)

	metricArr := []*metrics.FlowMetric{{
		GlobalRefID: 1,
		Labels: []labels.Label{{
			Name:  "t1",
			Value: "toolong",
		}},
		Value: 32,
	}, {
		GlobalRefID: 1,
		Labels: []labels.Label{{
			Name:  "t2",
			Value: "g",
		}},
		Value: 32,
	}}
	l.receiver.Receive(0, metricArr)
	require.True(t, passed == 1)
}
