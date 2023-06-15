package symtab

import (
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/metrics"
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/symtab/elf"
)

type ElfCache struct {
	BuildIDCache  *GCache[elf.BuildID, SymbolNameResolver]
	SameFileCache *GCache[stat, SymbolNameResolver]

	metrics *metrics.Metrics
}

func NewElfCache(buildIDCacheOptions GCacheOptions, sameFileCacheOptions GCacheOptions, metrics *metrics.Metrics) (*ElfCache, error) {
	buildIdCache, err := NewGCache[elf.BuildID, SymbolNameResolver](buildIDCacheOptions)
	if err != nil {
		return nil, err
	}

	statCache, err := NewGCache[stat, SymbolNameResolver](sameFileCacheOptions)
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

func (e *ElfCache) Update(buildIDCacheOptions GCacheOptions, sameFileCacheOptions GCacheOptions) {
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
