package simple

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"sort"
	"sync"
	"time"

	badgerdb "github.com/dgraph-io/badger/v3"
	"github.com/grafana/agent/pkg/flow/logging"
)

type signaldb struct {
	mut          sync.RWMutex
	d            *badgerdb.DB
	log          *logging.Logger
	currentIndex uint64
}

func newDb(dir string, l *logging.Logger) (*signaldb, error) {
	bdb, err := badgerdb.Open(badgerdb.DefaultOptions(dir))
	if err != nil {
		return nil, err
	}

	newDb := &signaldb{
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

func (d *signaldb) getNewKey() uint64 {
	d.mut.Lock()
	defer d.mut.Unlock()

	d.currentIndex = d.currentIndex + 1
	return d.currentIndex
}

func (d *signaldb) getOldestKey() uint64 {
	d.mut.RLock()
	defer d.mut.Unlock()

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

func (d *signaldb) getKeys() ([]uint64, error) {
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

func (d *signaldb) getCurrentKey() uint64 {
	d.mut.RLock()
	defer d.mut.RUnlock()

	return d.currentIndex
}

// getNextKey may return the passed in key if that is all that exists.
func (d *signaldb) getNextKey(k uint64) uint64 {
	d.mut.RLock()
	defer d.mut.RUnlock()

	if k == d.currentIndex {
		return k
	}
	return k + 1
}

func (d *signaldb) getValueForKey(k []byte) (any, bool, error) {
	var value []byte
	var found bool
	var t int8
	err := d.d.View(func(txn *badgerdb.Txn) error {
		item, err := txn.Get(k)
		if err == badgerdb.ErrKeyNotFound {
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
	buf := bytes.NewBuffer(value)
	dec := gob.NewDecoder(buf)
	val := getRecord(t)
	err = dec.Decode(val)
	return val, found, err
}

func (d *signaldb) getRecordByString(key string) (any, bool, error) {
	return d.getValueForKey([]byte(key))
}

func (d *signaldb) getRecordByUint(key uint64) (any, bool, error) {
	buf := make([]byte, 8)
	binary.PutUvarint(buf, key)
	return d.getValueForKey(buf)
}

// writeRecordWithAutoKey will gob encode the data and set a TTL from now.
// If the TTL expires you will not be able to retrieve the value though the space will not be retrieved until
// the system cleans up TTLs every few minutes.
// writeRecordWithAutoKey will return the key of the value inserted.
// This key is always greater than a previously entered key.
func (d *signaldb) writeRecordWithAutoKey(data any, ttl time.Duration) (uint64, error) {
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
func (d *signaldb) writeRecord(key []byte, data any, ttl time.Duration) error {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return err
	}
	signalType, err := getType(data)
	if err != nil {
		return err
	}
	err = d.d.Update(func(txn *badgerdb.Txn) error {
		entry := &badgerdb.Entry{
			Key:      key,
			Value:    buf.Bytes(),
			UserMeta: byte(signalType),
		}
		if ttl > 0*time.Second {
			entry.ExpiresAt = uint64(time.Now().Add(ttl).Unix())
		}
		inErr := txn.SetEntry(entry)
		return inErr
	})
	return err
}

func getType(data any) (int8, error) {
	switch v := data.(type) {
	case []Sample:
		return metric_signal, nil
	case []Exemplar:
		return exemplar_signal, nil
	case []Metadata:
		return metadata_signal, nil
	case []Histogram:
		return histogram_signal, nil
	case []FloatHistogram:
		return floathistogram_signal, nil
	case Bookmark:
		return bookmark_type, nil
	default:
		return 0, fmt.Errorf("unknown data type %v", v)
	}
}

func getRecord(t int8) any {
	switch t {
	case metric_signal:
		return []Sample{}
	case exemplar_signal:
		return []Exemplar{}
	case metadata_signal:
		return []Metadata{}
	case histogram_signal:
		return []Histogram{}
	case floathistogram_signal:
		return []FloatHistogram{}
	case bookmark_type:
		return &Bookmark{}
	default:
		return nil
	}
}

func (d *signaldb) evict() {
	d.mut.Lock()
	defer d.mut.Unlock()
	var err error
	for err == nil {
		// Reclaim if we can gain 10% of space back.
		err = d.d.RunValueLogGC(0.1)
		if err != nil {
			return
		}
	}

}
