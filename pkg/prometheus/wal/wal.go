package wal

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb/record"
	"github.com/prometheus/prometheus/tsdb/wal"
)

// TODO(rfratto):
// - Track active/deleted series for WAL checkpointing
// - Use some kind of tooling to read from the WAL to test and validate that
//   everything we're doing so far is being done correctly

// Storage implements storage.Storage, and just writes to the WAL.
type Storage struct {
	// Embed Queryable for compatibility, but don't actually implement it.
	storage.Queryable

	wal    *wal.WAL
	logger log.Logger

	appenderPool sync.Pool
	bufPool      sync.Pool

	mtx     sync.RWMutex
	labels  map[string]uint64
	nextRef uint64
}

// NewStorage makes a new Storage.
func NewStorage(logger log.Logger, registerer prometheus.Registerer, path string) (*Storage, error) {
	w, err := wal.NewSize(logger, registerer, filepath.Join(path, "wal"), wal.DefaultSegmentSize, true)
	if err != nil {
		return nil, err
	}

	storage := &Storage{
		labels: map[string]uint64{},
		wal:    w,
		logger: logger,
	}

	storage.bufPool.New = func() interface{} {
		// staticcheck wants slices in a sync.Pool to be pointers to
		// avoid overhead of allocating a struct with the length, capacity, and
		// pointer to underlying array.
		b := make([]byte, 0, 1024)
		return &b
	}

	storage.appenderPool.New = func() interface{} {
		return &appender{
			w:       storage,
			series:  make([]record.RefSeries, 0, 100),
			samples: make([]record.RefSample, 0, 100),
		}
	}

	// TODO(rfratto): we need to replay the WAL from the most recent checkpoint
	// and all segments after the checkpoint so we can track active series.
	//
	// A series becomes inactive once it hasn't been written to since the time
	// that is being truncated.

	return storage, nil
}

func (w *Storage) lookupLabels(l labels.Labels) (uint64, bool) {
	s := l.String()

	w.mtx.RLock()
	ref, ok := w.labels[s]
	w.mtx.RUnlock()

	if ok {
		return ref, false
	}

	w.mtx.Lock()
	ref, ok = w.labels[s]
	if ok {
		w.mtx.Unlock()
		return ref, false
	}

	ref = w.nextRef
	w.nextRef++
	w.labels[s] = ref
	w.mtx.Unlock()
	return ref, true
}

// StartTime returns the oldest timestamp stored in the storage.
func (*Storage) StartTime() (int64, error) {
	return 0, nil
}

// Appender returns a new appender against the storage.
func (w *Storage) Appender() (storage.Appender, error) {
	return w.appenderPool.Get().(storage.Appender), nil
}

// Truncate removes all data from the WAL prior to the timestamp specified by
// mint.
func (w *Storage) Truncate(mint int64) error {
	start := time.Now()

	// TODO(rfratto): garbage collect series that haven't
	// received an update since the last truncation

	first, last, err := w.wal.Segments()
	if err != nil {
		return errors.Wrap(err, "get segment range")
	}

	// Start a new segment, so low ingestion volume instance don't have more WAL
	// than needed.
	err = w.wal.NextSegment()
	if err != nil {
		return errors.Wrap(err, "next segment")
	}

	last-- // Never consider last segment for checkpoint.
	if last < 0 {
		return nil // no segments yet.
	}

	// The lower third of segments should contain mostly obsolete samples.
	// If we have less than three segments, it's not worth checkpointing yet.
	last = first + (last-first)/3
	if last <= first {
		return nil
	}

	keep := func(id uint64) bool {
		// TODO(rfratto): check to see if the series should be kept:
		// if it's still receiving writes, keep it. If it's been deleted
		// in the most recent GC cycle, keep it.
		return true
	}
	if _, err = wal.Checkpoint(w.wal, first, last, keep, mint); err != nil {
		return errors.Wrap(err, "create checkpoint")
	}
	if err := w.wal.Truncate(last + 1); err != nil {
		// If truncating fails, we'll just try again at the next checkpoint.
		// Leftover segments will just be ignored in the future if there's a checkpoint
		// that supersedes them.
		level.Error(w.logger).Log("msg", "truncating segments failed", "err", err)
	}

	// TODO(rfratto): now that the checkpoint is written and all segments before
	// it have been truncated, we can stop tracking deleted series. For all series
	// that were deleted before our first truncated segment, we can stop tracking
	// them.

	if err := wal.DeleteCheckpoints(w.wal.Dir(), last); err != nil {
		// Leftover old checkpoints do not cause problems down the line beyond
		// occupying disk space.
		// They will just be ignored since a higher checkpoint exists.
		level.Error(w.logger).Log("msg", "delete old checkpoints", "err", err)
	}

	level.Info(w.logger).Log("msg", "WAL checkpoint complete",
		"first", first, "last", last, "duration", time.Since(start))
	return nil
}

// Close closes the storage and all its underlying resources.
func (w *Storage) Close() error {
	return w.wal.Close()
}

type appender struct {
	w       *Storage
	series  []record.RefSeries
	samples []record.RefSample
}

func (a *appender) Add(l labels.Labels, t int64, v float64) (uint64, error) {
	ref, addSeries := a.w.lookupLabels(l)

	if addSeries {
		a.series = append(a.series, record.RefSeries{
			Ref:    ref,
			Labels: l,
		})
	}

	a.samples = append(a.samples, record.RefSample{
		Ref: ref,
		T:   t,
		V:   v,
	})

	return ref, nil
}

func (a *appender) AddFast(_ labels.Labels, ref uint64, t int64, v float64) error {
	a.samples = append(a.samples, record.RefSample{
		Ref: ref,
		T:   t,
		V:   v,
	})
	return nil
}

// Commit submits the collected samples and purges the batch.
func (a *appender) Commit() error {
	var encoder record.Encoder
	bufp := a.w.bufPool.Get().(*[]byte)
	buf := *bufp

	buf = encoder.Series(a.series, buf)
	if err := a.w.wal.Log(buf); err != nil {
		return err
	}

	buf = buf[:0]
	buf = encoder.Samples(a.samples, buf)
	if err := a.w.wal.Log(buf); err != nil {
		return err
	}

	buf = buf[:0]
	a.w.bufPool.Put(&buf)
	return a.Rollback()
}

func (a *appender) Rollback() error {
	a.series = a.series[:0]
	a.samples = a.samples[:0]
	a.w.appenderPool.Put(a)
	return nil
}
