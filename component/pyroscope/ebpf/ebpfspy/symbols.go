//go:build linux

// Package ebpfspy provides integration with Linux eBPF. It is a rough copy of profile.py from BCC tools:
//
//	https://github.com/iovisor/bcc/blob/master/tools/profile.py
package ebpfspy

import (
	"fmt"
	"os"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/metrics"
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/symtab"
	lru "github.com/hashicorp/golang-lru/v2"
)

type symbolCacheEntry struct {
	symbolTable symtab.SymbolTable
	roundNumber int
}
type pidKey uint32

type symbolCache struct {
	pidCache *lru.Cache[pidKey, *symbolCacheEntry]
	elfCache *symtab.ElfCache
	kallsyms symbolCacheEntry
	logger   log.Logger
	metrics  *metrics.Metrics
}

func newSymbolCache(logger log.Logger, options CacheOptions, metrics *metrics.Metrics) (*symbolCache, error) {
	pid2Cache, err := lru.New[pidKey, *symbolCacheEntry](options.PidCacheSize)
	if err != nil {
		return nil, fmt.Errorf("create pid symbol cache %w", err)
	}

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
		logger:   logger,
		metrics:  metrics,
		pidCache: pid2Cache,
		kallsyms: symbolCacheEntry{symbolTable: kallsyms},
		elfCache: elfCache,
	}, nil
}

func (sc *symbolCache) resolve(pid uint32, addr uint64, roundNumber int) symtab.Symbol {
	e := sc.getOrCreateCacheEntry(pidKey(pid))
	staleCheck := false
	if roundNumber != e.roundNumber {
		e.roundNumber = roundNumber
		staleCheck = true
	}
	if staleCheck {
		e.symbolTable.Refresh()
	}
	return e.symbolTable.Resolve(addr)
}

func (sc *symbolCache) getOrCreateCacheEntry(pid pidKey) *symbolCacheEntry {
	if pid == 0 {
		return &sc.kallsyms
	}

	if cache, ok := sc.pidCache.Get(pid); ok {
		sc.metrics.PidCacheHit.Inc()
		return cache
	}
	sc.metrics.PidCacheMiss.Inc()

	symbolTable := symtab.NewProcTable(sc.logger, symtab.ProcTableOptions{
		Pid: int(pid),
		ElfTableOptions: symtab.ElfTableOptions{
			UseDebugFiles: true,
			ElfCache:      sc.elfCache,
		},
	})
	e := &symbolCacheEntry{symbolTable: symbolTable, roundNumber: -1}
	sc.pidCache.Add(pid, e)
	return e
}

func (sc *symbolCache) updateOptions(options CacheOptions) {
	sc.pidCache.Resize(options.PidCacheSize)
	sc.elfCache.Resize(options.ElfCacheSize)
}
