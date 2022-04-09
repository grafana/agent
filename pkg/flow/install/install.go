package install

import (
	_ "github.com/grafana/agent/component/discovery/static"   // Import component
	_ "github.com/grafana/agent/component/integration/github" // Import component
	_ "github.com/grafana/agent/component/metrics-forwarder"  // Import component
	_ "github.com/grafana/agent/component/metrics-scraper"    // Import component
	_ "github.com/grafana/agent/component/remote/http"        // Import component
)
