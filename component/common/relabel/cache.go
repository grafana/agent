package relabel

import (
	"fmt"
	"sync"

	"github.com/grafana/agent/service/labelstore"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/prometheus/client_golang/prometheus"
	prometheus_client "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/prometheus/prometheus/model/value"
)

type Cache struct {
	cacheHits   prometheus_client.Counter
	cacheMisses prometheus_client.Counter
	cacheSize   prometheus_client.Gauge
	ls          labelstore.LabelStore

	cacheMut sync.RWMutex
	cache    *lru.Cache[uint64, *labelAndID]
}

// NewCache returns a cache to use with relabelling. maxCacheSize must be a positive number.
// name will be used to label cache metrics in the form `agent_<name>_relabel_cache_hits`.
func NewCache(ls labelstore.LabelStore, maxCacheSize int, name string, prom prometheus.Registerer) (*Cache, error) {
	cache, err := lru.New[uint64, *labelAndID](maxCacheSize)
	if err != nil {
		return nil, err
	}
	c := &Cache{
		cache: cache,
		ls:    ls,
	}

	c.cacheMisses = prometheus_client.NewCounter(prometheus_client.CounterOpts{
		Name: fmt.Sprintf("agent_%s_relabel_cache_misses", name),
		Help: "Total number of cache misses",
	})
	c.cacheHits = prometheus_client.NewCounter(prometheus_client.CounterOpts{
		Name: fmt.Sprintf("agent_%s_relabel_cache_hits", name),
		Help: "Total number of cache hits",
	})
	c.cacheSize = prometheus_client.NewGauge(prometheus_client.GaugeOpts{
		Name: fmt.Sprintf("agent_%s_relabel_cache_size", name),
		Help: "Total size of relabel cache",
	})

	for _, metric := range []prometheus_client.Collector{c.cacheMisses, c.cacheHits, c.cacheSize} {
		err = prom.Register(metric)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Cache) Relabel(val float64, globalRef uint64, lbls labels.Labels, rcs []*relabel.Config) labels.Labels {
	// Only retrieve global ref if necessary.
	if globalRef == 0 {
		globalRef = c.ls.GetOrAddGlobalRefID(lbls)
	}
	var (
		relabelled labels.Labels
		keep       bool
	)
	newLbls, found := c.GetFromCache(globalRef)
	if found {
		c.cacheHits.Inc()
		// If newLbls is nil but cache entry was found then we want to keep the value nil, if it's not we want to reuse the labels
		if newLbls != nil {
			relabelled = newLbls.Labels
		}
	} else {
		// Relabel against a copy of the labels to prevent modifying the original
		// slice.
		relabelled, keep = relabel.Process(lbls.Copy(), rcs...)
		c.cacheMisses.Inc()
		c.AddToCache(globalRef, relabelled, keep)
	}

	// If stale remove from the cache, the reason we don't exit early is so the stale value can propagate.
	// TODO: (@mattdurham) This caching can leak and likely needs a timed eviction at some point, but this is simple.
	// In the future the global ref cache may have some hooks to allow notification of when caches should be evicted.
	if value.IsStaleNaN(val) {
		c.DeleteFromCache(globalRef)
	}

	// Set the cache size to the cache.len
	c.cacheSize.Set(float64(c.cache.Len()))
	return relabelled
}

func (c *Cache) GetFromCache(id uint64) (*labelAndID, bool) {
	c.cacheMut.RLock()
	defer c.cacheMut.RUnlock()

	fm, found := c.cache.Get(id)
	return fm, found
}

func (c *Cache) ClearCache(cacheSize int) {
	c.cacheMut.Lock()
	defer c.cacheMut.Unlock()

	cache, _ := lru.New[uint64, *labelAndID](cacheSize)
	c.cache = cache
}

func (c *Cache) DeleteFromCache(id uint64) {
	c.cacheMut.Lock()
	defer c.cacheMut.Unlock()

	c.cache.Remove(id)
}

func (c *Cache) AddToCache(originalID uint64, lbls labels.Labels, keep bool) {
	c.cacheMut.Lock()
	defer c.cacheMut.Unlock()

	if !keep {
		c.cache.Add(originalID, nil)
		return
	}
	newGlobal := c.ls.GetOrAddGlobalRefID(lbls)
	c.cache.Add(originalID, &labelAndID{
		Labels: lbls,
		ID:     newGlobal,
	})
}

func (c *Cache) Len() int {
	c.cacheMut.RLock()
	defer c.cacheMut.RUnlock()

	return c.cache.Len()
}

// labelAndID stores both the globalrefid for the label and the id itself. We store the id so that it doesn't have
// to be recalculated again.
type labelAndID struct {
	Labels labels.Labels
	ID     uint64
}
