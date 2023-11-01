package wal

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/tsdb/record"
	"github.com/stretchr/testify/require"

	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/common/loki/utils"
	"github.com/grafana/loki/pkg/ingester/wal"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/grafana/loki/pkg/util"
)

type testWriteTo struct {
	ReadEntries         *utils.SyncSlice[loki.Entry]
	series              map[uint64]model.LabelSet
	logger              log.Logger
	ReceivedSeriesReset []int
}

func (t *testWriteTo) StoreSeries(series []record.RefSeries, _ int) {
	for _, seriesRec := range series {
		t.series[uint64(seriesRec.Ref)] = util.MapToModelLabelSet(seriesRec.Labels.Map())
	}
}

func (t *testWriteTo) SeriesReset(segmentNum int) {
	level.Debug(t.logger).Log("msg", fmt.Sprintf("received series reset with %d", segmentNum))
	t.ReceivedSeriesReset = append(t.ReceivedSeriesReset, segmentNum)
}

func (t *testWriteTo) AppendEntries(entries wal.RefEntries, _ int) error {
	var entry loki.Entry
	if l, ok := t.series[uint64(entries.Ref)]; ok {
		entry.Labels = l
		for _, e := range entries.Entries {
			entry.Entry = e
			t.ReadEntries.Append(entry)
		}
	} else {
		level.Debug(t.logger).Log("series for entry not found")
	}
	return nil
}

func (t *testWriteTo) AssertContainsLines(tst *testing.T, lines ...string) {
	seen := map[string]bool{}
	for _, l := range lines {
		seen[l] = false
	}
	for _, e := range t.ReadEntries.StartIterate() {
		if _, ok := seen[e.Line]; ok {
			seen[e.Line] = true
		}
	}
	t.ReadEntries.DoneIterate()

	allSeen := true
	for _, wasSeen := range seen {
		allSeen = allSeen && wasSeen
	}

	require.True(tst, allSeen, "expected all entries to have been received")
}

// watcherTestResources contains all resources necessary to test an individual Watcher functionality
type watcherTestResources struct {
	writeEntry             func(entry loki.Entry)
	notifyWrite            func()
	startWatcher           func()
	syncWAL                func() error
	nextWALSegment         func() error
	writeTo                *testWriteTo
	notifySegmentReclaimed func(segmentNum int)
}

type watcherTest func(t *testing.T, res *watcherTestResources)

// cases defines the watcher test cases
var cases = map[string]watcherTest{
	"read entries from WAL": func(t *testing.T, res *watcherTestResources) {
		res.startWatcher()

		lines := []string{
			"holis",
			"holus",
			"chau",
		}
		testLabels := model.LabelSet{
			"test": "watcher_read",
		}

		for _, line := range lines {
			res.writeEntry(loki.Entry{
				Labels: testLabels,
				Entry: logproto.Entry{
					Timestamp: time.Now(),
					Line:      line,
				},
			})
		}
		require.NoError(t, res.syncWAL())

		// notify watcher that entries have been written
		res.notifyWrite()

		require.Eventually(t, func() bool {
			return res.writeTo.ReadEntries.Length() == 3
		}, time.Second*10, time.Second, "expected watcher to catch up with written entries")
		defer res.writeTo.ReadEntries.DoneIterate()
		for _, readEntry := range res.writeTo.ReadEntries.StartIterate() {
			require.Contains(t, lines, readEntry.Line, "not expected log line")
		}
	},

	"read entries from WAL, just using backup timer to trigger reads": func(t *testing.T, res *watcherTestResources) {
		res.startWatcher()

		lines := []string{
			"holis",
			"holus",
			"chau",
		}
		testLabels := model.LabelSet{
			"test": "watcher_read",
		}

		for _, line := range lines {
			res.writeEntry(loki.Entry{
				Labels: testLabels,
				Entry: logproto.Entry{
					Timestamp: time.Now(),
					Line:      line,
				},
			})
		}
		require.NoError(t, res.syncWAL())

		// do not notify, let the backup timer trigger the watcher reads

		require.Eventually(t, func() bool {
			return res.writeTo.ReadEntries.Length() == 3
		}, time.Second*10, time.Second, "expected watcher to catch up with written entries")
		defer res.writeTo.ReadEntries.DoneIterate()
		for _, readEntry := range res.writeTo.ReadEntries.StartIterate() {
			require.Contains(t, lines, readEntry.Line, "not expected log line")
		}
	},

	"continue reading entries in next segment after initial segment is closed": func(t *testing.T, res *watcherTestResources) {
		res.startWatcher()
		lines := []string{
			"holis",
			"holus",
			"chau",
		}
		linesAfter := []string{
			"holis2",
			"holus2",
			"chau2",
		}
		testLabels := model.LabelSet{
			"test": "watcher_read",
		}

		for _, line := range lines {
			res.writeEntry(loki.Entry{
				Labels: testLabels,
				Entry: logproto.Entry{
					Timestamp: time.Now(),
					Line:      line,
				},
			})
		}
		require.NoError(t, res.syncWAL())

		res.notifyWrite()

		require.Eventually(t, func() bool {
			return res.writeTo.ReadEntries.Length() == 3
		}, time.Second*10, time.Second, "expected watcher to catch up with written entries")
		for _, readEntry := range res.writeTo.ReadEntries.StartIterate() {
			require.Contains(t, lines, readEntry.Line, "not expected log line")
		}
		res.writeTo.ReadEntries.DoneIterate()

		err := res.nextWALSegment()
		require.NoError(t, err, "expected no error when moving to next wal segment")

		for _, line := range linesAfter {
			res.writeEntry(loki.Entry{
				Labels: testLabels,
				Entry: logproto.Entry{
					Timestamp: time.Now(),
					Line:      line,
				},
			})
		}
		require.NoError(t, res.syncWAL())
		res.notifyWrite()

		require.Eventually(t, func() bool {
			return res.writeTo.ReadEntries.Length() == 6
		}, time.Second*10, time.Second, "expected watcher to catch up after new wal segment is cut")
		// assert over second half of entries
		defer res.writeTo.ReadEntries.DoneIterate()
		for _, readEntry := range res.writeTo.ReadEntries.StartIterate()[3:] {
			require.Contains(t, linesAfter, readEntry.Line, "not expected log line")
		}
	},

	"start reading from last segment": func(t *testing.T, res *watcherTestResources) {
		linesAfter := []string{
			"holis2",
			"holus2",
			"chau2",
		}
		testLabels := model.LabelSet{
			"test": "watcher_read",
		}

		// write something to first segment
		res.writeEntry(loki.Entry{
			Labels: testLabels,
			Entry: logproto.Entry{
				Timestamp: time.Now(),
				Line:      "this shouldn't be read",
			},
		})

		require.NoError(t, res.syncWAL())

		err := res.nextWALSegment()
		require.NoError(t, err, "expected no error when moving to next wal segment")

		res.startWatcher()

		for _, line := range linesAfter {
			res.writeEntry(loki.Entry{
				Labels: testLabels,
				Entry: logproto.Entry{
					Timestamp: time.Now(),
					Line:      line,
				},
			})
		}
		require.NoError(t, res.syncWAL())

		res.notifyWrite()

		require.Eventually(t, func() bool {
			return res.writeTo.ReadEntries.Length() == 3
		}, time.Second*10, time.Second, "expected watcher to catch up after new wal segment is cut")
		// assert over second half of entries
		defer res.writeTo.ReadEntries.DoneIterate()
		for _, readEntry := range res.writeTo.ReadEntries.StartIterate()[3:] {
			require.Contains(t, linesAfter, readEntry.Line, "not expected log line")
		}
	},

	"watcher receives segments reclaimed notifications correctly": func(t *testing.T, res *watcherTestResources) {
		res.startWatcher()
		testLabels := model.LabelSet{
			"test": "watcher_read",
		}

		writeAndWaitForWatcherToCatchUp := func(line string, expectedReadEntries int) {
			res.writeEntry(loki.Entry{
				Labels: testLabels,
				Entry: logproto.Entry{
					Timestamp: time.Now(),
					Line:      line,
				},
			})
			require.NoError(t, res.syncWAL())
			res.notifyWrite()
			require.Eventually(t, func() bool {
				return res.writeTo.ReadEntries.Length() == expectedReadEntries
			}, time.Second*10, time.Second, "expected watcher to catch up with written entries")
		}

		// writing segment 0
		writeAndWaitForWatcherToCatchUp("segment 0", 1)

		// moving to segment 1
		require.NoError(t, res.nextWALSegment(), "expected no error when moving to next wal segment")

		// moving on to segment 1
		writeAndWaitForWatcherToCatchUp("segment 1", 2)

		// collecting segment 0
		res.notifySegmentReclaimed(0)
		require.Eventually(t, func() bool {
			return len(res.writeTo.ReceivedSeriesReset) == 1 && res.writeTo.ReceivedSeriesReset[0] == 0
		}, time.Second*10, time.Second, "timed out waiting to receive series reset")

		// moving and writing to segment 2
		require.NoError(t, res.nextWALSegment(), "expected no error when moving to next wal segment")
		writeAndWaitForWatcherToCatchUp("segment 2", 3)
		time.Sleep(time.Millisecond)
		// moving and writing to segment 3
		require.NoError(t, res.nextWALSegment(), "expected no error when moving to next wal segment")
		writeAndWaitForWatcherToCatchUp("segment 3", 4)

		// collecting all segments up to 2
		res.notifySegmentReclaimed(2)
		// Expect second SeriesReset call to have the highest numbered deleted segment, 2
		require.Eventually(t, func() bool {
			t.Logf("received series reset: %v", res.writeTo.ReceivedSeriesReset)
			return len(res.writeTo.ReceivedSeriesReset) == 2 && res.writeTo.ReceivedSeriesReset[1] == 2
		}, time.Second*10, time.Second, "timed out waiting to receive series reset")
	},
}

type noMarker struct{}

func (n noMarker) LastMarkedSegment() int {
	return -1
}

// TestWatcher is the main test function, that works as framework to test different scenarios of the Watcher. It bootstraps
// necessary test components.
func TestWatcher(t *testing.T) {
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			// start test global resources
			reg := prometheus.NewRegistry()
			logger := level.NewFilter(log.NewLogfmtLogger(os.Stdout), level.AllowDebug())
			dir := t.TempDir()
			metrics := NewWatcherMetrics(reg)
			writeTo := &testWriteTo{
				series:      map[uint64]model.LabelSet{},
				logger:      logger,
				ReadEntries: utils.NewSyncSlice[loki.Entry](),
			}
			// create new watcher, and defer stop
			watcher := NewWatcher(dir, "test", metrics, writeTo, logger, DefaultWatchConfig, noMarker{})
			defer watcher.Stop()
			wl, err := New(Config{
				Enabled: true,
				Dir:     dir,
			}, logger, reg)
			require.NoError(t, err)
			defer wl.Close()
			ew := newEntryWriter()
			// run test case injecting resources
			testCase(
				t,
				&watcherTestResources{
					writeEntry: func(entry loki.Entry) {
						_ = ew.WriteEntry(entry, wl, logger)
					},
					notifyWrite: func() {
						watcher.NotifyWrite()
					},
					startWatcher: func() {
						watcher.Start()
					},
					syncWAL: func() error {
						return wl.Sync()
					},
					nextWALSegment: func() error {
						_, err := wl.NextSegment()
						return err
					},
					writeTo: writeTo,
					notifySegmentReclaimed: func(segmentNum int) {
						writeTo.SeriesReset(segmentNum)
					},
				},
			)
		})
	}
}

type mockMarker struct {
	LastMarkedSegmentFunc func() int
}

func (m mockMarker) LastMarkedSegment() int {
	return m.LastMarkedSegmentFunc()
}

func TestWatcher_Replay(t *testing.T) {
	labels := model.LabelSet{
		"app": "test",
	}
	segment1Lines := []string{
		"before 1",
		"before 2",
		"before 3",
	}
	segment2Lines := []string{
		"after 1",
		"after 2",
		"after 3",
	}

	t.Run("replay from marked segment if marker is not invalid", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		logger := level.NewFilter(log.NewLogfmtLogger(os.Stdout), level.AllowDebug())
		dir := t.TempDir()
		metrics := NewWatcherMetrics(reg)
		writeTo := &testWriteTo{
			series:      map[uint64]model.LabelSet{},
			logger:      logger,
			ReadEntries: utils.NewSyncSlice[loki.Entry](),
		}
		// create new watcher, and defer stop
		watcher := NewWatcher(dir, "test", metrics, writeTo, logger, DefaultWatchConfig, mockMarker{
			LastMarkedSegmentFunc: func() int {
				// when starting watcher, read from segment 0
				return 0
			},
		})
		defer watcher.Stop()
		wl, err := New(Config{
			Enabled: true,
			Dir:     dir,
		}, logger, reg)
		require.NoError(t, err)
		defer wl.Close()

		ew := newEntryWriter()

		// First, write to segment 0. This will be the last "marked" segment
		err = ew.WriteEntry(loki.Entry{
			Labels: labels,
			Entry: logproto.Entry{
				Timestamp: time.Now(),
				Line:      "this line should appear in received entries",
			},
		}, wl, logger)
		require.NoError(t, err)

		// cut segment and sync
		_, err = wl.NextSegment()
		require.NoError(t, err)

		// Now, write to segment 1, this will be a segment not marked, hence replayed
		for _, line := range segment1Lines {
			err = ew.WriteEntry(loki.Entry{
				Labels: labels,
				Entry: logproto.Entry{
					Timestamp: time.Now(),
					Line:      line,
				},
			}, wl, logger)
			require.NoError(t, err)
		}

		// cut segment and sync
		_, err = wl.NextSegment()
		require.NoError(t, err)

		// Finally, write some data to the last segment, this will be the write head
		for _, line := range segment2Lines {
			err = ew.WriteEntry(loki.Entry{
				Labels: labels,
				Entry: logproto.Entry{
					Timestamp: time.Now(),
					Line:      line,
				},
			}, wl, logger)
			require.NoError(t, err)
		}

		// sync wal, and start watcher
		require.NoError(t, wl.Sync())

		// start watcher
		watcher.Start()

		require.Eventually(t, func() bool {
			return writeTo.ReadEntries.Length() == 6 // wait for watcher to catch up with both segments
		}, time.Second*10, time.Second, "timed out waiting for watcher to catch up")
		writeTo.AssertContainsLines(t, segment1Lines...)
		writeTo.AssertContainsLines(t, segment2Lines...)
	})

	t.Run("do not replay at all if invalid marker", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		logger := level.NewFilter(log.NewLogfmtLogger(os.Stdout), level.AllowDebug())
		dir := t.TempDir()
		metrics := NewWatcherMetrics(reg)
		writeTo := &testWriteTo{
			series:      map[uint64]model.LabelSet{},
			logger:      logger,
			ReadEntries: utils.NewSyncSlice[loki.Entry](),
		}
		// create new watcher, and defer stop
		watcher := NewWatcher(dir, "test", metrics, writeTo, logger, DefaultWatchConfig, mockMarker{
			LastMarkedSegmentFunc: func() int {
				// when starting watcher, read from segment 0
				return -1
			},
		})
		defer watcher.Stop()
		wl, err := New(Config{
			Enabled: true,
			Dir:     dir,
		}, logger, reg)
		require.NoError(t, err)
		defer wl.Close()

		ew := newEntryWriter()

		// First, write to segment 0. This will be the last "marked" segment
		err = ew.WriteEntry(loki.Entry{
			Labels: labels,
			Entry: logproto.Entry{
				Timestamp: time.Now(),
				Line:      "this line should appear in received entries",
			},
		}, wl, logger)
		require.NoError(t, err)

		// cut segment and sync
		_, err = wl.NextSegment()
		require.NoError(t, err)

		// Now, write to segment 1, this will be a segment not marked, hence replayed
		for _, line := range segment1Lines {
			err = ew.WriteEntry(loki.Entry{
				Labels: labels,
				Entry: logproto.Entry{
					Timestamp: time.Now(),
					Line:      line,
				},
			}, wl, logger)
			require.NoError(t, err)
		}

		// cut segment and sync
		_, err = wl.NextSegment()
		require.NoError(t, err)

		// sync wal, and start watcher
		require.NoError(t, wl.Sync())

		// start watcher
		watcher.Start()

		// Write something after watcher started
		for _, line := range segment2Lines {
			err = ew.WriteEntry(loki.Entry{
				Labels: labels,
				Entry: logproto.Entry{
					Timestamp: time.Now(),
					Line:      line,
				},
			}, wl, logger)
			require.NoError(t, err)
		}

		// sync wal, and start watcher
		require.NoError(t, wl.Sync())

		require.Eventually(t, func() bool {
			return writeTo.ReadEntries.Length() == 3 // wait for watcher to catch up with both segments
		}, time.Second*10, time.Second, "timed out waiting for watcher to catch up")
		writeTo.AssertContainsLines(t, segment2Lines...)
	})
}
