package symtab

type SymbolTable interface {
	Refresh()
	Resolve(addr uint64) Symbol
}
