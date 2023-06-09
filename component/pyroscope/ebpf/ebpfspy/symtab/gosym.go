package symtab

import (
	"debug/elf"
	"debug/gosym"
	"errors"
	"fmt"

	gosym2 "github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/symtab/gosym"
	"golang.org/x/exp/slices"
)

//todo rename, move to gosym

func newGoSymbols(file string) (*SymTab, error) {
	obj, err := elf.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open elf file: %w", err)
	}
	defer obj.Close()

	symbols, err := getGoSymbolsFromPCLN(obj)
	if err != nil {
		return nil, err
	}
	return NewSymTab(symbols), nil
}

func getGoSymbolsFromPCLN(obj *elf.File) ([]Sym, error) {
	//var gosymtab []byte
	var err error
	var pclntab []byte
	text := obj.Section(".text")
	if text == nil {
		return nil, errors.New("empty .text")
	}
	//if sect := obj.Section(".gosymtab"); sect != nil {
	//if gosymtab, err = sect.Data(); err != nil {
	//	return nil, err
	//}
	//}
	if sect := obj.Section(".gopclntab"); sect != nil {
		if pclntab, err = sect.Data(); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("empty .gopclntab")
	}

	textStart := gosym2.ParseRuntimeTextFromPclntab18(pclntab)

	if textStart == 0 {
		// for older versions text.Addr is enough
		// https://github.com/golang/go/commit/b38ab0ac5f78ac03a38052018ff629c03e36b864
		textStart = text.Addr
	}
	if textStart < text.Addr || textStart >= text.Addr+text.Size {
		return nil, fmt.Errorf(" runtime.text out of .text bounds %d %d %d", textStart, text.Addr, text.Size)
	}
	pcln := gosym.NewLineTable(pclntab, textStart)
	table, err := gosym.NewTable(nil, pcln)
	if err != nil {
		return nil, err
	}
	if len(table.Funcs) == 0 {
		return nil, errors.New("gosymtab: no symbols found")
	}

	es := make([]Sym, 0, len(table.Funcs))
	for _, fun := range table.Funcs {
		es = append(es, Sym{Start: fun.Entry, Name: fun.Name})
	}

	slices.SortFunc(es, func(a, b Sym) bool {
		return a.Start < b.Start
	})
	return es, nil
}
