package queue

import (
	"sync"

	"github.com/prometheus/prometheus/model/labels"
)

type cache struct {
	mut    sync.RWMutex
	lookup map[string]uint64
	maxKey uint64
}

func (c *cache) encode(l []labels.Labels) [][]int64 {
	c.mut.Lock()
	defer c.mut.Unlock()

	results := make([][]uint64, len(l))
	for i, lbl := range l {
		results[i] = make([]uint64, 2*len(lbl))

		for lblI, v := range lbl {
			kindex, found := c.lookup[v.Name]
			if !found {
				c.maxKey++

				c.lookup[v.Name] = c.maxKey
				kindex = c.maxKey
			}
			vindex, found := c.lookup[v.Value]
			if !found {
				c.maxKey++
				c.lookup[v.Value] = c.maxKey
				vindex = c.maxKey
			}
			results[i][lblI] = kindex
			results[i][lblI+1] = vindex

		}
	}
	return nil
}
