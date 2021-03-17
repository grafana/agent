// Package install registers all in-source integrations for use.
package install

import (
	_ "github.com/grafana/agent/pkg/integrations/agent"                  // register agent
	_ "github.com/grafana/agent/pkg/integrations/consul_exporter"        // register consul_exporter
	_ "github.com/grafana/agent/pkg/integrations/dnsmasq_exporter"       // register dnsmasq_exporter
	_ "github.com/grafana/agent/pkg/integrations/elasticsearch_exporter" // register elasticsearch_exporter
	_ "github.com/grafana/agent/pkg/integrations/memcached_exporter"     // register memcached_exporter
	_ "github.com/grafana/agent/pkg/integrations/mysqld_exporter"        // register mysqld_exporter
	_ "github.com/grafana/agent/pkg/integrations/node_exporter"          // register node_exporter
	_ "github.com/grafana/agent/pkg/integrations/postgres_exporter"      // register postgres_exporter
	_ "github.com/grafana/agent/pkg/integrations/process_exporter"       // register process_exporter
	_ "github.com/grafana/agent/pkg/integrations/redis_exporter"         // register redis_exporter
	_ "github.com/grafana/agent/pkg/integrations/statsd_exporter"        // register statsd_exporter
	_ "github.com/prometheus-community/windows_exporter/exporter"        // register windows_exporter
)
