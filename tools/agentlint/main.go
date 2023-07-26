// Command agentlint provides custom linting utilities for the grafana/agent
// repo.
package main

import (
	"github.com/grafana/agent/tools/agentlint/internal/findcomponents"
	"github.com/grafana/agent/tools/agentlint/internal/rivertags"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(
		findcomponents.Analyzer,
		rivertags.Analyzer,
	)
}
