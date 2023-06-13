package symtab

type SymbolTable interface {
	Refresh()
	Cleanup()
	DebugString() string
	Resolve(addr uint64) Symbol
}

type SymbolNameResolver interface {
	Refresh()
	Cleanup()
	DebugString() string
	Resolve(addr uint64) string
}

type noopSymbolNameResolver struct {
}

func (n *noopSymbolNameResolver) Resolve(addr uint64) string {
	return ""
}

func (n *noopSymbolNameResolver) Refresh() {

}
func (n *noopSymbolNameResolver) Cleanup() {

}
func (n *noopSymbolNameResolver) DebugString() string {
	return "noopSymbolNameResolver"
}
