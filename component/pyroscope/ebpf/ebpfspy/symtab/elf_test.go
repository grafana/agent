package symtab

import (
	"testing"

	"github.com/grafana/agent/pkg/util"
)

func TestElf(t *testing.T) {
	elfCache, _ := NewElfCache(32)
	logger := util.TestLogger(t)
	tab, err := NewElfTable(logger, ".", "testdata/elfs/elf",
		ElfTableOptions{UseDebugFiles: false, ElfCache: elfCache})

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
		res := tab.Resolve(sym.pc)
		if res == nil || res.Name != sym.name {
			t.Errorf("failed to resolv %v got %v", sym, res)
		}
	}
}
