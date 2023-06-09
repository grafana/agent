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

// "same file check relies on file inode, which we check only in linux"
func TestSameFileNoBuildID(t *testing.T) {
	elfCache, _ := NewElfCache(32, metrics.NewMetrics(nil))
	logger := util.TestLogger(t)
	nobuildid1 := NewElfTable(logger, &ProcMap{StartAddr: 0x1000, Offset: 0x1000}, ".", "testdata/elfs/elf.nobuildid",
		ElfTableOptions{
			ElfCache: elfCache,
		})

	nobuildid2 := NewElfTable(logger, &ProcMap{StartAddr: 0x1000, Offset: 0x1000}, ".", "testdata/elfs/elf.nobuildid",
		ElfTableOptions{
			ElfCache: elfCache,
		})
	syms := []struct {
		name string
		pc   uint64
	}{
		{"iter", 0x1149},
		{"main", 0x115e},
	}
	for _, sym := range syms {
		res := nobuildid1.Resolve(sym.pc)
		require.Equal(t, sym.name, res)
		res = nobuildid2.Resolve(sym.pc)
		require.Equal(t, sym.name, res)
	}
	require.Equal(t, 1, elfCache.stat2Symbols.Len())
	require.Equal(t, 0, elfCache.buildID2Symbols.Len())
}

func TestMallocResolve(t *testing.T) {
	elfCache, _ := NewElfCache(32, metrics.NewMetrics(nil))
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
	gosym, _ := newGoSymbols("/proc/self/exe")
	logger := log.NewSyncLogger(log.NewLogfmtLogger(os.Stderr))
	proc := NewProcTable(logger, ProcTableOptions{Pid: os.Getpid()})
	proc.Refresh()
	if len(gosym.Symbols) < 1000 {
		b.FailNow()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, symbol := range gosym.Symbols {
			proc.Resolve(symbol.Start)
		}
	}
}

func TestSelfElfSymbolsLazy(t *testing.T) {
	f, err := os.Readlink("/proc/self/exe")
	require.NoError(t, err)

	e, err := elf0.Open(f)
	require.NoError(t, err)
	expectedSymbols := getELFSymbolsFromSymtab(e)

	me, err := elf.NewMMapedElfFile(f)
	require.NoError(t, err)

	symbolTable, err := me.NewSymbolTable()
	require.NoError(t, err)

	require.Greater(t, len(symbolTable.Symbols), 1000)

	for j, symbol := range expectedSymbols {
		_ = j
		name := symbolTable.Resolve(symbol.Start)
		if symbol.Name == "runtime.text" && name == "internal/cpu.Initialize" {
			continue
		}
		require.Equal(t, symbol.Name, name)
	}
}
