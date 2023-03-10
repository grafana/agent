package mysql

import (
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/grafana/agent/pkg/autodiscovery"
)

// MySQL ???
type MySQL struct {
	binary string
	dsn    []string
	ext    []string
}

// New creates a new auto-discovery MySQL mechanism instance.
func New() (*MySQL, error) {
	return &MySQL{}, nil
}

// Run check whether a MySQL instance is running, and if so, returns a
// `prometheus.exporter.mysql` component that can read metrics from it.
func (m *MySQL) Run() (*autodiscovery.Result, error) {
	return &autodiscovery.Result{
		RiverConfig: `prometheus.exporter.mysql "default" {
  data_source_name = env("AGENT_MYSQL_DSN")
}`,
		MetricsExport: "prometheus.exporter.mysql.default.targets",
	}, nil
}
