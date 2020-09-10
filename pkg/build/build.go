package build

import (
	// We do an empty import of Loki's build package to force it ahead of us on
	// the dependency graph. This makes sure that our init function runs after it
	// and retains the build info we set at compile time.
	_ "github.com/grafana/loki/pkg/build"

	"github.com/prometheus/common/version"
)

// Version information passed to Prometheus version package.
// Package path as used by linker changes based on vendoring being used or not,
// so it's easier just to use stable Agent path, and pass it to
// Prometheus in the code.
var (
	Version   string
	Revision  string
	Branch    string
	BuildUser string
	BuildDate string
)

func init() {
	version.Version = Version
	version.Revision = Revision
	version.Branch = Branch
	version.BuildUser = BuildUser
	version.BuildDate = BuildDate
}
