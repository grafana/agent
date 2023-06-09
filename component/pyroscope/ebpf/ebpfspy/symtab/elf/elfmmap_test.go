package elf

import (
	"debug/elf"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

type Sym struct { // dup to break import cycle
	Start uint64
	Name  string
}

func Test(t *testing.T) {
	fs := []string{
		"../testdata/elfs/elf",
		"../testdata/elfs/libexample.so",
	}
	for _, f := range fs {
		testOneElfFile(t, f)
	}

}

func testOneElfFile(t *testing.T, f string) {
	e, err := elf.Open(f)
	require.NoError(t, err)
	defer e.Close()

	if err != nil {
		fmt.Println(err)
		return
	}
	var genuineSymbols []Sym
	symbols, _ := e.Symbols()
	dynSymbols, _ := e.DynamicSymbols()
	namesLength := 0
	count := 0
	for _, symbol := range symbols {
		if symbol.Value != 0 && symbol.Info&0xf == byte(elf.STT_FUNC) {
			namesLength += len(symbol.Name)
			count += 1
			genuineSymbols = append(genuineSymbols, Sym{
				Name:  symbol.Name,
				Start: symbol.Value,
			})
		}
	}
	for _, symbol := range dynSymbols {
		if symbol.Value != 0 && symbol.Info&0xf == byte(elf.STT_FUNC) {
			namesLength += len(symbol.Name)
			count += 1
			genuineSymbols = append(genuineSymbols, Sym{
				Name:  symbol.Name,
				Start: symbol.Value,
			})
		}
	}
	fmt.Printf("%s names len %d cnt %d\n", f, namesLength, count)

	me, err := NewMMapedElfFile(f)
	require.NoError(t, err)
	defer me.Close()

	tab, _ := me.NewSymbolTable()
	if tab == nil {
		tab = &SymbolTable{}
	}
	var mySymbols []Sym

	for i := range tab.Symbols {
		sym := &tab.Symbols[i]
		name, _ := tab.symbolName(sym)
		mySymbols = append(mySymbols, Sym{
			Name:  name,
			Start: sym.Value,
		})
	}

	cmp := func(a, b Sym) bool {
		if a.Start == b.Start {
			return strings.Compare(a.Name, b.Name) < 0
		}
		return a.Start < b.Start
	}
	slices.SortFunc(genuineSymbols, cmp)
	slices.SortFunc(mySymbols, cmp)
	require.Equal(t, genuineSymbols, mySymbols)

}
