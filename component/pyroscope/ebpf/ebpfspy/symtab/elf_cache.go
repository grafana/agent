package symtab

import (
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/metrics"
	lru "github.com/hashicorp/golang-lru/v2"
)

type ElfCache struct {
	buildID2Symbols *lru.Cache[string, SymbolNameResolver]
	stat2Symbols    *lru.Cache[stat, SymbolNameResolver]
	metrics         *metrics.Metrics
}

func NewElfCache(sz int, metrics *metrics.Metrics) (*ElfCache, error) {
	buildID2Symbols, err := lru.New[string, SymbolNameResolver](sz)
	if err != nil {
		return nil, err
	}
	stat2Symbols, err := lru.New[stat, SymbolNameResolver](sz)
	if err != nil {
		return nil, err
	}
	return &ElfCache{
		buildID2Symbols: buildID2Symbols,
		stat2Symbols:    stat2Symbols,
		metrics:         metrics,
	}, nil
}

func (e *ElfCache) GetSymbolsByBuildID(buildID string) SymbolNameResolver {
	if buildID == "" {
		return nil
	}
	entry, ok := e.buildID2Symbols.Get(buildID)
	if ok && entry != nil {
		e.metrics.ElfCacheBuildIDHit.Inc()
		return entry
	}
	e.metrics.ElfCacheBuildIDMiss.Inc()
	return nil
}

func (e *ElfCache) GetSymbolsByStat(s stat) SymbolNameResolver {
	if s.isNil() {
		return nil
	}
	entry, ok := e.stat2Symbols.Get(s)
	if ok && entry != nil {
		e.metrics.ElfCacheStatHit.Inc()
		return entry
	}
	e.metrics.ElfCacheStatMiss.Inc()
	return nil
}

func (e *ElfCache) CacheByBuildID(buildID string, v SymbolNameResolver) {
	if buildID == "" || v == nil {
		return
	}
	e.buildID2Symbols.Add(buildID, v)
}

func (e *ElfCache) CacheByStat(s stat, v SymbolNameResolver) {
	if s.isNil() || v == nil {
		return
	}
	e.stat2Symbols.Add(s, v)
}

func (e *ElfCache) Resize(size int) {
	e.stat2Symbols.Resize(size)
	e.buildID2Symbols.Resize(size)
}

func (e *ElfCache) Cleanup() {
	cleanup(e.buildID2Symbols)
	cleanup(e.stat2Symbols)
}

func cleanup[k comparable](m *lru.Cache[k, SymbolNameResolver]) {
	keys := m.Keys()
	for _, pid := range keys {
		tab, ok := m.Peek(pid)
		if !ok || tab == nil {
			continue
		}
		tab.Cleanup()
	}
}

func (s *stat) isNil() bool {
	return s.dev == 0 && s.ino == 0
}
