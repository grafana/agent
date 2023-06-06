package symtab

import (
	"testing"

	"github.com/grafana/agent/pkg/util"
)

func TestElfCacheStrippedEmpty(t *testing.T) {
	logger := util.TestLogger(t)
	elfCache, _ := NewElfCache(32)
	stripped, err := NewElfTable(logger, ".", "testdata/elfs/elf.stripped",
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
		res := stripped.Resolve(sym.pc)
		if res != nil {
			t.Errorf("broken stripped elf ")
		}
	}
}

func TestElfCache(t *testing.T) {
	elfCache, _ := NewElfCache(32)
	logger := util.TestLogger(t)
	debug, err := NewElfTable(logger, ".", "testdata/elfs/elf.debug",
		ElfTableOptions{
			UseDebugFiles: false,
			ElfCache:      elfCache,
		})
	if err != nil {
		t.Fatal(err)
	}
	stripped, err := NewElfTable(logger, ".", "testdata/elfs/elf.stripped",
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
		res := debug.Resolve(sym.pc)
		if res == nil || res.Name != sym.name {
			t.Errorf("failed to resolve from debug elf %v got %v", sym, res)
		}

		res = stripped.Resolve(sym.pc)
		if res == nil || res.Name != sym.name {
			t.Errorf("failed to resolve from stripped elf %v got %v", sym, res)
		}
	}
}
