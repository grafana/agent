package elf

import (
	"debug/elf"
	"strings"

	"golang.org/x/exp/slices"
)

type TestSym struct {
	Name  string
	Start uint64
}

func GetELFSymbolsFromSymtab(elfFile *elf.File) []TestSym {
	symtab, _ := elfFile.Symbols()
	dynsym, _ := elfFile.DynamicSymbols()
	var symbols []TestSym
	add := func(t []elf.Symbol) {
		for _, sym := range t {
			if sym.Value != 0 && sym.Info&0xf == byte(elf.STT_FUNC) {
				symbols = append(symbols, TestSym{
					Name:  sym.Name,
					Start: sym.Value,
				})
			}
		}
	}
	add(symtab)
	add(dynsym)
	slices.SortFunc(symbols, func(a, b TestSym) bool {
		if a.Start == b.Start {
			return strings.Compare(a.Name, b.Name) < 0
		}
		return a.Start < b.Start
	})
	return symbols
}
