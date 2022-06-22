// Package all imports all known component packages.
package all

import (
	_ "github.com/grafana/agent/component/integrations/node_exporter"
	_ "github.com/grafana/agent/component/local/file"     // Import local.file
	_ "github.com/grafana/agent/component/targets/mutate" // Import targets.mutate
)
