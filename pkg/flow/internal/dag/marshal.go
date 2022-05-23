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
	for _, edge := range sortedEdges(g.Edges()) {
		fmt.Fprintf(&buf, "\t%q -> %q\n", edge.From.NodeID(), edge.To.NodeID())
	}

	fmt.Fprintf(&buf, "}")
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

func sortedEdges(edge []Edge) []Edge {
	res := make([]Edge, len(edge))
	copy(res, edge)

	sort.Slice(res, func(i, j int) bool {
		var (
			fromNodeI = res[i].From.NodeID()
			fromNodeJ = res[j].From.NodeID()

			toNodeI = res[i].To.NodeID()
			toNodeJ = res[j].To.NodeID()
		)

		// Sort first by from nodes, then by to nodes
		if fromNodeI != fromNodeJ {
			return fromNodeI < fromNodeJ
		}
		return toNodeI < toNodeJ
	})

	return res
}
