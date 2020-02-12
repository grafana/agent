package wal

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/util/test"
	"github.com/go-kit/kit/log"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/timestamp"
	"github.com/prometheus/prometheus/tsdb/record"
	"github.com/prometheus/prometheus/tsdb/wal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	walDir, err := ioutil.TempDir(os.TempDir(), "wal")
	require.NoError(t, err)
	defer os.RemoveAll(walDir)
	fmt.Println("walDir", walDir)

	s, err := NewStorage(log.NewNopLogger(), nil, walDir)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, s.Close())
	}()

	app, err := s.Appender()
	require.NoError(t, err)

	collector := walDataCollector{}

	watcher := wal.NewWatcher(nil, wal.NewWatcherMetrics(nil), nil, "test", &collector, walDir)
	go watcher.Start()
	defer watcher.Stop()

	type tsPair struct {
		ts  int64
		val float64
	}

	payload := []struct {
		name string
		data []tsPair
	}{
		{name: "foo", data: []tsPair{{1, 10.0}, {10, 100.0}}},
		{name: "bar", data: []tsPair{{2, 20.0}, {20, 200.0}}},
		{name: "baz", data: []tsPair{{3, 30.0}, {30, 300.0}}},
	}

	// Sample timestamps have to at least be after when the WAL watcher started
	baseTs := timestamp.FromTime(time.Now())

	for _, metric := range payload {
		labels := labels.FromMap(map[string]string{"__name__": metric.name})
		ref, err := app.Add(labels, baseTs+metric.data[0].ts, metric.data[1].val)
		require.NoError(t, err)

		// Write other data points with AddFast
		for _, sample := range metric.data[1:] {
			err := app.AddFast(labels, ref, baseTs+sample.ts, sample.val)
			require.NoError(t, err)
		}
	}

	require.NoError(t, app.Commit())

	// Wait for series to be written. Expect them to be in same order from earlier.
	test.Poll(t, 30*time.Second, true, func() interface{} {
		collector.mut.Lock()
		defer collector.mut.Unlock()

		names := []string{}
		for _, series := range collector.series {
			names = append(names, series.Labels.Get("__name__"))
		}

		return assert.ObjectsAreEqual([]string{"foo", "bar", "baz"}, names)
	})

	expectedSamples := []record.RefSample{}
	for ref, metric := range payload {
		// Only the last sample from each call Append per series should be written;
		// only look for that.
		sample := metric.data[len(metric.data)-1]

		expectedSamples = append(expectedSamples, record.RefSample{
			Ref: uint64(ref),
			T:   baseTs + sample.ts,
			V:   sample.val,
		})
	}

	// Wait for samples to be written.
	test.Poll(t, 30*time.Second, true, func() interface{} {
		collector.mut.Lock()
		defer collector.mut.Unlock()

		return assert.ObjectsAreEqual(expectedSamples, collector.samples)
	})
}

func TestStorage_ExistingWAL(t *testing.T)           { t.Skip("NYI") }
func TestStorage_Truncate(t *testing.T)              { t.Skip("NYI") }
func TestStorage_WriteStalenessMarkers(t *testing.T) { t.Skip("NYI") }

type walDataCollector struct {
	mut     sync.Mutex
	samples []record.RefSample
	series  []record.RefSeries
}

func (c *walDataCollector) Append(samples []record.RefSample) bool {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.samples = append(c.samples, samples...)
	return true
}

func (c *walDataCollector) StoreSeries(series []record.RefSeries, _ int) {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.series = append(c.series, series...)
}

func (c *walDataCollector) SeriesReset(_ int) {}
