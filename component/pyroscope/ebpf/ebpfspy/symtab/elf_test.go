package symtab

import (
	"testing"

	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/metrics"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestElf(t *testing.T) {
	elfCache, _ := NewElfCache(testCacheOptions, testCacheOptions, metrics.NewMetrics(nil))
	logger := util.TestLogger(t)
	tab := NewElfTable(logger, &ProcMap{StartAddr: 0x1000, Offset: 0x1000}, ".", "elf/testdata/elfs/elf",
		ElfTableOptions{
			ElfCache: elfCache,
		})

	syms := []struct {
		name string
		pc   uint64
	}{
		{"", 0x0},
		{"iter", 0x1149},
		{"main", 0x115e},
	}
	for _, sym := range syms {
		res := tab.Resolve(sym.pc)
		require.Equal(t, res, sym.name)

	}
}
