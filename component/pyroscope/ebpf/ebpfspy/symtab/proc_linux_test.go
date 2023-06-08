//go:build linux

package symtab

import (
	"os"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/require"

	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/metrics"
	"github.com/grafana/agent/pkg/util"
)

// "same file check relies on file inode, which we check only in linux"
func TestSameFileNoBuildID(t *testing.T) {
	elfCache, _ := NewElfCache(32, metrics.NewMetrics(nil))
	logger := util.TestLogger(t)
	nobuildid1, err := NewElfTable(logger, ".", "testdata/elfs/elf.nobuildid",
		ElfTableOptions{
			UseDebugFiles: false,
			ElfCache:      elfCache,
		})
	if err != nil {
		t.Fatal(err)
	}
	nobuildid2, err := NewElfTable(logger, ".", "testdata/elfs/elf.nobuildid",
		ElfTableOptions{
			UseDebugFiles: false,
			ElfCache:      elfCache,
		})
	if err != nil {
		t.Fatal(err)
	}

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
			UseDebugFiles: false,
			ElfCache:      elfCache,
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
	if len(gosym.symbols) < 1000 {
		b.FailNow()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, symbol := range gosym.symbols {
			proc.Resolve(symbol.Start)
		}
	}
}
