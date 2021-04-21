package contextkeys

type key int

// Loki is a constant used to pass *loki.Loki through the context
const Loki key = iota
