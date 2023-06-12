package symtab

import (
	"testing"

	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/metrics"
	"github.com/grafana/agent/pkg/util"
)

func TestElf(t *testing.T) {
	elfCache, _ := NewElfCache(32, metrics.NewMetrics(nil))
	logger := util.TestLogger(t)
	tab := NewElfTable(logger, &ProcMap{StartAddr: 0x1000, Offset: 0x1000}, ".", "elf/testdata/elfs/elf",
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
		res := tab.Resolve(sym.pc)
		if res != sym.name {
			t.Errorf("failed to resolv %v got %v", sym, res)
		}
	}
}
