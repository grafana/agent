package wal

import (
	"sync"
	"unsafe"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb"
	tsdbLabels "github.com/prometheus/prometheus/tsdb/labels"
	"github.com/prometheus/prometheus/tsdb/wal"
)

// TODO
// - Regularly checkpoint to a separate directory just like Prometheus,
//   so the remote write code picks up checkpoints and garbage collects series.

// Storage implements storage.Storage, and just writes to the WAL.
type Storage struct {
	// Embed Queryable for compatibility, but don't actually implement it.
	storage.Queryable

	wal *wal.WAL

	appenderPool sync.Pool
	bufPool      sync.Pool

	mtx     sync.RWMutex
	labels  map[string]uint64
	nextRef uint64
}

// NewStorage makes a new Storage.
func NewStorage(logger log.Logger, registerer prometheus.Registerer, path string) (*Storage, error) {
	w, err := wal.Open(logger, registerer, path)
	if err != nil {
		return nil, err
	}

	storage := &Storage{
		labels: map[string]uint64{},
		wal:    w,
	}

	storage.bufPool.New = func() interface{} {
		return make([]byte, 0, 1024)
	}

	storage.appenderPool.New = func() interface{} {
		return &appender{
			w:       storage,
			series:  make([]tsdb.RefSeries, 0, 100),
			samples: make([]tsdb.RefSample, 0, 100),
		}
	}

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

// Close closes the storage and all its underlying resources.
func (w *Storage) Close() error {
	return w.wal.Close()
}

type appender struct {
	w       *Storage
	series  []tsdb.RefSeries
	samples []tsdb.RefSample
}

func (a *appender) Add(l labels.Labels, t int64, v float64) (uint64, error) {
	ref, addSeries := a.w.lookupLabels(l)

	if addSeries {
		a.series = append(a.series, tsdb.RefSeries{
			Ref:    ref,
			Labels: toTSDBLabels(l),
		})
	}

	a.samples = append(a.samples, tsdb.RefSample{
		Ref: ref,
		T:   t,
		V:   v,
	})

	return ref, nil
}

func toTSDBLabels(l labels.Labels) tsdbLabels.Labels {
	return *(*tsdbLabels.Labels)(unsafe.Pointer(&l))
}

func (a *appender) AddFast(_ labels.Labels, ref uint64, t int64, v float64) error {
	a.samples = append(a.samples, tsdb.RefSample{
		Ref: ref,
		T:   t,
		V:   v,
	})
	return nil
}

// Commit submits the collected samples and purges the batch.
func (a *appender) Commit() error {
	var encoder tsdb.RecordEncoder
	buf := a.w.bufPool.Get().([]byte)
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
	a.w.bufPool.Put(buf)
	return a.Rollback()
}

func (a *appender) Rollback() error {
	a.series = a.series[:0]
	a.samples = a.samples[:0]
	a.w.appenderPool.Put(a)
	return nil
}
