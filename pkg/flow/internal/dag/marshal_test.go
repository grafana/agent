package dag_test

import (
	"testing"

	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/stretchr/testify/require"
)

func TestMarshalDOT(t *testing.T) {
	var (
		nodeA = stringNode("a")
		nodeB = stringNode("b")
		nodeC = stringNode("c")
	)

	var g dag.Graph
	g.Add(nodeA)
	g.Add(nodeB)
	g.Add(nodeC)

	g.AddEdge(dag.Edge{From: nodeA, To: nodeB})
	g.AddEdge(dag.Edge{From: nodeA, To: nodeC})

	expect := `digraph {
	rankdir="LR"

	// Vertices:
	"a"
	"b"
	"c"

	// Edges:
	"a" -> "b"
	"a" -> "c"
}`

	marshaled := dag.MarshalDOT(&g)
	require.Equal(t, expect, string(marshaled))
}

type stringNode string

func (s stringNode) NodeID() string { return string(s) }
