// Package all imports all known component packages.
package all

import (
	_ "github.com/grafana/agent/component/local/file"        // Import local.file
	_ "github.com/grafana/agent/component/otel/otlpreceiver" // Import otel.receiver_otlp
	_ "github.com/grafana/agent/component/targets/mutate"    // Import targets.mutate
)
