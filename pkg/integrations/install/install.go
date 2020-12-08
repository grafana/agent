// Package install registers all in-source integrations for use.
package install

import (
	_ "github.com/grafana/agent/pkg/integrations/agent"
	_ "github.com/grafana/agent/pkg/integrations/consul_exporter"
	_ "github.com/grafana/agent/pkg/integrations/dnsmasq_exporter"
	_ "github.com/grafana/agent/pkg/integrations/memcached_exporter"
	_ "github.com/grafana/agent/pkg/integrations/mysqld_exporter"
	_ "github.com/grafana/agent/pkg/integrations/node_exporter"
	_ "github.com/grafana/agent/pkg/integrations/postgres_exporter"
	_ "github.com/grafana/agent/pkg/integrations/process_exporter"
	_ "github.com/grafana/agent/pkg/integrations/redis_exporter"
	_ "github.com/grafana/agent/pkg/integrations/statsd_exporter"
)
