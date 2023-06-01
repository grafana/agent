package symtab

import "github.com/pyroscope-io/pyroscope/pkg/util/genericlru"

type ElfCache struct {
	buildID2Symbols *genericlru.GenericLRU[string, elfCacheEntry]
	stat2Symbols    *genericlru.GenericLRU[stat, elfCacheEntry]
}

type elfCacheEntry struct {
	symbols []Symbol
}

func NewElfCache(sz int) (*ElfCache, error) {
	buildID2Symbols, err := genericlru.NewGenericLRU[string, elfCacheEntry](sz, func(k string, v *elfCacheEntry) {})
	if err != nil {
		return nil, err
	}
	stat2Symbols, err := genericlru.NewGenericLRU[stat, elfCacheEntry](sz, func(k stat, v *elfCacheEntry) {})
	if err != nil {
		return nil, err
	}
	return &ElfCache{
		buildID2Symbols: buildID2Symbols,
		stat2Symbols:    stat2Symbols,
	}, nil
}

func (e *ElfCache) GetSymbolsByBuildID(buildID string) []Symbol {
	if buildID == "" {
		return nil
	}
	entry, ok := e.buildID2Symbols.Get(buildID)
	if ok && entry != nil {
		return entry.symbols
	}
	return nil
}

func (e *ElfCache) GetSymbolsByStat(s stat) []Symbol {
	if s.isNil() {
		return nil
	}
	entry, ok := e.stat2Symbols.Get(s)
	if ok && entry != nil {
		return entry.symbols
	}
	return nil
}

func (e *ElfCache) CacheByBuildID(buildID string, symbols []Symbol) {
	if buildID == "" || len(symbols) == 0 {
		return
	}
	e.buildID2Symbols.Add(buildID, &elfCacheEntry{symbols: symbols})
}

func (e *ElfCache) CacheByStat(s stat, symbols []Symbol) {
	if s.isNil() || len(symbols) == 0 {
		return
	}
	e.stat2Symbols.Add(s, &elfCacheEntry{symbols: symbols})
}

func (s *stat) isNil() bool {
	return s.dev == 0 && s.ino == 0
}
