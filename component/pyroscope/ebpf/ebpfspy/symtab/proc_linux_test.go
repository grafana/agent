//go:build linux

package symtab

import (
	elf0 "debug/elf"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/require"

	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/metrics"
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/symtab/elf"
	"github.com/grafana/agent/pkg/util"
)

func TestMallocResolve(t *testing.T) {
	elfCache, _ := NewElfCache(testCacheOptions, testCacheOptions, metrics.NewMetrics(nil))
	logger := util.TestLogger(t)
	gosym := NewProcTable(logger, ProcTableOptions{
		Pid: os.Getpid(),
		ElfTableOptions: ElfTableOptions{
			ElfCache: elfCache,
		},
	})
	gosym.Refresh()
	malloc := testHelperGetMalloc()
	res := gosym.Resolve(uint64(malloc))
	require.Contains(t, res.Name, "malloc")
	if !strings.Contains(res.Module, "/libc.so") && !strings.Contains(res.Module, "/libc-") {
		t.Errorf("expected libc, got %v", res.Module)
	}
}

func BenchmarkProc(b *testing.B) {
	gosym, _ := elf.GetGoSymbols("/proc/self/exe")
	logger := log.NewSyncLogger(log.NewLogfmtLogger(os.Stderr))
	proc := NewProcTable(logger, ProcTableOptions{Pid: os.Getpid()})
	proc.Refresh()
	if len(gosym) < 1000 {
		b.FailNow()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, symbol := range gosym {
			proc.Resolve(symbol.Start)
		}
	}
}

func TestSelfElfSymbolsLazy(t *testing.T) {
	f, err := os.Readlink("/proc/self/exe")
	require.NoError(t, err)

	e, err := elf0.Open(f)
	require.NoError(t, err)
	expectedSymbols := elf.GetELFSymbolsFromSymtab(e)

	me, err := elf.NewMMapedElfFile(f)
	require.NoError(t, err)

	symbolTable, err := me.NewSymbolTable()
	require.NoError(t, err)

	require.Greater(t, len(symbolTable.Index.Names), 1000)

	for _, symbol := range expectedSymbols {
		name := symbolTable.Resolve(symbol.Start)
		if symbol.Name == "runtime.text" && name == "internal/cpu.Initialize" {
			continue
		}
		require.Equal(t, symbol.Name, name)
	}
}
