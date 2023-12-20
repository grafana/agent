package controller

import "github.com/grafana/agent/pkg/flow/internal/dag"

// ModuleContentProvider is used by import and declare nodes to provide module content for the declare component nodes.
type ModuleContentProvider interface {
	dag.Node

	// ModuleContent returns the content of a given module.
	ModuleContent(string) (string, error)
}
