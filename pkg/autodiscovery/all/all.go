// Package all imports all known autodiscovery mechanisms.
package all

import (
	_ "github.com/grafana/agent/pkg/autodiscovery/docker" // Import autodiscovery.docker
	_ "github.com/grafana/agent/pkg/autodiscovery/mysql"  // Import autodiscovery.mysql
)
