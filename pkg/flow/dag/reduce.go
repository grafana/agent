package dag

// Reduce performs a transitive reduction on g. A transitive reduction removes
// as many edges as possible while maintaining the same "reachability" as the
// original graph: any node N reachable from node S will still be reachable
// after a reduction.
func Reduce(g *Graph) {
	// A direct edge between two vertices can be removed if that same target
	// vertex is indirectly rechable through another edge.
	//
	// To detect this, we iterate through all vertices in the graph, performing a
	// depth-first search at its dependencies. If the target vertex is reachable
	// from the source vertex, the edge is removed.
	for source := range g.nodes {
		_ = Walk(g, g.Dependencies(source), func(direct Node) error {
			// Iterate over (direct, indirect) edges and remove (source, indirect)
			// edges if they exist. This is a safe operaration because other is still
			// reachable by source via its (source, direct) edge.
			for indirect := range g.outEdges[direct] {
				g.RemoveEdge(Edge{From: source, To: indirect})
			}
			return nil
		})
	}
}
