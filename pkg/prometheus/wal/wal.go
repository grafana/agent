package wal

import (
	"fmt"
	"math"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/timestamp"
	"github.com/prometheus/prometheus/pkg/value"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb/record"
	"github.com/prometheus/prometheus/tsdb/wal"
)

// Storage implements storage.Storage, and just writes to the WAL.
type Storage struct {
	// Embed Queryable for compatibility, but don't actually implement it.
	storage.Queryable

	wal    *wal.WAL
	logger log.Logger

	appenderPool sync.Pool
	bufPool      sync.Pool

	mtx     sync.RWMutex
	nextRef uint64
	series  *stripeSeries

	deletedMtx sync.Mutex
	deleted    map[uint64]int // Deleted series, and what WAL segment they must be kept until.
}

// NewStorage makes a new Storage.
func NewStorage(logger log.Logger, registerer prometheus.Registerer, path string) (*Storage, error) {
	w, err := wal.NewSize(logger, registerer, filepath.Join(path, "wal"), wal.DefaultSegmentSize, true)
	if err != nil {
		return nil, err
	}

	storage := &Storage{
		wal:     w,
		logger:  logger,
		deleted: map[uint64]int{},
		series:  newStripeSeries(),
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

	if err := storage.replayWAL(); err != nil {
		level.Warn(storage.logger).Log("msg", "encountered WAL read error, attempting repair", "err", err)
		if err := w.Repair(err); err != nil {
			return nil, errors.Wrap(err, "repair corrupted WAL")
		}
	}

	return storage, nil
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

	// Garbage collect series that haven't received an update since mint.
	w.gc(mint)
	level.Info(w.logger).Log("msg", "series GC completed", "duration", time.Since(start))

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
		if w.series.getByID(id) != nil {
			return true
		}

		w.deletedMtx.Lock()
		_, ok := w.deleted[id]
		w.deletedMtx.Unlock()
		return ok
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

	// The checkpoint is written and segments before it is truncated, so we no
	// longer need to track deleted series that are before it.
	w.deletedMtx.Lock()
	for ref, segment := range w.deleted {
		if segment < first {
			delete(w.deleted, ref)
		}
	}
	w.deletedMtx.Unlock()

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

// WriteStalenessMarkers appends a staleness sample for all active series.
func (w *Storage) WriteStalenessMarkers(remoteTsFunc func() int64) error {
	var lastErr error
	var lastTs int64

	s := w.series

	app, err := w.Appender()
	if err != nil {
		return err
	}

	// TODO(rfratto): write a series iterator to refactor this
	for i := 0; i < s.size; i++ {
		s.locks[i].RLock()

		for _, all := range s.hashes[i] {
			for _, series := range all {
				series.Lock()

				j := int(series.ref) & (s.size - 1)

				if i != j {
					s.locks[j].RLock()
				}

				// Get values from the series before unlocking it
				var (
					labels = series.lset
					ref    = series.ref
				)

				if i != j {
					s.locks[j].RUnlock()
				}

				series.Unlock()

				ts := timestamp.FromTime(time.Now())
				err = app.AddFast(labels, ref, ts, math.Float64frombits(value.StaleNaN))
				if err != nil {
					lastErr = err
				}

				// Remove millisecond precision; the remote write timestamp we get
				// only has second precision.
				lastTs = (ts / 1000) * 1000
			}
		}

		s.locks[i].RUnlock()
	}

	if lastErr == nil {
		if err := app.Commit(); err != nil {
			return fmt.Errorf("failed to commit staleness markers: %w", err)
		}

		// Wait for remote write to write the lastTs, but give up after 1m
		level.Info(w.logger).Log("msg", "waiting for remote write to write staleness markers...")

		stopCh := time.After(1 * time.Minute)
		start := time.Now()

	Outer:
		for {
			select {
			case <-stopCh:
				level.Error(w.logger).Log("msg", "timed out waiting for staleness markers to be written")
				break Outer
			default:
				writtenTs := remoteTsFunc()
				if writtenTs >= lastTs {
					duration := time.Since(start)
					level.Info(w.logger).Log("msg", "remote write wrote staleness markers", "duration", duration)
					break Outer
				}

				level.Info(w.logger).Log("msg", "remote write hasn't written staleness markers yet", "remoteTs", writtenTs, "lastTs", lastTs)

				// Wait a bit before reading again
				time.Sleep(5 * time.Second)
			}
		}
	}

	return lastErr
}

// gc removes data before the minimum timestamp from the head.
func (w *Storage) gc(mint int64) {
	deleted := w.series.gc(mint)

	_, last, _ := w.wal.Segments()
	w.deletedMtx.Lock()
	defer w.deletedMtx.Unlock()

	// We want to keep series records for any newly deleted series
	// until we've passed the last recorded segment. The WAL will
	// still contain samples records with all of the ref IDs until
	// the segment's samples has been deleted from the checkpoint.
	//
	// If the series weren't kept on startup when the WAL was replied,
	// the samples wouldn't be able to be used since there wouldn't
	// be any labels for that ref ID.
	for ref := range deleted {
		w.deleted[ref] = last
	}
}

func (w *Storage) replayWAL() error {
	level.Info(w.logger).Log("msg", "replaying WAL, this may take awhile", "dir", w.wal.Dir())
	dir, startFrom, err := wal.LastCheckpoint(w.wal.Dir())
	if err != nil && err != record.ErrNotFound {
		return errors.Wrap(err, "find last checkpoint")
	}

	if err == nil {
		sr, err := wal.NewSegmentsReader(dir)
		if err != nil {
			return errors.Wrap(err, "open checkpoint")
		}
		defer func() {
			if err := sr.Close(); err != nil {
				level.Warn(w.logger).Log("msg", "error while closing the wal segments reader", "err", err)
			}
		}()

		// A corrupted checkpoint is a hard error for now and requires user
		// intervention. There's likely little data that can be recovered anyway.
		if err := w.loadWAL(wal.NewReader(sr)); err != nil {
			return errors.Wrap(err, "backfill checkpoint")
		}
		startFrom++
		level.Info(w.logger).Log("msg", "WAL checkpoint loaded")
	}

	// Find the last segment.
	_, last, err := w.wal.Segments()
	if err != nil {
		return errors.Wrap(err, "finding WAL segments")
	}

	// Backfill segments from the most recent checkpoint onwards.
	for i := startFrom; i <= last; i++ {
		s, err := wal.OpenReadSegment(wal.SegmentName(w.wal.Dir(), i))
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("open WAL segment: %d", i))
		}

		sr := wal.NewSegmentBufReader(s)
		err = w.loadWAL(wal.NewReader(sr))
		if err := sr.Close(); err != nil {
			level.Warn(w.logger).Log("msg", "error while closing the wal segments reader", "err", err)
		}
		if err != nil {
			return err
		}
		level.Info(w.logger).Log("msg", "WAL segment loaded", "segment", i, "maxSegment", last)
	}

	return nil
}

func (w *Storage) loadWAL(r *wal.Reader) (err error) {
	var (
		dec record.Decoder
	)

	var (
		decoded    = make(chan interface{}, 10)
		errCh      = make(chan error, 1)
		seriesPool = sync.Pool{
			New: func() interface{} {
				return []record.RefSeries{}
			},
		}
	)

	go func() {
		defer close(decoded)
		for r.Next() {
			rec := r.Record()
			switch dec.Type(rec) {
			case record.Series:
				series := seriesPool.Get().([]record.RefSeries)[:0]
				series, err = dec.Series(rec, series)
				if err != nil {
					errCh <- &wal.CorruptionErr{
						Err:     errors.Wrap(err, "decode series"),
						Segment: r.Segment(),
						Offset:  r.Offset(),
					}
					return
				}
				decoded <- series
			case record.Samples:
				// We don't care about samples
				continue
			case record.Tombstones:
				// We don't care about tombstones
				continue
			default:
				errCh <- &wal.CorruptionErr{
					Err:     errors.Errorf("invalid record type %v", dec.Type(rec)),
					Segment: r.Segment(),
					Offset:  r.Offset(),
				}
				return
			}
		}
	}()

	for d := range decoded {
		switch v := d.(type) {
		case []record.RefSeries:
			for _, s := range v {
				// Create the series in memory with the time it was read. This
				// guarantees it will exist for at least two Truncation cycles
				// in case it gets appended to late.
				//
				// TODO(rfratto): we should expect some samples from the series to
				// exist. Iterate over the samples and use the TS as lastTs instead
				// so churned series don't stick around longer than they need to.
				ts := timestamp.FromTime(time.Now())
				series := &memSeries{ref: s.Ref, lastTs: ts, lset: s.Labels}
				w.series.set(s.Labels.Hash(), series)

				w.mtx.Lock()
				if w.nextRef <= s.Ref {
					w.nextRef = s.Ref + 1
				}
				w.mtx.Unlock()
			}

			//nolint:staticcheck
			seriesPool.Put(v)
		default:
			panic(fmt.Errorf("unexpected decoded type: %T", d))
		}
	}

	select {
	case err := <-errCh:
		return err
	default:
	}

	if r.Err() != nil {
		return errors.Wrap(r.Err(), "read records")
	}

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
	var (
		series *memSeries

		hash = l.Hash()
	)

	series = a.w.series.getByHash(hash, l)
	if series == nil {
		a.w.mtx.Lock()
		ref := a.w.nextRef
		level.Debug(a.w.logger).Log("msg", "new series", "ref", ref, "labels", l.String())
		a.w.nextRef++
		a.w.mtx.Unlock()

		series = &memSeries{ref: ref, lset: l, lastTs: t}
		a.w.series.set(hash, series)

		a.series = append(a.series, record.RefSeries{
			Ref:    ref,
			Labels: l,
		})
	}

	return series.ref, a.AddFast(l, series.ref, t, v)
}

func (a *appender) AddFast(_ labels.Labels, ref uint64, t int64, v float64) error {
	series := a.w.series.getByID(ref)
	if series == nil {
		return storage.ErrNotFound
	}
	series.Lock()
	defer series.Unlock()

	// Update last recorded timestamp. Used by Storage.gc to determine if a
	// series is dead.
	series.lastTs = t

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
