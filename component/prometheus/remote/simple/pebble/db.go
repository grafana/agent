package pebble

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"sync"
	"time"

	"github.com/go-kit/log"

	"github.com/go-kit/log/level"
	"go.uber.org/atomic"

	pdb "github.com/cockroachdb/pebble"
	"github.com/golang/snappy"
)

// DB is a wrapper around the pebbleDB.
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

// NewDB creates a new DB, getValue and getType allow conversion of the []byte into real types. GetType returns the type
// of an object and that is encoded into the stored value. Then getValue takes in that type to convert it into a real object.
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
		// Return a 1 MB buffer, this may not be big enough and we should maybe have several tiers of buffers.
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

// GetNewKey increments the current key and returns the new value.
func (d *DB) GetNewKey() uint64 {
	d.mut.Lock()
	defer d.mut.Unlock()

	d.currentIndex = d.currentIndex + 1
	d.keyCache.add(d.currentIndex, 0, 0)
	return d.currentIndex
}

// GetOldestKey returns the oldest key, it returns 0 if no keys are found.
func (d *DB) GetOldestKey() uint64 {
	ks := d.keyCache.keys()
	if len(ks) == 0 {
		return 0
	}
	// Keys are garaunteed to be sorted oldest to newest.
	return ks[0]
}

// GetKeys returns all keys sorted by oldest to newest.
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

// GetCurrentKey returns the current index.
func (d *DB) GetCurrentKey() uint64 {
	d.mut.RLock()
	defer d.mut.RUnlock()

	return d.currentIndex
}

// GetNextKey returns the next key that has been allocated. If k is the newest key return k.
func (d *DB) GetNextKey(k uint64) uint64 {
	keys, _ := d.GetKeys()
	for _, lk := range keys {
		if lk > k {
			return lk
		}
	}
	return k
}

// Delete any keys older than k.
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

// GetValueByByte returns the value specified by k, whether it was found and any error.
// An expired TTL is considered not found.
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

// GetValueByString follows GetValueByByte conventions.
func (d *DB) GetValueByString(k string) (any, bool, error) {
	return d.GetValueByByte([]byte(k))
}

// GetValueByKey follows GetValueByByte conventions but also updates the keycache if not found.
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

// WriteValueWithAutokey is GetNewKey + WriteValue and returns the next key. If ttl > 0 it will honor it.
func (d *DB) WriteValueWithAutokey(data any, ttl time.Duration) (uint64, error) {
	nextKey := d.GetNewKey()
	err := d.WriteValue(keyToByte(nextKey), data, ttl)
	return nextKey, err
}

// WriteValue writes a given value into the database.
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

// Evict clears out any expired TTLs and compacts the database.
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

// Size returns the estimated disk usage.
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

// SeriesCount returns the total number of samples in the database.
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

// item represents a value stored in the database.
type item struct {
	Value []byte
	// Type is used to convert Value to a concrete value. Value is Gob Encoded and Snappy compressed.
	Type int8
	// Unix timestamp to expire.
	TTL   int64
	Count int
}
