// Package all imports all known component packages.
package all

import (
	_ "github.com/grafana/agent/component/discovery/transformer" // Import discovery.transformer
	_ "github.com/grafana/agent/component/local/file"            // Import local.file
)
