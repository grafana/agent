package badger

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"sort"
	"strconv"
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

func (d *db) getValueForKey(k uint64, into any) (bool, error) {
	var value []byte
	var found bool
	d.d.View(func(txn *badgerdb.Txn) error {
		buf := make([]byte, 8)
		binary.PutUvarint(buf, k)
		item, err := txn.Get(buf)
		if err == badgerdb.ErrKeyNotFound {
			found = false
			return nil
		}
		found = true
		value, err = item.ValueCopy(nil)
		return err
	})
	buf := bytes.NewBuffer(value)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(into)
	return found, err
}

func (d *db) writeRecords(data any, ttl time.Duration) error {
	if data == nil {
		return nil
	}
	id := d.getNewKey()

	key := []byte(strconv.FormatUint(id, 10))
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	enc.Encode(data)
	err := d.d.Update(func(txn *badgerdb.Txn) error {
		inErr := txn.SetEntry(&badgerdb.Entry{
			Key:       key,
			Value:     buf.Bytes(),
			ExpiresAt: uint64(time.Now().Add(ttl).Unix()),
			UserMeta:  0,
		})
		return inErr
	})
	if err != nil {
		return err
	}
	return nil
}
