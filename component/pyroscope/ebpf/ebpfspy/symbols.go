//go:build linux

// Package ebpfspy provides integration with Linux eBPF. It is a rough copy of profile.py from BCC tools:
//
//	https://github.com/iovisor/bcc/blob/master/tools/profile.py
package ebpfspy

import (
	"fmt"
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/symtab/gcache"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/metrics"
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/symtab"
)

type pidKey uint32

type symbolCache struct {
	pidCache *gcache.GCache[pidKey, symtab.SymbolTable]

	elfCache *symtab.ElfCache
	kallsyms symtab.SymbolTable
	logger   log.Logger
	metrics  *metrics.Metrics
}

func newSymbolCache(logger log.Logger, options CacheOptions, metrics *metrics.Metrics) (*symbolCache, error) {
	elfCache, err := symtab.NewElfCache(options.BuildIDCacheOptions, options.SameFileCacheOptions, metrics)
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
	cache, err := gcache.NewGCache[pidKey, symtab.SymbolTable](options.PidCacheOptions)
	if err != nil {
		return nil, fmt.Errorf("create pid cache %w", err)
	}
	return &symbolCache{
		logger:   logger,
		metrics:  metrics,
		pidCache: cache,
		kallsyms: kallsyms,
		elfCache: elfCache,
	}, nil
}

func (sc *symbolCache) NextRound() {
	sc.pidCache.NextRound()
	sc.elfCache.NextRound()
}

func (sc *symbolCache) resolve(pid uint32, addr uint64) symtab.Symbol {
	e := sc.getOrCreateCacheEntry(pidKey(pid))
	return e.Resolve(addr)
}

func (sc *symbolCache) Cleanup() {
	sc.elfCache.Cleanup()

	sc.pidCache.Cleanup()
	level.Debug(sc.logger).Log("buildIdCache", sc.elfCache.BuildIDCache.DebugString())
	level.Debug(sc.logger).Log("sameFileCache", sc.elfCache.SameFileCache.DebugString())
	level.Debug(sc.logger).Log("pidCache", sc.pidCache.DebugString())

}

func (sc *symbolCache) getOrCreateCacheEntry(pid pidKey) symtab.SymbolTable {
	if pid == 0 {
		return sc.kallsyms
	}
	cached := sc.pidCache.Get(pid)
	if cached != nil {
		return cached
	}

	level.Debug(sc.logger).Log("msg", "NewProcTable", "pid", pid)
	fresh := symtab.NewProcTable(sc.logger, symtab.ProcTableOptions{
		Pid: int(pid),
		ElfTableOptions: symtab.ElfTableOptions{
			ElfCache: sc.elfCache,
		},
	})

	sc.pidCache.Cache(pid, fresh)
	return fresh
}

func (sc *symbolCache) updateOptions(options CacheOptions) {
	sc.pidCache.Update(options.PidCacheOptions)
	sc.elfCache.Update(options.BuildIDCacheOptions, options.SameFileCacheOptions)
}
