package symtab

import (
	lru "github.com/hashicorp/golang-lru/v2"
)

type ElfCache struct {
	buildID2Symbols *lru.Cache[string, *elfCacheEntry]
	stat2Symbols    *lru.Cache[stat, *elfCacheEntry]
}

type elfCacheEntry struct {
	symbols []Symbol
}

func NewElfCache(sz int) (*ElfCache, error) {
	buildID2Symbols, err := lru.New[string, *elfCacheEntry](sz)
	if err != nil {
		return nil, err
	}
	stat2Symbols, err := lru.New[stat, *elfCacheEntry](sz)
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

func (e *ElfCache) Resize(size int) {
	e.stat2Symbols.Resize(size)
	e.buildID2Symbols.Resize(size)
}

func (s *stat) isNil() bool {
	return s.dev == 0 && s.ino == 0
}
