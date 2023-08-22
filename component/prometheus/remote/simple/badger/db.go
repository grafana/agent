package badger

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	badgerdb "github.com/dgraph-io/badger/v4"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/flow/logging"
)

type signaldb struct {
	mut          sync.RWMutex
	d            *badgerdb.DB
	log          *logging.Logger
	currentIndex uint64
	getValue     func([]byte, int8) any
	getType      func(data any) (int8, error)
}

func NewDB(dir string, defaultDBSize int64, l *logging.Logger, getValue func([]byte, int8) any, getType func(data any) (int8, error)) (*signaldb, error) {
	opts := badgerdb.DefaultOptions(dir)
	opts.SyncWrites = true
	opts.MetricsEnabled = true
	opts.ValueLogFileSize = defaultDBSize
	opts.NumVersionsToKeep = 0
	opts.MaxLevels = 1
	opts.NumLevelZeroTables = 0
	opts.NumLevelZeroTablesStall = 2
	opts.CompactL0OnClose = true
	opts.Logger = &dbLog{l: l}
	bdb, err := badgerdb.Open(opts)
	if err != nil {
		return nil, err
	}

	newDb := &signaldb{
		d:        bdb,
		log:      l,
		getValue: getValue,
		getType:  getType,
	}
	keys, err := newDb.GetKeys()
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		newDb.currentIndex = 0
	} else {
		newDb.currentIndex = keys[len(keys)-1]
	}
	return newDb, nil
}

func (d *signaldb) GetNewKey() uint64 {
	d.mut.Lock()
	defer d.mut.Unlock()

	d.currentIndex = d.currentIndex + 1
	return d.currentIndex
}

func (d *signaldb) GetOldestKey() uint64 {
	var buf []byte
	d.d.View(func(txn *badgerdb.Txn) error {
		iterator := txn.NewIterator(badgerdb.IteratorOptions{
			PrefetchSize: 1,
		})
		defer iterator.Close()
		if iterator.Valid() {
			buf = iterator.Item().KeyCopy(nil)
		}
		return nil
	})
	buff := bytes.NewBuffer(buf)
	key, _ := binary.ReadUvarint(buff)
	return key
}

func (d *signaldb) GetKeys() ([]uint64, error) {
	ret := make([]uint64, 0)
	err := d.d.View(func(txn *badgerdb.Txn) error {
		opt := badgerdb.DefaultIteratorOptions
		opt.PrefetchValues = false
		it := txn.NewIterator(opt)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			buf := bytes.NewBuffer(it.Item().Key())
			k, _ := binary.ReadUvarint(buf)
			ret = append(ret, k)
		}
		return nil
	})
	if err != nil {
		return []uint64{}, err
	}

	sort.Slice(ret, func(i, j int) bool { return ret[i] < ret[j] })
	return ret, nil
}

func (d *signaldb) GetCurrentKey() uint64 {
	d.mut.RLock()
	defer d.mut.RUnlock()

	return d.currentIndex
}

// GetNextKey may return the passed in key if that is all that exists.
func (d *signaldb) GetNextKey(k uint64) uint64 {
	d.mut.RLock()
	defer d.mut.RUnlock()

	if k == d.currentIndex {
		return k
	}
	if k > d.currentIndex {
		return d.currentIndex
	}
	return k + 1
}

func (d *signaldb) GetValueByByte(k []byte) (any, bool, error) {
	var value []byte
	var found bool
	var t int8
	err := d.d.View(func(txn *badgerdb.Txn) error {
		item, err := txn.Get(k)
		if errors.Is(err, badgerdb.ErrKeyNotFound) {
			found = false
			return nil
		}
		found = true
		value, err = item.ValueCopy(nil)
		t = int8(item.UserMeta())
		return err
	})
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}
	smp := d.getValue(value, t)
	return smp, found, err
}

func (d *signaldb) GetValueByString(key string) (any, bool, error) {
	return d.GetValueByByte([]byte(key))
}

func (d *signaldb) GetValueByUint(key uint64) (any, bool, error) {
	buf := make([]byte, 8)
	binary.PutUvarint(buf, key)
	return d.GetValueByByte(buf)
}

// WriteValueWithAutokey will gob encode the data and set a TTL from now.
// If the TTL expires you will not be able to retrieve the value though the space will not be retrieved until
// the system cleans up TTLs every few minutes.
// writeRecordWithAutoKey will return the key of the value inserted.
// This key is always greater than a previously entered key.
func (d *signaldb) WriteValueWithAutokey(data any, ttl time.Duration) (uint64, error) {
	if data == nil {
		return 0, nil
	}
	id := d.GetNewKey()
	keyBuf := make([]byte, 8)
	binary.PutUvarint(keyBuf, id)
	err := d.WriteValue(keyBuf, data, ttl)
	return id, err
}

// WriteValue writes a value and assumes the data is a pointer and will gob encode it. If a TTL is specified then will set
// the expiration.
func (d *signaldb) WriteValue(key []byte, data any, ttl time.Duration) error {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return err
	}
	signalType, err := d.getType(data)
	if err != nil {
		return err
	}
	err = d.d.Update(func(txn *badgerdb.Txn) error {
		entry := &badgerdb.Entry{
			Key:      key,
			Value:    buf.Bytes(),
			UserMeta: byte(signalType),
		}
		if ttl > 0 {
			entry.ExpiresAt = uint64(time.Now().Add(ttl).Unix())
		}
		inErr := txn.SetEntry(entry)
		return inErr
	})
	return err
}

func (d *signaldb) Evict() error {
	d.mut.Lock()
	defer d.mut.Unlock()
	var err error
	oldLsm, oldVlog := d.d.Size()
	oldTotal := oldLsm + oldVlog
	for i := 0; i < 10; i++ {
		for !errors.Is(err, badgerdb.ErrNoRewrite) {
			// Reclaim if we can gain 10% of space back.
			err = d.d.RunValueLogGC(0.01)
		}
	}
	newLsm, newVlog := d.d.Size()
	level.Info(d.log).Log("msg", "eviction completed and reclaimed bytes", "reclaimed", oldTotal-(newLsm+newVlog))
	return nil
}

func (d *signaldb) Size() uint64 {
	lsm, vlog := d.d.Size()
	return uint64(lsm + vlog)
}

func (d *signaldb) DeleteKeysOlderThan(oldKey uint64) {
	keys, _ := d.GetKeys()
	_ = d.d.Update(func(txn *badgerdb.Txn) error {
		for _, k := range keys {
			// Only delete keys older than this and since keys are ALWAYS autoincrementing we are good.
			if k >= oldKey {
				continue
			}
			kbuf := make([]byte, 8)
			binary.PutUvarint(kbuf, k)
			err := txn.Delete(kbuf)
			if err != nil {
				level.Error(d.log).Log("msg", "error deleting key", "key", k, "err", err)
			}
		}
		return nil
	})
}

type dbLog struct {
	l *logging.Logger
}

func (dbl *dbLog) Errorf(s string, args ...interface{}) {
	level.Error(dbl.l).Log("msg", fmt.Sprintf(s, args...))
}
func (dbl *dbLog) Warningf(s string, args ...interface{}) {
	level.Warn(dbl.l).Log("msg", fmt.Sprintf(s, args...))
}
func (dbl *dbLog) Infof(s string, args ...interface{}) {
	level.Info(dbl.l).Log("msg", fmt.Sprintf(s, args...))
}
func (dbl *dbLog) Debugf(s string, args ...interface{}) {
	level.Debug(dbl.l).Log("msg", fmt.Sprintf(s, args...))
}
