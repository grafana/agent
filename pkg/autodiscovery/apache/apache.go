package apache

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/autodiscovery"
	"github.com/grafana/agent/pkg/river"
	"github.com/mitchellh/go-ps"
)

type Config struct {
	Binary     string   `river:"binary,attr"`
	Extensions []string `river:"ext,attr,optional"`
}

type Apache struct {
	binary string
	ext    []string
}

func New() (*Apache, error) {
	bb, err := os.ReadFile("apache.river")
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = river.Unmarshal(bb, &cfg)
	if err != nil {
		return nil, err
	}

	return &Apache{
		binary: cfg.Binary,
	}, nil
}

// Run check whether a Apache instance is running, and if so, returns a
// `prometheus.exporter.apache` component that can read metrics from it.
func (m *Apache) Run() (*autodiscovery.Result, error) {
	procs, err := ps.Processes()
	if err != nil {
		return nil, fmt.Errorf("could not read processes from host system: %w", err)
	}

	pid := -1
	for _, p := range procs {
		if p.Executable() == m.binary {
			pid = p.Pid()
			break
		}
	}
	if pid == -1 {
		return nil, fmt.Errorf("no running instance of process '%s' was found", m.binary)
	}

	// Apache is running on the host system, so we'll try to return _something_.
	res := &autodiscovery.Result{}
	lsof := autodiscovery.LSOF{}

	fns, err := autodiscovery.GetOpenFilenames(lsof, pid, m.ext...)
	if err != nil {
		return nil, err
	}
	for _, fn := range fns {
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
			fmt.Println("Got the db!", db)
			res.RiverConfig = fmt.Sprintf(`prometheus.exporter.mysql "default" {
  data_source_name = "%s"
}`, dsn)
			res.MetricsExport = "prometheus.exporter.mysql.default.targets"
			return res, nil
		}
	}

	// Our predefined configurations didn't work; but MySQL is running.
	// Let's return a Flow component template for the user to fill out.
	res.RiverConfig = `prometheus.exporter.apache "default" {
  data_source_name = env("AGENT_MYSQL_DSN")
}`
	res.MetricsExport = "prometheus.exporter.apache.default.targets"

	return res, nil
}
