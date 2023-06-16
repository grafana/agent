package symtab

import (
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/metrics"
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/symtab/elf"
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/symtab/gcache"
)

type ElfCache struct {
	BuildIDCache  *gcache.GCache[elf.BuildID, SymbolNameResolver]
	SameFileCache *gcache.GCache[stat, SymbolNameResolver]

	metrics *metrics.Metrics
}

func NewElfCache(buildIDCacheOptions gcache.GCacheOptions, sameFileCacheOptions gcache.GCacheOptions, metrics *metrics.Metrics) (*ElfCache, error) {
	buildIdCache, err := gcache.NewGCache[elf.BuildID, SymbolNameResolver](buildIDCacheOptions)
	if err != nil {
		return nil, err
	}

	statCache, err := gcache.NewGCache[stat, SymbolNameResolver](sameFileCacheOptions)
	if err != nil {
		return nil, err
	}
	return &ElfCache{
		BuildIDCache:  buildIdCache,
		SameFileCache: statCache,

		metrics: metrics,
	}, nil
}

func (e *ElfCache) GetSymbolsByBuildID(buildID elf.BuildID) SymbolNameResolver {
	return e.BuildIDCache.Get(buildID)
}

func (e *ElfCache) CacheByBuildID(buildID elf.BuildID, v SymbolNameResolver) {
	e.BuildIDCache.Cache(buildID, v)
}

func (e *ElfCache) GetSymbolsByStat(s stat) SymbolNameResolver {
	return e.SameFileCache.Get(s)
}

func (e *ElfCache) CacheByStat(s stat, v SymbolNameResolver) {
	e.SameFileCache.Cache(s, v)
}

func (e *ElfCache) Update(buildIDCacheOptions gcache.GCacheOptions, sameFileCacheOptions gcache.GCacheOptions) {
	e.BuildIDCache.Update(buildIDCacheOptions)
	e.SameFileCache.Update(sameFileCacheOptions)
}

func (e *ElfCache) NextRound() {
	e.BuildIDCache.NextRound()
	e.SameFileCache.NextRound()
}

func (e *ElfCache) Cleanup() {
	e.BuildIDCache.Cleanup()
	e.SameFileCache.Cleanup()
}
