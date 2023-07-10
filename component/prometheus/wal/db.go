package wal

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"sort"
	"sync"
	"time"

	badgerdb "github.com/dgraph-io/badger/v3"
	"github.com/grafana/agent/pkg/flow/logging"
)

type db struct {
	mut          sync.RWMutex
	d            *badgerdb.DB
	log          *logging.Logger
	currentIndex uint64
}

func newDb(dir string, l *logging.Logger) (*db, error) {
	bdb, err := badgerdb.Open(badgerdb.DefaultOptions(dir))
	if err != nil {
		return nil, err
	}

	newDb := &db{
		d:   bdb,
		log: l,
	}
	keys, err := newDb.getKeys()
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

func (d *db) getNewKey() uint64 {
	d.mut.Lock()
	defer d.mut.Unlock()

	d.currentIndex = d.currentIndex + 1
	return d.currentIndex
}

func (d *db) getKeys() ([]uint64, error) {
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
		return []uint64{}, nil
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i] < ret[j] })
	return ret, nil
}

func (d *db) getCurrentKey() uint64 {
	d.mut.RLock()
	defer d.mut.RUnlock()

	return d.currentIndex
}

// getNextKey may return the passed in key if that is all that exists.
func (d *db) getNextKey(k uint64) uint64 {
	d.mut.RLock()
	defer d.mut.RUnlock()

	if k == d.currentIndex {
		return k
	}
	return k + 1
}

func (d *db) getValueForKey(k []byte, into any) (bool, error) {
	var value []byte
	var found bool
	err := d.d.View(func(txn *badgerdb.Txn) error {
		item, err := txn.Get(k)
		if err == badgerdb.ErrKeyNotFound {
			found = false
			return nil
		}
		found = true
		value, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		return false, err
	}
	buf := bytes.NewBuffer(value)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(into)
	return found, err
}

func (d *db) getRecordByString(key string, into any) (bool, error) {
	return d.getValueForKey([]byte(key), into)
}

func (d *db) getRecordByUint(key uint64, into any) (bool, error) {
	buf := make([]byte, 8)
	binary.PutUvarint(buf, key)
	return d.getValueForKey(buf, into)
}

// writeRecordWithAutoKey will gob encode the data and set a TTL from now. Note the TTL may not trigger at exactly
// the TTL. The system checks for TTLs every few minutes. writeRecordWithAutoKey will return the key of the value inserted.
// This key is always greater than a previously entered key.
func (d *db) writeRecordWithAutoKey(data any, ttl time.Duration) (uint64, error) {
	if data == nil {
		return 0, nil
	}
	id := d.getNewKey()
	keyBuf := make([]byte, 8)
	binary.PutUvarint(keyBuf, id)
	err := d.writeRecord(keyBuf, data, ttl)
	return id, err
}

// writeRecord writes a value and assumes the data is a pointer and will gob encode it. If a TTL is specified then will set
// the expiration.
func (d *db) writeRecord(key []byte, data any, ttl time.Duration) error {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	enc.Encode(data)
	err := d.d.Update(func(txn *badgerdb.Txn) error {
		entry := &badgerdb.Entry{
			Key:      key,
			Value:    buf.Bytes(),
			UserMeta: 0,
		}
		if ttl > 0*time.Second {
			entry.ExpiresAt = uint64(time.Now().Add(ttl).Unix())
		}
		inErr := txn.SetEntry(entry)
		return inErr
	})
	return err
}
