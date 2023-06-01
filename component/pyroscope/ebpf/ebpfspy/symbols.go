//go:build linux

// Package ebpfspy provides integration with Linux eBPF. It is a rough copy of profile.py from BCC tools:
//
//	https://github.com/iovisor/bcc/blob/master/tools/profile.py
package ebpfspy

import (
	"fmt"
	"os"

	"github.com/go-kit/log"
	symtab2 "github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/symtab"
	"github.com/pyroscope-io/pyroscope/pkg/util/genericlru"
)

type symbolCacheEntry struct {
	symbolTable symtab2.SymbolTable
	roundNumber int
}
type pidKey uint32

type symbolCache struct {
	pid2Cache *genericlru.GenericLRU[pidKey, symbolCacheEntry]
	kallsyms  symbolCacheEntry
	elfCache  *symtab2.ElfCache
	logger    log.Logger
}

func newSymbolCache(logger log.Logger, pidCacheSize int, elfCacheSize int) (*symbolCache, error) {
	pid2Cache, err := genericlru.NewGenericLRU[pidKey, symbolCacheEntry](pidCacheSize, func(pid pidKey, e *symbolCacheEntry) {

	})
	if err != nil {
		return nil, fmt.Errorf("create pid symbol cache %w", err)
	}

	elfCache, err := symtab2.NewElfCache(elfCacheSize)
	if err != nil {
		return nil, fmt.Errorf("create elf cache %w", err)
	}

	kallsymsData, err := os.ReadFile("/proc/kallsyms")
	if err != nil {
		return nil, fmt.Errorf("read kallsyms %w", err)
	}
	kallsyms, err := symtab2.NewKallsyms(kallsymsData)
	if err != nil {
		return nil, fmt.Errorf("create kallsyms %w ", err)
	}
	return &symbolCache{
		logger:    logger,
		pid2Cache: pid2Cache,
		kallsyms:  symbolCacheEntry{symbolTable: kallsyms},
		elfCache:  elfCache,
	}, nil
}

func (sc *symbolCache) resolve(pid uint32, addr uint64, roundNumber int) *symtab2.Symbol {
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

	if cache, ok := sc.pid2Cache.Get(pid); ok {
		return cache
	}

	symbolTable := symtab2.NewProcTable(sc.logger, symtab2.ProcTableOptions{
		Pid: int(pid),
		ElfTableOptions: symtab2.ElfTableOptions{
			UseDebugFiles: true,
			ElfCache:      sc.elfCache,
		},
	})
	e := &symbolCacheEntry{symbolTable: symbolTable}
	sc.pid2Cache.Add(pid, e)
	return e
}
