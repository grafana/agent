package pebble

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"github.com/golang/snappy"
	"github.com/grafana/agent/pkg/flow/logging"
	"sort"
	"sync"
	"time"
)
import pdb "github.com/cockroachdb/pebble"

type DB struct {
	mut sync.RWMutex
	db  *pdb.DB
	log *logging.Logger
	// Trying to avoid unbounded lists, this thankfully is one key for each commit so its unlikely to be in the millions
	// of active commits.
	keys         []uint64
	currentIndex uint64
	getValue     func([]byte, int8) any
	getType      func(data any) (int8, error)
}

func NewDB(dir string, getValue func([]byte, int8) any, getType func(data any) (int8, error), l *logging.Logger) (*DB, error) {
	pebbleDB, err := pdb.Open(dir, &pdb.Options{})

	if err != nil {
		return nil, err
	}
	d := &DB{
		db:       pebbleDB,
		getType:  getType,
		getValue: getValue,
		log:      l,
		keys:     make([]uint64, 0),
	}
	keys, err := d.GetKeys()
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		d.currentIndex = 1
	} else {
		d.currentIndex = keys[len(keys)-1]
	}
	return d, nil
}

func (d *DB) GetNewKey() uint64 {
	d.mut.Lock()
	defer d.mut.Unlock()

	d.currentIndex = d.currentIndex + 1
	d.keys = append(d.keys, d.currentIndex)
	return d.currentIndex
}

func (d *DB) GetOldestKey() uint64 {
	iter, _ := d.db.NewIter(&pdb.IterOptions{})
	defer iter.Close()
	if iter.Last() {
		return byteToKey(iter.Key())
	}
	return 0
}

func (d *DB) GetKeys() ([]uint64, error) {
	d.mut.Lock()
	defer d.mut.Unlock()

	// Return the cached keys
	if len(d.keys) != 0 {
		retKeys := make([]uint64, len(d.keys))
		copy(retKeys, d.keys)
		return retKeys, nil
	}

	iter, _ := d.db.NewIter(&pdb.IterOptions{})
	defer iter.Close()
	if iter.First() {
		d.keys = append(d.keys, byteToKey(iter.Key()))
	}

	for iter.Next() {
		d.keys = append(d.keys, byteToKey(iter.Key()))
	}
	sort.Slice(d.keys, func(i, j int) bool { return d.keys[i] < d.keys[j] })
	return d.keys, nil
}

func (d *DB) GetCurrentKey() uint64 {
	d.mut.RLock()
	defer d.mut.RUnlock()

	return d.currentIndex
}

func (d *DB) GetNextKey(k uint64) uint64 {
	keys, _ := d.GetKeys()
	for _, lk := range keys {
		if lk > k {
			return lk
		}
	}
	return k
}

func (d *DB) DeleteKeysOlderThan(k uint64) {
	d.mut.Lock()
	defer d.mut.Unlock()

	keys, _ := d.GetKeys()
	batch := d.db.NewBatch()

	for _, lk := range keys {
		if lk >= k {
			continue
		}
		batch.Delete(keyToByte(lk), nil)
	}
	// Force a refresh of keys.
	d.keys = make([]uint64, 0)
	_, _ = d.GetKeys()
	batch.Commit(&pdb.WriteOptions{Sync: true})
	batch.Close()
}

func (d *DB) GetValueByByte(k []byte) (any, bool, error) {
	val, closer, err := d.db.Get(k)
	if closer != nil {
		defer closer.Close()
	}
	if errors.Is(err, pdb.ErrNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	unsnapped, err := snappy.Decode(nil, val)
	if err != nil {
		return nil, false, err
	}
	buf := bytes.NewBuffer(unsnapped)
	dec := gob.NewDecoder(buf)
	it := &item{}
	err = dec.Decode(it)
	if err != nil {
		return nil, false, err
	}
	// TTL is implemented on pulling the record.
	if it.TTL < time.Now().Unix() {
		return nil, false, nil
	}
	finalVal := d.getValue(it.Value, it.Type)
	return finalVal, true, nil
}

func (d *DB) GetValueByString(k string) (any, bool, error) {
	return d.GetValueByByte([]byte(k))
}

func (d *DB) GetValueByUint(k uint64) (any, bool, error) {
	return d.GetValueByByte(keyToByte(k))
}

func (d *DB) WriteValueWithAutokey(data any, ttl time.Duration) (uint64, error) {
	nextKey := d.GetNewKey()
	return nextKey, d.WriteValue(keyToByte(nextKey), data, ttl)
}

func (d *DB) WriteValue(key []byte, data any, ttl time.Duration) error {
	t, err := d.getType(data)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err = enc.Encode(data)
	if err != nil {
		return err
	}
	it := &item{}
	it.Value = buf.Bytes()
	it.Type = t
	if ttl > 0 {
		it.TTL = time.Now().Add(ttl).Unix()
	}
	buf = bytes.NewBuffer(nil)
	enc = gob.NewEncoder(buf)
	err = enc.Encode(it)
	if err != nil {
		return err
	}
	snappied := snappy.Encode(nil, buf.Bytes())
	return d.db.Set(key, snappied, &pdb.WriteOptions{Sync: true})
}

func (d *DB) Evict() error {
	if len(d.keys) == 0 {
		return nil
	}
	return d.db.Compact(keyToByte(d.keys[0]), keyToByte(d.keys[len(d.keys)-1]), true)
}

func (d *DB) Size() uint64 {
	if len(d.keys) == 0 {
		return 0
	}
	size, _ := d.db.EstimateDiskUsage(keyToByte(d.keys[0]), keyToByte(d.keys[len(d.keys)-1]))
	return size
}

func byteToKey(b []byte) uint64 {
	buf := bytes.NewBuffer(b)
	key, err := binary.ReadUvarint(buf)
	if err != nil {
		return 0
	}
	return key
}

func keyToByte(k uint64) []byte {
	buf := make([]byte, 8)
	binary.PutUvarint(buf, k)
	return buf
}

type item struct {
	Value []byte
	Type  int8
	TTL   int64
}
