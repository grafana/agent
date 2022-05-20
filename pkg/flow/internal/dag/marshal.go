package dag

import (
	"bytes"
	"fmt"
	"sort"
)

// MarshalDOT marshals g into the Graphviz DOT format.
func MarshalDOT(g *Graph) []byte {
	var buf bytes.Buffer

	fmt.Fprintln(&buf, "digraph {")
	fmt.Fprintf(&buf, "\trankdir=%q\n", "LR")

	fmt.Fprintf(&buf, "\n\t// Vertices:\n")
	for _, n := range sortedNodeNames(g.Nodes()) {
		fmt.Fprintf(&buf, "\t%q\n", n)
	}

	fmt.Fprintf(&buf, "\n\t// Edges:\n")
	for _, edge := range g.Edges() {
		fmt.Fprintf(&buf, "\t%q -> %q\n", edge.From.NodeID(), edge.To.NodeID())
	}

	fmt.Fprintln(&buf, "}")
	return buf.Bytes()
}

func sortedNodeNames(nn []Node) []string {
	names := make([]string, len(nn))
	for i, n := range nn {
		names[i] = n.NodeID()
	}
	sort.Strings(names)
	return names
}
