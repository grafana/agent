package prometheus

import (
	"testing"

	"github.com/prometheus/prometheus/model/labels"
	promrelabel "github.com/prometheus/prometheus/model/relabel"
	"github.com/stretchr/testify/require"
)

func TestRelabel(t *testing.T) {
	fm := NewFlowMetric(0, labels.FromStrings("key", "value"), 0)
	require.True(t, fm.globalRefID != 0)
	rg, _ := promrelabel.NewRegexp("(.*)")
	newfm := fm.Relabel(&promrelabel.Config{
		Replacement: "${1}_new",
		Action:      "replace",
		TargetLabel: "new",
		Regex:       rg,
	})
	require.Len(t, fm.labels, 1)
	require.True(t, fm.labels.Has("key"))

	require.Len(t, newfm.labels, 2)
	require.True(t, newfm.labels.Has("new"))
}

func TestRelabelTheSame(t *testing.T) {
	fm := NewFlowMetric(0, labels.FromStrings("key", "value"), 0)
	require.True(t, fm.globalRefID != 0)
	rg, _ := promrelabel.NewRegexp("bad")
	newfm := fm.Relabel(&promrelabel.Config{
		Replacement: "${1}_new",
		Action:      "replace",
		TargetLabel: "new",
		Regex:       rg,
	})
	require.Len(t, fm.labels, 1)
	require.True(t, fm.labels.Has("key"))
	require.Len(t, newfm.labels, 1)
	require.True(t, newfm.globalRefID == fm.globalRefID)
	require.True(t, labels.Equal(newfm.labels, fm.labels))
}
