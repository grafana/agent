package elf

import (
	"debug/elf"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestElfSymbolComparison(t *testing.T) {

	testOneElfFile := func(t *testing.T, f string) {
		e, err := elf.Open(f)
		require.NoError(t, err)
		defer e.Close()

		if err != nil {
			fmt.Println(err)
			return
		}
		genuineSymbols := GetELFSymbolsFromSymtab(e)

		me, err := NewMMapedElfFile(f)
		require.NoError(t, err)
		defer me.Close()

		tab, _ := me.NewSymbolTable()
		if tab == nil {
			tab = &SymbolTable{}
		}
		var mySymbols []TestSym

		for i := range tab.Symbols {
			sym := &tab.Symbols[i]
			name, _ := tab.symbolName(sym)
			mySymbols = append(mySymbols, TestSym{
				Name:  name,
				Start: sym.Value,
			})
		}

		cmp := func(a, b TestSym) bool {
			if a.Start == b.Start {
				return strings.Compare(a.Name, b.Name) < 0
			}
			return a.Start < b.Start
		}
		slices.SortFunc(mySymbols, cmp)
		slices.SortFunc(genuineSymbols, cmp)
		require.Equal(t, genuineSymbols, mySymbols)

	}

	fs := []string{
		"./testdata/elfs/elf",
		"./testdata/elfs/elf.debug",
		"./testdata/elfs/elf.nopie",
		"./testdata/elfs/libexample.so",
		"./testdata/elfs/go12",
		"./testdata/elfs/go16",
		"./testdata/elfs/go18",
		"./testdata/elfs/go20",
		"./testdata/elfs/go12-static",
		"./testdata/elfs/go16-static",
		"./testdata/elfs/go18-static",
		"./testdata/elfs/go20-static",
	}
	for _, f := range fs {
		testOneElfFile(t, f)
	}

}
