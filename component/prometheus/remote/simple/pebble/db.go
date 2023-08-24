package pebble

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"sync"
	"time"

	"github.com/go-kit/log/level"

	"github.com/golang/snappy"
	"github.com/grafana/agent/pkg/flow/logging"

	pdb "github.com/cockroachdb/pebble"
)

type DB struct {
	mut sync.RWMutex
	db  *pdb.DB
	log *logging.Logger
	// Trying to avoid unbounded lists, this thankfully is one key for each commit so its unlikely to be in the millions
	// of active commits.
	keyCache     *metadata
	currentIndex uint64
	getValue     func([]byte, int8) any
	getType      func(data any) (int8, int, error)
}

func NewDB(dir string, getValue func([]byte, int8) any, getType func(data any) (int8, int, error), l *logging.Logger) (*DB, error) {
	pebbleDB, err := pdb.Open(dir, &pdb.Options{})
	if err != nil {
		return nil, err
	}
	d := &DB{
		db:       pebbleDB,
		getType:  getType,
		getValue: getValue,
		log:      l,
		keyCache: newMetadata(),
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
	d.keyCache.add(d.currentIndex, 0, 0)
	return d.currentIndex
}

func (d *DB) GetOldestKey() uint64 {
	ks := d.keyCache.keys()
	if len(ks) == 0 {
		return 0
	}
	// Keys are garaunteed to be sorted oldest to newest.
	return ks[0]
}

func (d *DB) GetKeys() ([]uint64, error) {
	d.mut.Lock()
	defer d.mut.Unlock()

	// Return the cached keys, if they exist.
	ks := d.keyCache.keys()
	if len(ks) > 0 {
		return ks, nil
	}

	iter, _ := d.db.NewIter(&pdb.IterOptions{})
	defer iter.Close()
	if iter.First() {
		it, err := d.convertItem(iter.Value())
		if err != nil {
			return nil, err
		}
		d.keyCache.add(byteToKey(iter.Key()), it.TTL, it.Count)
	}

	for iter.Next() {
		it, err := d.convertItem(iter.Value())
		if err != nil {
			return nil, err
		}
		d.keyCache.add(byteToKey(iter.Key()), it.TTL, it.Count)
	}
	return d.keyCache.keys(), nil
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
	ks, _ := d.GetKeys()
	batch := d.db.NewBatch()

	for _, lk := range ks {
		if lk >= k {
			continue
		}
		err := batch.Delete(keyToByte(lk), nil)
		if err != nil {
			level.Error(d.log).Log("msg", "error deleting key", "key", lk, "err", err)
		}
	}
	// Force a refresh of keys.
	d.keyCache.clear()
	_, _ = d.GetKeys()
	err := batch.Commit(&pdb.WriteOptions{Sync: true})
	if err != nil {
		level.Error(d.log).Log("msg", "error committing batch", "err", err)
	}
	err = batch.Close()
	if err != nil {
		level.Error(d.log).Log("msg", "error closing batch", "err", err)
	}
}

func (d *DB) GetValueByByte(k []byte) (any, bool, error) {
	it, found, err := d.getItem(k)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, found, err
	}
	// TTL is implemented on pulling the record.
	if it.TTL < time.Now().Unix() {
		return nil, false, nil
	}
	finalVal := d.getValue(it.Value, it.Type)
	return finalVal, true, nil
}

func (d *DB) getItem(k []byte) (*item, bool, error) {
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
	it, err := d.convertItem(val)
	return it, true, err
}

func (d *DB) convertItem(val []byte) (*item, error) {
	unsnapped, err := snappy.Decode(nil, val)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(unsnapped)
	dec := gob.NewDecoder(buf)
	it := &item{}
	err = dec.Decode(it)
	if err != nil {
		return nil, err
	}
	return it, nil
}

func (d *DB) GetValueByString(k string) (any, bool, error) {
	return d.GetValueByByte([]byte(k))
}

func (d *DB) GetValueByUint(k uint64) (any, bool, error) {
	return d.GetValueByByte(keyToByte(k))
}

func (d *DB) WriteValueWithAutokey(data any, ttl time.Duration, buf *bytes.Buffer) (uint64, *bytes.Buffer, error) {
	nextKey := d.GetNewKey()
	retBuf, err := d.WriteValue(keyToByte(nextKey), data, ttl, buf)
	return nextKey, retBuf, err
}

func (d *DB) WriteValue(key []byte, data any, ttl time.Duration, buf *bytes.Buffer) (*bytes.Buffer, error) {
	t, count, err := d.getType(data)
	if err != nil {
		return buf, err
	}
	if buf == nil {
		buf = bytes.NewBuffer(nil)
	}
	enc := gob.NewEncoder(buf)
	err = enc.Encode(data)
	if err != nil {
		return buf, err
	}
	it := &item{}
	it.Value = buf.Bytes()
	it.Type = t
	if ttl > 0 {
		it.TTL = time.Now().Add(ttl).Unix()
	}
	it.Count = count
	buf.Reset()
	enc = gob.NewEncoder(buf)
	err = enc.Encode(it)
	if err != nil {
		return buf, err
	}
	snappied := snappy.Encode(nil, buf.Bytes())
	buf.Reset()
	return buf, d.db.Set(key, snappied, &pdb.WriteOptions{Sync: true})
}

func (d *DB) Evict() error {
	d.mut.Lock()
	defer d.mut.Unlock()

	if d.keyCache.len() == 0 {
		return nil
	}

	// Find all the expired TTLs and remove them.
	expired := d.keyCache.keysWithExpiredTTL(time.Now().Unix())
	for _, k := range expired {
		err := d.db.Delete(keyToByte(k), &pdb.WriteOptions{Sync: true})
		if err != nil {
			return err
		}
	}
	d.keyCache.removeKeys(expired)
	ks := d.keyCache.keys()
	if len(ks) == 0 {
		return nil
	}
	return d.db.Compact(keyToByte(ks[0]), keyToByte(ks[len(ks)-1]), true)
}

func (d *DB) Size() uint64 {
	if d.keyCache.len() == 0 {
		return 0
	}
	ks := d.keyCache.keys()
	if len(ks) == 0 {
		return 0
	}
	size, _ := d.db.EstimateDiskUsage(keyToByte(ks[0]), keyToByte(ks[len(ks)-1]))
	return size
}

func (d *DB) SeriesCount() int64 {
	return d.keyCache.seriesLen()
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
	Count int
}
