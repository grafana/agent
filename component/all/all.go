// Package all imports all known component packages.
package all

import (
	_ "github.com/grafana/agent/component/discovery/kubernetes"       // Import discovery.k8s
	_ "github.com/grafana/agent/component/integrations/node_exporter" // Import targets.mutate
	_ "github.com/grafana/agent/component/local/file"                 // Import local.file
	_ "github.com/grafana/agent/component/metrics/mutate"             // Import metrics.mutate
	_ "github.com/grafana/agent/component/metrics/remotewrite"        // Import metrics.remotewrite
	_ "github.com/grafana/agent/component/metrics/scrape"             // Import metrics.scrape
	_ "github.com/grafana/agent/component/remote/s3"                  // Import s3.file
	_ "github.com/grafana/agent/component/targets/mutate"             // Import targets.mutate
)
