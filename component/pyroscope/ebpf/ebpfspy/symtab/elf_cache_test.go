package symtab

import (
	"testing"

	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/metrics"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestElfCacheStrippedEmpty(t *testing.T) {
	logger := util.TestLogger(t)
	elfCache, _ := NewElfCache(32, metrics.NewMetrics(nil))
	fs := "." // make it unable to find debug file by buildID
	stripped := NewElfTable(logger, &ProcMap{StartAddr: 0x1000, Offset: 0x1000}, fs, "elf/testdata/elfs/elf.stripped",
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
		res := stripped.Resolve(sym.pc)
		require.Error(t, stripped.err)
		require.Equal(t, "", res)
	}
}

func TestElfCache(t *testing.T) {
	elfCache, _ := NewElfCache(32, metrics.NewMetrics(nil))
	logger := util.TestLogger(t)
	debug := NewElfTable(logger, &ProcMap{StartAddr: 0x1000, Offset: 0x1000}, ".", "elf/testdata/elfs/elf",
		ElfTableOptions{
			ElfCache: elfCache,
		})

	stripped := NewElfTable(logger, &ProcMap{StartAddr: 0x1000, Offset: 0x1000}, ".", "elf/testdata/elfs/elf.stripped",
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
		res := debug.Resolve(sym.pc)
		require.NoError(t, debug.err)
		require.Equal(t, sym.name, res)
		res = stripped.Resolve(sym.pc)
		require.NoError(t, stripped.err)
		require.Equal(t, sym.name, res)
	}
}
