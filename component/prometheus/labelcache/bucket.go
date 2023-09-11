package labelcache

import (
	"arena"
	"bytes"
	"errors"
	"fmt"

	"github.com/cockroachdb/pebble"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// bucket is a prefixed accessor for values in a pebbledb.
// Both the prefix and name have to be unique.
type bucket struct {
	prefix uint8
	name   string
	db     *pebble.DB
	l      log.Logger
}

func newBucket(db *pebble.DB, prefix uint8, name string, l log.Logger) *bucket {
	return &bucket{
		prefix: prefix,
		name:   name,
		db:     db,
		l:      l,
	}
}

func (b *bucket) writeValues(keys [][]byte, values [][]byte, mem *arena.Arena) error {
	if len(keys) != len(values) {
		return fmt.Errorf("keys %d and values %d must be the same length", len(keys), len(values))
	}
	if mem == nil {
		return fmt.Errorf("arena must not be nil")
	}
	batch := b.db.NewBatch()
	for i := 0; i < len(keys); i++ {
		buf := b.makeKey(keys[i], mem)
		err := batch.Set(buf, values[i], nil)
		if err != nil {
			return err
		}
	}
	return batch.Commit(pebble.NoSync)
}

func (b *bucket) getValues(keys [][]byte, mem *arena.Arena) ([][]byte, error) {
	if mem == nil {
		return nil, fmt.Errorf("arena must not be nil")
	}
	returnBuf := arena.MakeSlice[[]byte](mem, len(keys), len(keys))

	for i := 0; i < len(keys); i++ {
		keyBuf := b.makeKey(keys[i], mem)
		val, closer, err := b.db.Get(keyBuf)
		if errors.Is(err, pebble.ErrNotFound) {
			returnBuf[i] = nil
			continue
		}
		// val is only usable until closed is called so we need to copy it.
		valBuf := arena.MakeSlice[byte](mem, len(val), len(val))
		copy(valBuf, val)
		closer.Close()
		returnBuf[i] = valBuf
	}
	return returnBuf, nil
}

func (b *bucket) getNewestID() []byte {
	iter, err := b.db.NewIter(&pebble.IterOptions{LowerBound: []byte{b.prefix}, UpperBound: []byte{b.prefix, 0}})
	if err != nil {
		level.Error(b.l).Log("msg", "unable to create iterator", "name", b.name)
		return nil
	}
	defer iter.Close()
	if !iter.Last() {
		// nothing found so return nil
		return nil
	}
	val := iter.Value()
	return bytes.Clone(val)
}

func (b *bucket) makeKey(key []byte, mem *arena.Arena) []byte {
	buf := arena.MakeSlice[byte](mem, len(key)+1, len(key)+1)
	buf[0] = b.prefix
	copy(buf[1:], key)
	return buf
}
