package mysql

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/autodiscovery"
	"github.com/grafana/agent/pkg/river"
	"github.com/mitchellh/go-ps"
)

// Config holds the predefined DSNs we will try to connect to a running MySQL
// instance with.
type Config struct {
	Binary     string   `river:"binary,attr"`
	DSN        []string `river:"dsn,attr,optional"`
	Extensions []string `river:"ext,attr,optional"`
}

// MySQL is an autodiscovery mechanism for MySQL-compatible databases.
type MySQL struct {
	binary string
	dsn    []string
	ext    []string
}

// New creates a new auto-discovery MySQL mechanism instance.
func New() (*MySQL, error) {
	bb, err := os.ReadFile("mysql.river")
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = river.Unmarshal(bb, &cfg)
	if err != nil {
		return nil, err
	}

	return &MySQL{
		binary: cfg.Binary,
		dsn:    cfg.DSN,
		ext:    cfg.Extensions,
	}, nil
}

// Run check whether a MySQL instance is running, and if so, returns a
// `prometheus.exporter.mysql` component that can read metrics from it.
func (m *MySQL) Run() (*autodiscovery.Result, error) {
	procs, err := ps.Processes()
	if err != nil {
		return nil, fmt.Errorf("could not read processes from host system: %w", err)
	}

	var mysqlPID int
	for _, p := range procs {
		if p.Executable() == m.binary {
			mysqlPID = p.Pid()
			break
		}
	}
	if mysqlPID == 0 {
		return nil, fmt.Errorf("no running mysql instance was found")
	}

	// MySQL is running on the host system, so we'll try to return _something_.
	res := &autodiscovery.Result{}
	lsof := autodiscovery.LSOF{}

	fns, err := autodiscovery.GetOpenFilenames(lsof, mysqlPID, m.ext...)
	if err != nil {
		return nil, err
	}
	for fn, _ := range fns {
		res.LogfileTargets = append(res.LogfileTargets,
			discovery.Target{"__path__": fn, "component": "postgres"},
		)
	}

	// Let's try to use the configuration to connect using predefined DSNs.
	for _, dsn := range m.dsn {
		db, err := sql.Open("mysql", dsn)
		defer db.Close()
		if err != nil {
			continue
		} else {
			res.RiverConfig = fmt.Sprintf(`prometheus.exporter.mysql "default" {
  data_source_name = "%s"
}`, dsn)
			res.MetricsExport = "prometheus.exporter.mysql.default.targets"
			return res, nil
		}
	}

	// Our predefined configurations didn't work; but MySQL is running.
	// Let's return a Flow component template for the user to fill out.
	res.RiverConfig = `prometheus.exporter.mysql "default" {
  data_source_name = env("AGENT_MYSQL_DSN")
}`
	res.MetricsExport = "prometheus.exporter.mysql.default.targets"

	return res, nil
}
