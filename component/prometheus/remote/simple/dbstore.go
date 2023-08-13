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
	ctx       context.Context
}

const (
	metricSignal int8 = iota
	histogramSignal
	floathistogramSignal
	metadataSignal
	exemplarSignal
	bookmarkType
)

func newDBStore(inMemory bool, ttl time.Duration, ttlUpdate time.Duration, directory string, l *logging.Logger) (*dbstore, error) {
	bookmark, err := newDb(path.Join(directory, "bookmark"), l)
	if err != nil {
		return nil, err
	}
	sample, err := newDb(path.Join(directory, "signals"), l)
	if err != nil {
		return nil, err
	}
	store := &dbstore{
		ttlUpdate: ttlUpdate,
		inMemory:  inMemory,
		bookmark:  bookmark,
		sampleDB:  sample,
		ttl:       ttl,
		l:         l,
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
	if bk == nil {
		return &Bookmark{Key: 1}, false
	}
	return bk.(*Bookmark), found
}

func (dbs *dbstore) WriteSignal(value any) (uint64, error) {

	key, err := dbs.sampleDB.writeRecordWithAutoKey(value, dbs.ttl)
	level.Info(dbs.l).Log("writing signals to db with key", key)
	return key, err
}

func (dbs *dbstore) GetOldestKey() uint64 {
	return dbs.sampleDB.getOldestKey()
}

func (dbs *dbstore) GetNextKey(k uint64) uint64 {
	return dbs.sampleDB.getNextKey(k)
}

func (dbs *dbstore) GetSignal(key uint64) (any, bool) {
	val, found, err := dbs.sampleDB.getRecordByUint(key)
	if err != nil {
		level.Error(dbs.l).Log("error finding key", err, "key", key)
		return nil, false
	}
	return val, found
}
