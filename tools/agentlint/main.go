// Command agentlint provides custom linting utilities for the grafana/agent
// repo.
package main

import (
	"github.com/grafana/agent/tools/agentlint/internal/findcomponents"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(findcomponents.Analyzer)
}
