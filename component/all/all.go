// Package all imports all known component packages.
package all

import (
	_ "github.com/grafana/agent/component/local/file"          // Import local.file
	_ "github.com/grafana/agent/component/metrics/mutate"      // Import metrics.mutate
	_ "github.com/grafana/agent/component/metrics/remotewrite" // Import metrics.remotewrite
	_ "github.com/grafana/agent/component/metrics/scrape"      // Import metrics.scrape
	_ "github.com/grafana/agent/component/s3/file"             // Import s3.file
	_ "github.com/grafana/agent/component/targets/mutate"      // Import targets.mutate
)
