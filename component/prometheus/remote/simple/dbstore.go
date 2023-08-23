package simple

import (
	"context"
	"path"
	"sync"
	"time"

	"github.com/grafana/agent/component/prometheus/remote/simple/pebble"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/prometheus/client_golang/prometheus"
)

type dbstore struct {
	mut           sync.RWMutex
	l             *logging.Logger
	dir           string
	ttl           time.Duration
	inMemory      bool
	sampleDB      SignalDB
	bookmark      SignalDB
	ctx           context.Context
	metrics       *dbmetrics
	oldestInUsKey uint64
}

const (
	MetricSignal int8 = iota
	HistogramSignal
	FloathistogramSignal
	MetadataSignal
	ExemplarSignal
	BookmarkType
)

func newDBStore(inMemory bool, ttl time.Duration, directory string, r prometheus.Registerer, l *logging.Logger) (*dbstore, error) {
	// bookmarkSize := 1024 * 1024 * 1 // 1MB
	// bookmark, err := badger.NewDB(path.Join(directory, "bookmark"), int64(bookmarkSize), l, GetValue, GetType)
	bookmark, err := pebble.NewDB(path.Join(directory, "bookmark"), GetValue, GetType, l)
	if err != nil {
		return nil, err
	}
	// sampleSize := 1024 * 1024 * 256 // 256MB
	// sample, err := badger.NewDB(path.Join(directory, "signals"), int64(sampleSize), l, GetValue, GetType)
	sample, err := pebble.NewDB(path.Join(directory, "sample"), GetValue, GetType, l)
	if err != nil {
		return nil, err
	}
	store := &dbstore{
		inMemory: inMemory,
		bookmark: bookmark,
		sampleDB: sample,
		ttl:      ttl,
		l:        l,
	}

	dbm := newDbMetrics(r, store)
	store.metrics = dbm

	return store, nil
}

func (dbs *dbstore) Run(ctx context.Context) {
	dbs.ctx = ctx
	// Evict on startup to clean up any TTL files.
	dbs.evict()
}

func (dbs *dbstore) evict() {
	dbs.mut.Lock()
	defer dbs.mut.Unlock()

	start := time.Now()
	defer dbs.metrics.evictionTime.Observe(time.Now().Sub(start).Seconds())
	dbs.bookmark.Evict()
	dbs.sampleDB.Evict()
}

func (dbs *dbstore) WriteBookmark(key string, value any) error {
	return dbs.bookmark.WriteValue([]byte(key), value, 0*time.Second)
}

func (dbs *dbstore) GetBookmark(key string) (*Bookmark, bool) {
	bk, found, _ := dbs.bookmark.GetValueByString(key)
	if bk == nil {
		return &Bookmark{Key: 1}, false
	}
	return bk.(*Bookmark), found
}

func (dbs *dbstore) WriteSignal(value any) (uint64, error) {
	start := time.Now()
	defer dbs.metrics.writeTime.Observe(float64(time.Now().Sub(start).Seconds()))

	key, err := dbs.sampleDB.WriteValueWithAutokey(value, dbs.ttl)
	dbs.metrics.currentKey.Set(float64(key))
	level.Debug(dbs.l).Log("msg", "writing signals to WAL", "key", key)
	return key, err
}

func (dbs *dbstore) GetOldestKey() uint64 {
	return dbs.sampleDB.GetOldestKey()
}

func (dbs *dbstore) GetNextKey(k uint64) uint64 {
	return dbs.sampleDB.GetNextKey(k)
}

func (dbs *dbstore) UpdateOldestKey(k uint64) {
	dbs.mut.Lock()
	defer dbs.mut.Unlock()
	dbs.oldestInUsKey = k
}

func (dbs *dbstore) GetSignal(key uint64) (any, bool) {
	start := time.Now()
	defer dbs.metrics.readTime.Observe(float64(time.Now().Sub(start).Seconds()))

	val, found, err := dbs.sampleDB.GetValueByUint(key)
	if err != nil {
		level.Error(dbs.l).Log("error finding key", err, "key", key)
		return nil, false
	}
	return val, found
}

func (dbs *dbstore) getKeyCount() uint64 {
	keys, _ := dbs.sampleDB.GetKeys()
	return uint64(len(keys))
}

func (dbs *dbstore) getFileSize() float64 {
	return float64(dbs.sampleDB.Size() + dbs.bookmark.Size())
}

func (dbs *dbstore) sampleCount() float64 {
	return float64(dbs.sampleDB.SeriesCount())
}
