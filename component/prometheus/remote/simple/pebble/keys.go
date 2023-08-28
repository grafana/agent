package pebble

import (
	"sort"
	"sync"
)

// metadata is used to hold various metadata for pebbledb. This is so we don't have to slam the disk
// whenever we want to check for something. It is an unbounded queue but it has a 1 to 1 mapping with commits
// and therefore should not grow to millions.
type metadata struct {
	mut  sync.RWMutex
	ks   []uint64
	ttls map[uint64]int64
	size map[uint64]int
}

func newMetadata() *metadata {
	return &metadata{
		ks:   []uint64{},
		ttls: map[uint64]int64{},
		size: map[uint64]int{},
	}
}

func (ks *metadata) add(k uint64, ttl int64, size int) {
	ks.mut.Lock()
	defer ks.mut.Unlock()

	ks.ks = append(ks.ks, k)
	ks.ttls[k] = ttl
	ks.size[k] = size
	sort.Slice(ks.ks, func(i, j int) bool { return ks.ks[i] < ks.ks[j] })
}

func (ks *metadata) keys() []uint64 {
	ks.mut.RLock()
	defer ks.mut.RUnlock()

	if len(ks.ks) != 0 {
		retKeys := make([]uint64, len(ks.ks))
		copy(retKeys, ks.ks)
		return retKeys
	}
	return make([]uint64, 0)
}

func (ks *metadata) clear() {
	ks.mut.Lock()
	defer ks.mut.Unlock()

	ks.ks = make([]uint64, 0)
	ks.ttls = make(map[uint64]int64)
	ks.size = make(map[uint64]int)
}

func (ks *metadata) len() int {
	ks.mut.RLock()
	defer ks.mut.RUnlock()

	return len(ks.ks)
}

// keysWithExpiredTTL returns any keys that are older than the TTL (unix timestamp).
func (ks *metadata) keysWithExpiredTTL(ttl int64) []uint64 {
	ks.mut.RLock()
	defer ks.mut.RUnlock()

	expired := make([]uint64, 0)
	for k, v := range ks.ttls {
		if v < ttl {
			expired = append(expired, k)
		}
	}
	return expired
}

func (ks *metadata) removeKeys(keys []uint64) {
	ks.mut.Lock()
	defer ks.mut.Unlock()

	if len(keys) == 0 {
		return
	}

	if len(keys) == 1 {
		// If there is only one key then see if it even exists.
		if _, found := ks.size[keys[0]]; !found {
			return
		}
	}
	newKS := make([]uint64, 0)
	// Find all non matching items.
	for _, k := range ks.ks {
		found := false
		for _, in := range keys {
			if k == in {
				found = true
				break
			}
		}
		if !found {
			newKS = append(newKS, k)
		}
	}
	// Delete the TTLs and size
	for _, k := range keys {
		delete(ks.ttls, k)
		delete(ks.size, k)
	}

	ks.ks = newKS
}

func (ks *metadata) seriesLen() int64 {
	ks.mut.RLock()
	defer ks.mut.RUnlock()

	var total int64
	for _, v := range ks.size {
		total = total + int64(v)
	}
	return total
}
