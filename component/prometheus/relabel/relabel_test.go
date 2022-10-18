package relabel

import (
	"math"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/prometheus"
	"github.com/grafana/agent/pkg/util"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/prometheus/prometheus/model/value"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	relabeller := generateRelabel(t)
	fm := prometheus.NewFlowMetric(0, labels.FromStrings("__address__", "localhost"), 0)

	relabeller.Receive(time.Now().Unix(), []*prometheus.FlowMetric{fm})
	require.Len(t, relabeller.cache, 1)
	entry, found := relabeller.cache[fm.GlobalRefID()]
	require.True(t, found)
	require.NotNil(t, entry)
	require.True(t, entry.GlobalRefID() != fm.GlobalRefID())
}

func TestEviction(t *testing.T) {
	relabeller := generateRelabel(t)
	fm := prometheus.NewFlowMetric(0, labels.FromStrings("__address__", "localhost"), 0)

	relabeller.Receive(time.Now().Unix(), []*prometheus.FlowMetric{fm})
	require.Len(t, relabeller.cache, 1)
	fmstale := prometheus.NewFlowMetric(0, labels.FromStrings("__address__", "localhost"), math.Float64frombits(value.StaleNaN))
	relabeller.Receive(time.Now().Unix(), []*prometheus.FlowMetric{fmstale})
	require.Len(t, relabeller.cache, 0)
}

func TestUpdateReset(t *testing.T) {
	relabeller := generateRelabel(t)
	fm := prometheus.NewFlowMetric(0, labels.FromStrings("__address__", "localhost"), 0)

	relabeller.Receive(time.Now().Unix(), []*prometheus.FlowMetric{fm})
	require.Len(t, relabeller.cache, 1)
	relabeller.Update(Arguments{
		MetricRelabelConfigs: []*flow_relabel.Config{},
	})
	require.Len(t, relabeller.cache, 0)
}

func generateRelabel(t *testing.T) *Component {
	rec := &prometheus.Receiver{
		Receive: func(timestamp int64, metrics []*prometheus.FlowMetric) {
			require.True(t, metrics[0].LabelsCopy().Has("new_label"))
		},
	}
	relabeller, err := New(component.Options{
		ID:     "1",
		Logger: util.TestLogger(t),
		OnStateChange: func(e component.Exports) {
		},
		Registerer: prom.NewRegistry(),
	}, Arguments{
		ForwardTo: []*prometheus.Receiver{rec},
		MetricRelabelConfigs: []*flow_relabel.Config{
			{
				SourceLabels: []string{"__address__"},
				Regex:        flow_relabel.Regexp(relabel.MustNewRegexp("(.+)")),
				TargetLabel:  "new_label",
				Replacement:  "new_value",
				Action:       "replace",
			},
		},
	})
	require.NotNil(t, relabeller)
	require.NoError(t, err)
	return relabeller
}
