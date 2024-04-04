// Package install registers all in-source integrations for use.
package install

import (
	//
	// v1 integrations
	//

	_ "github.com/grafana/agent/internal/static/integrations/agent"                  // register agent
	_ "github.com/grafana/agent/internal/static/integrations/blackbox_exporter"      // register blackbox_exporter
	_ "github.com/grafana/agent/internal/static/integrations/consul_exporter"        // register consul_exporter
	_ "github.com/grafana/agent/internal/static/integrations/elasticsearch_exporter" // register elasticsearch_exporter
	_ "github.com/grafana/agent/internal/static/integrations/node_exporter"          // register node_exporter
	_ "github.com/grafana/agent/internal/static/integrations/windows_exporter"       // register windows_exporter

	//
	// v2 integrations
	//

	_ "github.com/grafana/agent/internal/static/integrations/v2/agent"             // register agent
	_ "github.com/grafana/agent/internal/static/integrations/v2/apache_http"       // register apache_exporter
	_ "github.com/grafana/agent/internal/static/integrations/v2/blackbox_exporter" // register blackbox_exporter
	_ "github.com/grafana/agent/internal/static/integrations/v2/snmp_exporter"     // register snmp_exporter
)
