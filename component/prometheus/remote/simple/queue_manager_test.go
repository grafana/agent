package simple

import (
	"strconv"
	"testing"

	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/require"
)

func TestSmallQueue(t *testing.T) {
	protoSamples := make([]prompb.TimeSeries, 0)
	for i := 0; i < 100; i++ {
		protoSamples = append(protoSamples, prompb.TimeSeries{
			Labels: []prompb.Label{prompb.Label{Name: strconv.Itoa(i)}},
		})
	}
	queues := fillQueues(protoSamples, 500)
	require.Len(t, queues, 1)
	// 1 batch
	require.Len(t, queues[0], 1)
	require.Len(t, queues[0][0], 100)

	checkQueues(t, queues, 100)
}

func TestJaggedQueue(t *testing.T) {
	protoSamples := make([]prompb.TimeSeries, 0)
	for i := 0; i < 11; i++ {
		protoSamples = append(protoSamples, prompb.TimeSeries{
			Labels: []prompb.Label{prompb.Label{Name: strconv.Itoa(i)}},
		})
	}
	queues := fillQueues(protoSamples, 5)
	require.Len(t, queues, 3)

	require.Len(t, queues[0], 1)
	require.Len(t, queues[0][0], 5)

	require.Len(t, queues[1], 1)
	require.Len(t, queues[1][0], 5)

	require.Len(t, queues[2], 1)
	require.Len(t, queues[2][0], 1)

	checkQueues(t, queues, 11)
}

func checkQueues(t *testing.T, queues map[int][][]prompb.TimeSeries, total int) {
	foundTotal := 0
	for i := 0; i < total; i++ {
		found := false
		for _, v := range queues {
			for _, b := range v {
				for _, sample := range b {
					if sample.Labels[0].Name == strconv.Itoa(i) {
						found = true
						foundTotal++
						break
					}
				}
				if found {
					break
				}
			}
			if found {
				break
			}
		}
		require.True(t, found)
	}
	require.True(t, foundTotal == total)
}
