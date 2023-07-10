package wal

import (
	"path"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/flow/logging"
)

var _ Store = (*dbstore)(nil)

type dbstore struct {
	mut       sync.RWMutex
	l         *logging.Logger
	dir       string
	ttlUpdate time.Duration
	ttl       time.Duration
	inMemory  bool
	databases map[string]*db
	bookmark  *db
	callbacks []func(table string, deletedIDs []uint64)
}

func newDBStore(inMemory bool, ttl time.Duration, ttlUpdate time.Duration, directory string, l *logging.Logger) (*dbstore, error) {
	bookmark, err := newDb(path.Join(directory, "bookmark"), l)
	if err != nil {
		return nil, err
	}

	return &dbstore{
		ttlUpdate: ttlUpdate,
		inMemory:  inMemory,
		databases: make(map[string]*db),
		bookmark:  bookmark,
		callbacks: make([]func(table string, deletedIDs []uint64), 0),
	}, nil
}

func (dbs *dbstore) WriteBookmark(key string, value any) error {
	dbs.bookmark.writeRecord([]byte(key), value, 0*time.Second)
	return nil
}

func (dbs *dbstore) GetBookmark(key string, into any) bool {
	found, _ := dbs.bookmark.getRecordByString(key, into)
	return found
}

func (dbs *dbstore) WriteSignal(table string, value any) (uint64, error) {
	foundStore, err := dbs.getTable(table)
	if err != nil {
		level.Error(dbs.l).Log("error finding table", err, "table", table)
	}
	return foundStore.writeRecordWithAutoKey(value, dbs.ttl)
}

func (dbs *dbstore) GetSignal(table string, key uint64, value any) bool {
	foundStore, err := dbs.getTable(table)
	if err != nil {
		level.Error(dbs.l).Log("error finding table", err, "table", table, "key", key)
	}
	found, err := foundStore.getRecordByUint(key, value)
	if err != nil {
		level.Error(dbs.l).Log("error finding key", err, "table", table, "key", key)
		return false
	}
	return found
}

func (dbs *dbstore) getTable(table string) (*db, error) {
	dbs.mut.RLock()
	foundStore, found := dbs.databases[table]
	dbs.mut.RUnlock()
	if found {
		return foundStore, nil
	}
	dbs.mut.Lock()
	foundStore, err := newDb(path.Join(dbs.dir, table), dbs.l)
	dbs.mut.Unlock()
	if err != nil {
		level.Error(dbs.l).Log("error create table", err, "name", table)
		return nil, err
	}
	return foundStore, nil
}

func (dbs *dbstore) RegisterTTLCallback(f func(table string, deletedIDs []uint64)) {
	dbs.mut.Lock()
	defer dbs.mut.Unlock()

	dbs.callbacks = append(dbs.callbacks, f)
}

func (dbs *dbstore) WriteSignalCache(table string, key string, value any) error {
	return nil
}

func (dbs *dbstore) GetSignalCache(table string, key string, into any) bool {
	return false
}
