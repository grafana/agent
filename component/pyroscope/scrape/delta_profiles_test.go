package scrape

import (
	"context"
	"testing"

	googlev1 "github.com/grafana/phlare/api/gen/proto/go/google/v1"

	"github.com/grafana/agent/component/pyroscope"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
)

func TestDeltaProfilerAppender(t *testing.T) {
	strings := map[string]int{}
	pprofIn := googlev1.Profile{
		SampleType: []*googlev1.ValueType{
			{Type: addString(strings, "alloc_objects"), Unit: addString(strings, "count")},
			{Type: addString(strings, "alloc_space"), Unit: addString(strings, "bytes")},
		},
	}
	lbs := labels.Labels{
		{Name: model.MetricNameLabel, Value: pprofMemory},
	}
	rawIn, err := pprofIn.MarshalVT()
	require.NoError(t, err)

	outSamples := []*pyroscope.RawSample{}
	appender := NewDeltaAppender(pyroscope.AppendableFunc(func(ctx context.Context, lbs labels.Labels, samples []*pyroscope.RawSample) error {
		outSamples = append(outSamples, samples...)
		return nil
	}), lbs)

	err = appender.Append(context.Background(), lbs, []*pyroscope.RawSample{{RawProfile: rawIn}})
	require.NoError(t, err)

	require.Len(t, outSamples, 1)
	out := &googlev1.Profile{}
	err = out.UnmarshalVT(outSamples[0].RawProfile)
	require.NoError(t, err)
}

func addString(strings map[string]int, s string) int64 {
	i, ok := strings[s]
	if !ok {
		i = len(strings)
		strings[s] = i
	}
	return int64(i)
}
