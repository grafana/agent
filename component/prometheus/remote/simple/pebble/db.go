package pebble

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"github.com/go-kit/log"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"go.uber.org/atomic"

	pdb "github.com/cockroachdb/pebble"
	"github.com/golang/snappy"
)

type DB struct {
	mut sync.RWMutex
	db  *pdb.DB
	log log.Logger
	// Trying to avoid unbounded lists, this thankfully is one key for each commit so its unlikely to be in the millions
	// of active commits. KeyCache really doesnt make sense for bookmarks.
	keyCache              *metadata
	currentIndex          uint64
	getValue              func([]byte, int8) (any, error)
	getType               func(data any) (int8, int, error)
	bufPool               sync.Pool
	numberOfCompressions  *atomic.Uint64
	totalCompressionRatio *atomic.Float64
}

func NewDB(dir string, getValue func([]byte, int8) (any, error), getType func(data any) (int8, int, error), l log.Logger) (*DB, error) {
	pebbleDB, err := pdb.Open(dir, &pdb.Options{})
	if err != nil {
		return nil, err
	}
	d := &DB{
		db:                    pebbleDB,
		getType:               getType,
		getValue:              getValue,
		log:                   l,
		keyCache:              newMetadata(),
		numberOfCompressions:  atomic.NewUint64(0),
		totalCompressionRatio: atomic.NewFloat64(0),
	}
	d.bufPool.New = func() any {
		// Return a 1 MB buffer
		b := make([]byte, 0, 1024*1024)
		return b
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
	if it.TTL != 0 && it.TTL < time.Now().Unix() {
		// Lets go ahead and clear it out.
		return nil, false, nil
	}
	finalVal, err := d.getValue(it.Value, it.Type)
	return finalVal, true, err
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
	tempBuf := d.bufPool.Get().([]byte)
	defer clear(tempBuf)

	unsnapped, err := snappy.Decode(tempBuf, val)
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

func (d *DB) GetValueByKey(k uint64) (any, bool, error) {
	val, found, err := d.GetValueByByte(keyToByte(k))
	// We are going to do a bit of sleight of hand to keep the keycache in check.
	// Since GetValueByByte doesnt know if its working on a key it cannot handle this.
	// So unilaterly delete this if we dont find it.
	if !found {
		d.keyCache.removeKeys([]uint64{k})
	}
	return val, found, err
}

func (d *DB) WriteValueWithAutokey(data any, ttl time.Duration) (uint64, error) {
	nextKey := d.GetNewKey()
	err := d.WriteValue(keyToByte(nextKey), data, ttl)
	return nextKey, err
}

func (d *DB) WriteValue(key []byte, data any, ttl time.Duration) error {
	t, count, err := d.getType(data)
	if err != nil {
		return err
	}

	tempBuf := d.bufPool.Get().([]byte)
	defer clear(tempBuf)
	buf := bytes.NewBuffer(tempBuf)
	enc := gob.NewEncoder(buf)
	err = enc.Encode(data)
	if err != nil {
		return err
	}
	rawByteCount := buf.Len()
	it := &item{}
	it.Value = buf.Bytes()
	it.Type = t
	if ttl > 0 {
		it.TTL = time.Now().Add(ttl).Unix()
	}
	it.Count = count

	gobBuf := bytes.NewBuffer(nil)
	enc = gob.NewEncoder(gobBuf)
	err = enc.Encode(it)
	if err != nil {
		return err
	}
	snappyBuf := d.bufPool.Get().([]byte)
	defer clear(snappyBuf)
	snappied := snappy.Encode(snappyBuf, gobBuf.Bytes())
	ratio := float64(rawByteCount) / float64(len(snappied))
	d.totalCompressionRatio.Add(ratio)
	d.numberOfCompressions.Add(1)
	return d.db.Set(key, snappied, &pdb.WriteOptions{Sync: true})
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

func (d *DB) AverageCompressionRatio() float64 {
	d.mut.RLock()
	defer d.mut.RUnlock()

	if d.numberOfCompressions.Load() == 0 {
		return 0
	}
	return d.totalCompressionRatio.Load() / float64(d.numberOfCompressions.Load())
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
