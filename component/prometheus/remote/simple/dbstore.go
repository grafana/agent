package simple

import (
	"context"
	"path"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/flow/logging"
)

type dbstore struct {
	mut       sync.RWMutex
	l         *logging.Logger
	dir       string
	ttlUpdate time.Duration
	ttl       time.Duration
	inMemory  bool
	sampleDB  *signaldb
	bookmark  *signaldb
	callbacks []func(oldestID uint64)
	ctx       context.Context
}

const (
	metric_signal int8 = iota
	histogram_signal
	floathistogram_signal
	metadata_signal
	exemplar_signal
	bookmark_type
)

func newDBStore(inMemory bool, ttl time.Duration, ttlUpdate time.Duration, directory string, l *logging.Logger) (*dbstore, error) {
	bookmark, err := newDb(path.Join(directory, "bookmark"), l)
	if err != nil {
		return nil, err
	}

	store := &dbstore{
		ttlUpdate: ttlUpdate,
		inMemory:  inMemory,
		bookmark:  bookmark,
		callbacks: make([]func(oldestID uint64), 0),
	}
	return store, nil
}

func (dbs *dbstore) Run(ctx context.Context) {
	dbs.ctx = ctx
	go dbs.startTTL()
}

func (dbs *dbstore) startTTL() {
	ttlTimer := time.NewTicker(dbs.ttlUpdate)
	for {
		select {
		case <-ttlTimer.C:
			// Start eviction
			dbs.evict()
		case <-dbs.ctx.Done():
			return
		}

	}
}

func (dbs *dbstore) evict() {
	dbs.mut.Lock()
	dbs.sampleDB.evict()
	dbs.mut.Unlock()
}

func (dbs *dbstore) WriteBookmark(key string, value any) error {
	return dbs.bookmark.writeRecord([]byte(key), value, 0*time.Second)
}

func (dbs *dbstore) GetBookmark(key string) (*Bookmark, bool) {
	bk, found, _ := dbs.bookmark.getRecordByString(key)
	return bk.(*Bookmark), found
}

func (dbs *dbstore) WriteSignal(value any) (uint64, error) {
	return dbs.sampleDB.writeRecordWithAutoKey(value, dbs.ttl)
}

func (dbs *dbstore) GetSignal(key uint64) (any, bool) {
	val, found, err := dbs.sampleDB.getRecordByUint(key)
	if err != nil {
		level.Error(dbs.l).Log("error finding key", err, "key", key)
		return false
	}
	return found
}
