// Package install registers all in-source integrations for use.
package install

import (
	//
	// v1 integrations
	//

	_ "github.com/grafana/agent/pkg/integrations/agent"                  // register agent
	_ "github.com/grafana/agent/pkg/integrations/cadvisor"               // register cadvisor
	_ "github.com/grafana/agent/pkg/integrations/consul_exporter"        // register consul_exporter
	_ "github.com/grafana/agent/pkg/integrations/dnsmasq_exporter"       // register dnsmasq_exporter
	_ "github.com/grafana/agent/pkg/integrations/elasticsearch_exporter" // register elasticsearch_exporter
	_ "github.com/grafana/agent/pkg/integrations/github_exporter"        // register github_exporter
	_ "github.com/grafana/agent/pkg/integrations/kafka_exporter"         // register kafka_exporter
	_ "github.com/grafana/agent/pkg/integrations/memcached_exporter"     // register memcached_exporter
	_ "github.com/grafana/agent/pkg/integrations/mongodb_exporter"       // register mongodb_exporter
	_ "github.com/grafana/agent/pkg/integrations/mysqld_exporter"        // register mysqld_exporter
	_ "github.com/grafana/agent/pkg/integrations/node_exporter"          // register node_exporter
	_ "github.com/grafana/agent/pkg/integrations/postgres_exporter"      // register postgres_exporter
	_ "github.com/grafana/agent/pkg/integrations/process_exporter"       // register process_exporter
	_ "github.com/grafana/agent/pkg/integrations/redis_exporter"         // register redis_exporter
	_ "github.com/grafana/agent/pkg/integrations/snmp_exporter"          // register snmp_exporter
	_ "github.com/grafana/agent/pkg/integrations/statsd_exporter"        // register statsd_exporter
	_ "github.com/grafana/agent/pkg/integrations/windows_exporter"       // register windows_exporter

	//
	// v2 integrations
	//

	_ "github.com/grafana/agent/pkg/integrations/v2/agent" // register agent
	_ "github.com/grafana/agent/pkg/integrations/v2/eventhandler"
	_ "github.com/grafana/agent/pkg/integrations/v2/snmp_exporter"
)
