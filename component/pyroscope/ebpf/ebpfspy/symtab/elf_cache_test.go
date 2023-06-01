package symtab

import (
	"testing"

	"github.com/prometheus/common/version"

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

func TestSameFileNoBuildID(t *testing.T) {
	if version.GoOS != "linux" {
		t.Skip("same file check relies on file inode, which we check only in linux")
		return
	}
	elfCache, _ := NewElfCache(32)
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
		if res == nil || res.Name != sym.name {
			t.Errorf("failed to resolve from debug elf %v got %v", sym, res)
		}

		res = nobuildid2.Resolve(sym.pc)
		if res == nil || res.Name != sym.name {
			t.Errorf("failed to resolve from stripped elf %v got %v", sym, res)
		}
	}
	if 1 != elfCache.stat2Symbols.Len() {
		t.Errorf("expected a single stat entry in the cache, got %d", elfCache.stat2Symbols.Len())
	}
	if 0 != elfCache.buildID2Symbols.Len() {
		t.Errorf("expected no buildID entris in the cache, got %d", elfCache.buildID2Symbols.Len())
	}
}
