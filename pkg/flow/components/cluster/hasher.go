package cluster

import (
	"github.com/buraksezer/consistent"
	"github.com/cespare/xxhash"
)

type hasher struct {
	c *consistent.Consistent
}

func newHasher() *hasher {
	h := &hasher{}
	// Create a new consistent instance
	cfg := consistent.Config{
		PartitionCount:    256,
		ReplicationFactor: 2,
		Load:              1.25,
		Hasher:            h,
	}
	c := consistent.New(nil, cfg)
	h.c = c
	return h
}

func (h *hasher) ownedKeys(keys []string, self string, members []string) []string {
	ownedKeys := make([]string, 0)
	for _, m := range members {
		h.c.Add(myMember(m))
	}

	for _, k := range keys {
		m := h.c.LocateKey([]byte(k))
		if m.String() == self {
			ownedKeys = append(ownedKeys, k)
			continue
		}
	}
	return ownedKeys

}

// In your code, you probably have a custom data type
// for your cluster members. Just add a String function to implement
// consistent.Member interface.
type myMember string

func (m myMember) String() string {
	return string(m)
}

func (h *hasher) Sum64(data []byte) uint64 {
	// you should use a proper hash function for uniformity.
	return xxhash.Sum64(data)
}
