package symtab

type SymbolTable interface {
	Refresh()
	Cleanup()
	Resolve(addr uint64) Symbol
}

type SymbolNameResolver interface {
	Resolve(addr uint64) string
	Cleanup()
}

type noopSymbolNameResolver struct {
}

func (n *noopSymbolNameResolver) Resolve(addr uint64) string {
	return ""
}

func (n *noopSymbolNameResolver) Cleanup() {

}
