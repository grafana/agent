//go:build linux

// Package ebpfspy provides integration with Linux eBPF. It is a rough copy of profile.py from BCC tools:
//
//	https://github.com/iovisor/bcc/blob/master/tools/profile.py
package ebpfspy

import (
	"fmt"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/metrics"
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/symtab"
)

type symbolCacheEntry struct {
	symbolTable symtab.SymbolTable
	roundNumber int
}
type pidKey uint32

type symbolCache struct {
	round      int
	roundCache map[pidKey]*symbolCacheEntry
	elfCache   *symtab.ElfCache
	kallsyms   symbolCacheEntry
	logger     log.Logger
	metrics    *metrics.Metrics
}

func newSymbolCache(logger log.Logger, options CacheOptions, metrics *metrics.Metrics) (*symbolCache, error) {
	elfCache, err := symtab.NewElfCache(options.ElfCacheSize, metrics)
	if err != nil {
		return nil, fmt.Errorf("create elf cache %w", err)
	}

	kallsymsData, err := os.ReadFile("/proc/kallsyms")
	if err != nil {
		return nil, fmt.Errorf("read kallsyms %w", err)
	}
	kallsyms, err := symtab.NewKallsyms(kallsymsData)
	if err != nil {
		return nil, fmt.Errorf("create kallsyms %w ", err)
	}
	return &symbolCache{
		logger:     logger,
		metrics:    metrics,
		roundCache: make(map[pidKey]*symbolCacheEntry),
		kallsyms:   symbolCacheEntry{symbolTable: kallsyms},
		elfCache:   elfCache,
	}, nil
}

func (sc *symbolCache) NextRound() {
	sc.round++
}

func (sc *symbolCache) resolve(pid uint32, addr uint64) symtab.Symbol {
	e := sc.getOrCreateCacheEntry(pidKey(pid))
	refresh := false
	if e.roundNumber != sc.round {
		e.roundNumber = sc.round
		refresh = true
	}
	if refresh {
		e.symbolTable.Refresh()
	}
	return e.symbolTable.Resolve(addr)
}

func (sc *symbolCache) Cleanup() {
	sc.elfCache.Cleanup()

	prev := sc.roundCache
	for _, entry := range prev {
		entry.symbolTable.Cleanup()
	}

	sc.roundCache = make(map[pidKey]*symbolCacheEntry)
	for key, entry := range prev {
		if entry.roundNumber == sc.round {
			sc.roundCache[key] = entry
		} else {
			level.Debug(sc.logger).Log("msg", "symbolCache removing pid",
				"pid", key,
				"now", entry.roundNumber)
		}
	}
	level.Debug(sc.logger).Log("msg", "symbolCache cleanup", "was", len(prev), "now", len(sc.roundCache))
}

func (sc *symbolCache) getOrCreateCacheEntry(pid pidKey) *symbolCacheEntry {
	if pid == 0 {
		return &sc.kallsyms
	}

	if cache, ok := sc.roundCache[pid]; ok {
		return cache
	}

	symbolTable := symtab.NewProcTable(sc.logger, symtab.ProcTableOptions{
		Pid: int(pid),
		ElfTableOptions: symtab.ElfTableOptions{
			ElfCache: sc.elfCache,
		},
	})
	e := &symbolCacheEntry{symbolTable: symbolTable, roundNumber: -1}

	sc.roundCache[pid] = e
	return e
}

func (sc *symbolCache) updateOptions(options CacheOptions) {
	//sc.pidCache.Resize(options.PidCacheSize)
	sc.elfCache.Resize(options.ElfCacheSize)
}
