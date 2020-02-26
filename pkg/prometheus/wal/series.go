package wal

import (
	"sync"

	"github.com/prometheus/prometheus/pkg/labels"
)

type memSeries struct {
	sync.Mutex

	ref    uint64
	lset   labels.Labels
	lastTs int64

	// TODO(rfratto): this solution below isn't perfect, and there's still
	// the possibility for a series to be deleted before it's
	// completely gone from the WAL. Rather, we should have gc return
	// a "should delete" map and be given a "deleted" map.
	// If a series that is going to be marked for deletion is in the
	// "deleted" map, then it should be deleted instead.
	//
	// The "deleted" map will be populated by the Truncate function.
	// It will be cleared with every call to gc.

	// willDelete marks a series as to be deleted on the next garbage
	// collection. If it receives a write, willDelete is disabled.
	willDelete bool
}

func (s *memSeries) updateTs(ts int64) {
	s.lastTs = ts
	s.willDelete = false
}

// seriesHashmap is a simple hashmap for memSeries by their label set. It is
// built on top of a regular hashmap and holds a slice of series to resolve
// hash collisions. Its methods require the hash to be submitted with it to
// avoid re-computations throughout the code.
//
// This code is copied from the Prometheus TSDB.
type seriesHashmap map[uint64][]*memSeries

func (m seriesHashmap) get(hash uint64, lset labels.Labels) *memSeries {
	for _, s := range m[hash] {
		if labels.Equal(s.lset, lset) {
			return s
		}
	}
	return nil
}

func (m seriesHashmap) set(hash uint64, s *memSeries) {
	l := m[hash]
	for i, prev := range l {
		if labels.Equal(prev.lset, s.lset) {
			l[i] = s
			return
		}
	}
	m[hash] = append(l, s)
}

func (m seriesHashmap) del(hash uint64, lset labels.Labels) {
	var rem []*memSeries
	for _, s := range m[hash] {
		if !labels.Equal(s.lset, lset) {
			rem = append(rem, s)
		}
	}
	if len(rem) == 0 {
		delete(m, hash)
	} else {
		m[hash] = rem
	}
}

const (
	// defaultStripeSize is the default number of entries to allocate in the
	// stripeSeries hash map.
	defaultStripeSize = 1 << 14
)

// stripeSeries locks modulo ranges of IDs and hashes to reduce lock contention.
// The locks are padded to not be on the same cache line. Filling the padded space
// with the maps was profiled to be slower – likely due to the additional pointer
// dereferences.
//
// This code is copied from the Prometheus TSDB.
type stripeSeries struct {
	size   int
	series []map[uint64]*memSeries
	hashes []seriesHashmap
	locks  []stripeLock
}

type stripeLock struct {
	sync.RWMutex
	// Padding to avoid multiple locks being on the same cache line.
	_ [40]byte
}

func newStripeSeries() *stripeSeries {
	stripeSize := defaultStripeSize
	s := &stripeSeries{
		size:   stripeSize,
		series: make([]map[uint64]*memSeries, stripeSize),
		hashes: make([]seriesHashmap, stripeSize),
		locks:  make([]stripeLock, stripeSize),
	}

	for i := range s.series {
		s.series[i] = map[uint64]*memSeries{}
	}
	for i := range s.hashes {
		s.hashes[i] = seriesHashmap{}
	}
	return s
}

// gc garbage collects old chunks that are strictly before mint and removes
// series entirely that have no chunks left.
func (s *stripeSeries) gc(mint int64) map[uint64]struct{} {
	var (
		deleted = map[uint64]struct{}{}
	)

	// Run through all series and find series that haven't been written to
	// since mint. Mark those series as deleted and store their ID.
	for i := 0; i < s.size; i++ {
		s.locks[i].Lock()

		for hash, all := range s.hashes[i] {
			for _, series := range all {
				series.Lock()

				// If the series has received a write after mint, there's still
				// data and it's not gone yet.
				if series.lastTs >= mint {
					series.willDelete = false
					series.Unlock()
					continue
				}

				// The series hasn't received any data and might be gone, but we
				// want to give it an opportunity to survive one more cycle before
				// marking it as deleted.
				if !series.willDelete {
					series.willDelete = true
					series.Unlock()
					continue
				}

				// The series is gone entirely. We need to keep the series lock
				// and make sure we have acquired the stripe locks for hash and ID of the
				// series alike.
				// If we don't hold them all, there's a very small chance that a series receives
				// samples again while we are half-way into deleting it.
				j := int(series.ref) & (s.size - 1)

				if i != j {
					s.locks[j].Lock()
				}

				deleted[series.ref] = struct{}{}
				s.hashes[i].del(hash, series.lset)
				releaseLabels(series.lset)
				delete(s.series[j], series.ref)

				if i != j {
					s.locks[j].Unlock()
				}

				series.Unlock()
			}
		}

		s.locks[i].Unlock()
	}

	return deleted
}

func (s *stripeSeries) getByID(id uint64) *memSeries {
	i := id & uint64(s.size-1)

	s.locks[i].RLock()
	series := s.series[i][id]
	s.locks[i].RUnlock()

	return series
}

func (s *stripeSeries) getByHash(hash uint64, lset labels.Labels) *memSeries {
	i := hash & uint64(s.size-1)

	s.locks[i].RLock()
	series := s.hashes[i].get(hash, lset)
	s.locks[i].RUnlock()

	return series
}

func (s *stripeSeries) set(hash uint64, series *memSeries) {
	i := hash & uint64(s.size-1)

	s.locks[i].Lock()

	s.hashes[i].set(hash, series)
	s.locks[i].Unlock()

	i = series.ref & uint64(s.size-1)

	s.locks[i].Lock()
	s.series[i][series.ref] = series
	s.locks[i].Unlock()
}

func (s *stripeSeries) iterator() *stripeSeriesIterator {
	return &stripeSeriesIterator{s}
}

// stripeSeriesIterator allows to iterate over series through a channel.
// The channel should always be completely consumed to not leak.
type stripeSeriesIterator struct {
	s *stripeSeries
}

func (it *stripeSeriesIterator) Channel() <-chan *memSeries {
	ret := make(chan *memSeries)

	go func() {
		for i := 0; i < it.s.size; i++ {
			it.s.locks[i].RLock()

			for _, all := range it.s.hashes[i] {
				for _, series := range all {
					series.Lock()

					j := int(series.ref) & (it.s.size - 1)
					if i != j {
						it.s.locks[j].RLock()
					}

					ret <- series

					if i != j {
						it.s.locks[j].RUnlock()
					}

					series.Unlock()
				}
			}

			it.s.locks[i].RUnlock()
		}

		close(ret)
	}()

	return ret
}
