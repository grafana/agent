// Package autodiscovery describes the interfaces which the various mechanisms
// implement.
package autodiscovery

import (
	"github.com/grafana/agent/component/discovery"
)

// Mechanism is the base interface for an autodiscovery mechanism.
// Implementations may also implement extension interfaces (named
// <Extension>Autodiscovery) to implement extra known behavior.
type Mechanism interface {
	Run() (*Result, error)
	String() string
}

// Result ???
type Result struct {
	RiverConfig    string
	MetricsExport  string
	MetricsTargets []discovery.Target
	LogfileTargets []discovery.Target
}
