package simple

import (
	"bytes"
	"context"
	"encoding/gob"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-kit/log"

	"github.com/grafana/agent/component/prometheus/remote/simple/pebble"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// dbstore is a helper interface around the bookmark and sample stores.
type dbstore struct {
	mut            sync.RWMutex
	directory      string
	l              log.Logger
	ttl            time.Duration
	sampleDB       *pebble.DB
	bookmark       *pebble.DB
	ctx            context.Context
	metrics        *dbmetrics
	oldestInUseKey uint64
	bookmarkPool   sync.Pool
}

func newDBStore(ttl time.Duration, directory string, r prometheus.Registerer, l log.Logger) (*dbstore, error) {
	bookmark, err := pebble.NewDB(path.Join(directory, "bookmark"), l)
	if err != nil {
		return nil, err
	}
	sample, err := pebble.NewDB(path.Join(directory, "sample"), l)
	if err != nil {
		return nil, err
	}
	store := &dbstore{
		bookmark:  bookmark,
		sampleDB:  sample,
		ttl:       ttl,
		l:         l,
		directory: directory,
	}
	store.bookmarkPool.New = func() any {
		return make([]byte, 0, 1024*1024)
	}

	dbm := newDbMetrics(r, store)
	store.metrics = dbm

	return store, nil
}

func (dbs *dbstore) Run(ctx context.Context) {
	dbs.ctx = ctx
	// Evict on startup to clean up any TTL files.
	dbs.evict()
	<-ctx.Done()
}

// WriteBookmark writes a bookmark for Writer.
func (dbs *dbstore) WriteBookmark(key string, value *Bookmark) error {
	tempBuf := dbs.bookmarkPool.Get().([]byte)
	defer dbs.bookmarkPool.Put(tempBuf)
	buf := bytes.NewBuffer(tempBuf)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(value)
	if err != nil {
		return err
	}

	return dbs.bookmark.WriteValue([]byte(key), buf.Bytes(), 0, 1, 0*time.Second)
}

// GetBookmark returns the bookmark for a given write name.
func (dbs *dbstore) GetBookmark(key string) (*Bookmark, bool) {
	bk, _, found, _ := dbs.bookmark.GetValueByString(key)
	if bk == nil {
		return &Bookmark{Key: 1}, false
	}
	buf := bytes.NewBuffer(bk)
	dec := gob.NewDecoder(buf)
	book := &Bookmark{}
	err := dec.Decode(book)
	if err != nil {
		return nil, false
	}

	return book, found
}

// WriteSignal writes a signal and applies an autokey.
func (dbs *dbstore) WriteSignal(value []byte, valType int8, count int) (uint64, error) {
	start := time.Now()
	defer dbs.metrics.writeTime.Observe(time.Since(start).Seconds())

	key, err := dbs.sampleDB.WriteValueWithAutokey(value, valType, count, dbs.ttl)
	dbs.metrics.currentKey.Set(float64(key))
	level.Debug(dbs.l).Log("msg", "writing signals to WAL", "key", key)
	return key, err
}

// GetOldestKey returns the oldest key in use.
func (dbs *dbstore) GetOldestKey() uint64 {
	return dbs.sampleDB.GetOldestKey()
}

// GetNextKey returns the next key that is in use, returns k if no newer items found.
func (dbs *dbstore) GetNextKey(k uint64) uint64 {
	return dbs.sampleDB.GetNextKey(k)
}

// UpdateOldestKey updates the oldest key in use to k.
func (dbs *dbstore) UpdateOldestKey(k uint64) {
	dbs.mut.Lock()
	defer dbs.mut.Unlock()

	dbs.oldestInUseKey = k
}

// GetSignal returns the value and whether it was found.
func (dbs *dbstore) GetSignal(key uint64) ([]byte, int8, bool) {
	start := time.Now()
	defer dbs.metrics.readTime.Observe(time.Since(start).Seconds())

	val, valType, found, err := dbs.sampleDB.GetValueByKey(key)
	if err != nil {
		level.Error(dbs.l).Log("error finding key", err, "key", key)
		return nil, -1, false
	}
	return val, valType, found
}

func (dbs *dbstore) getKeyCount() uint64 {
	keys, _ := dbs.sampleDB.GetKeys()
	return uint64(len(keys))
}

func (dbs *dbstore) getFileSize() float64 {
	return DirSize(dbs.directory)
}

func (dbs *dbstore) sampleCount() float64 {
	return float64(dbs.sampleDB.SeriesCount())
}

func (dbs *dbstore) averageCompressionRatio() float64 {
	return dbs.sampleDB.AverageCompressionRatio()
}

func (dbs *dbstore) evict() {
	dbs.mut.Lock()
	defer dbs.mut.Unlock()

	start := time.Now()
	defer dbs.metrics.evictionTime.Observe(time.Since(start).Seconds())
	err := dbs.bookmark.Evict()
	if err != nil {
		level.Error(dbs.l).Log("msg", "failure evicting bookmark db", "err", err)
	}
	err = dbs.sampleDB.Evict()
	if err != nil {
		level.Error(dbs.l).Log("msg", "failure evicting sample db", "err", err)
	}
}

// DirSize returns the size of the WAL on the filesystem.
func DirSize(path string) float64 {
	var size int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return float64(size)
}
