package linear

import (
	"bytes"
	"testing"
	"time"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
)

func TestLinear(t *testing.T) {
	l := newLinear()
	lbls := labels.FromMap(map[string]string{
		"__name__": "test",
	})
	ts := time.Now().Unix()
	l.AddMetric(lbls, ts, 10)

	bb := &bytes.Buffer{}
	l.Serialize(bb)
	out := bytes.NewBuffer(bb.Bytes())
	metrics, err := l.Deserialize(out, 100)
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Len(t, metrics[0].lbls, 1)

	require.True(t, metrics[0].lbls[0].Name == "__name__")
	require.True(t, metrics[0].lbls[0].Value == "test")
}

func TestLinearMultiple(t *testing.T) {
	l := newLinear()
	lbls := labels.FromMap(map[string]string{
		"__name__": "test",
	})
	ts := time.Now().Unix()
	l.AddMetric(lbls, ts, 10)

	lbls2 := labels.FromMap(map[string]string{
		"__name__": "test",
		"lbl":      "label_1",
	})

	l.AddMetric(lbls2, ts, 11)

	bb := &bytes.Buffer{}
	l.Serialize(bb)
	out := bytes.NewBuffer(bb.Bytes())
	metrics, err := l.Deserialize(out, 100)
	require.NoError(t, err)
	require.Len(t, metrics, 2)

	require.True(t, hasLabel(lbls, metrics, ts, 10))
	require.True(t, hasLabel(lbls2, metrics, ts, 11))
}

func TestLinearReuse(t *testing.T) {
	l := LinearPool.Get().(*linear)
	lbls := labels.FromMap(map[string]string{
		"__name__": "test",
	})
	ts := time.Now().Unix()
	l.AddMetric(lbls, ts, 10)

	lbls2 := labels.FromMap(map[string]string{
		"__name__": "test",
		"lbl":      "label_1",
	})
	l.AddMetric(lbls2, ts, 11)

	bb := &bytes.Buffer{}
	l.Serialize(bb)
	out := bytes.NewBuffer(bb.Bytes())
	metrics, err := l.Deserialize(out, 100)
	require.NoError(t, err)
	require.Len(t, metrics, 2)

	require.True(t, hasLabel(lbls, metrics, ts, 10))
	require.True(t, hasLabel(lbls2, metrics, ts, 11))

	l.Reset()
	LinearPool.Put(l)

	l = LinearPool.Get().(*linear)
	l.AddMetric(lbls, ts, 10)
	bb = &bytes.Buffer{}
	l.Serialize(bb)
	out = bytes.NewBuffer(bb.Bytes())
	metrics, err = l.Deserialize(out, 100)
	require.NoError(t, err)
	require.Len(t, metrics, 1)

	require.True(t, hasLabel(lbls, metrics, ts, 10))
}

func hasLabel(lbls labels.Labels, metrics []*deserializedMetric, ts int64, val float64) bool {
	for _, m := range metrics {
		if labels.Compare(m.lbls, lbls) == 0 {
			return ts == m.ts && val == m.val
		}
	}
	return false
}
