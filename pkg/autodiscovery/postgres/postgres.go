package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/autodiscovery"
	"github.com/grafana/agent/pkg/river"
	"github.com/jackc/pgx/v5"
	"github.com/mitchellh/go-ps"
)

// Config holds the predefined DSNs we will try to connect to a running
// Postgres instance with.
type Config struct {
	Binary     string   `river:"binary,attr"`
	DSN        []string `river:"dsn,attr,optional"`
	Extensions []string `river:"ext,attr,optional"`
}

// Postgres is an autodiscovery mechanism for a Postgres database.
type Postgres struct {
	binary string
	dsn    []string
	ext    []string
}

// New creates a new auto-discovery Postgres mechanism instance.
func New() (*Postgres, error) {
	bb, err := os.ReadFile("pkg/autodiscovery/postgres/postgres.river")
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = river.Unmarshal(bb, &cfg)
	if err != nil {
		return nil, err
	}

	return &Postgres{
		binary: cfg.Binary,
		dsn:    cfg.DSN,
		ext:    cfg.Extensions,
	}, nil
}

// Run check whether a Postgres instance is running, and if so, returns a
// `prometheus.exporter.postgres` component that can read metrics from it.
func (pg *Postgres) Run() (*autodiscovery.Result, error) {
	procs, err := ps.Processes()
	if err != nil {
		return nil, fmt.Errorf("could not read processes from host system: %w", err)
	}

	var postgresPID int
	for _, p := range procs {
		// We have multiple processes with the 'postgres' binary name, each
		// each with a different goal.
		// $ ps aux | grep postgres
		// tpaschalis       84827   0.0  0.0 409117248   1872   ??  Ss    1:47PM   0:00.00 postgres: logical replication launcher
		// tpaschalis       84826   0.0  0.0 408866368   1040   ??  Ss    1:47PM   0:00.02 postgres: stats collector
		// tpaschalis       84825   0.0  0.0 409010752   2080   ??  Ss    1:47PM   0:00.02 postgres: autovacuum launcher
		// tpaschalis       84824   0.0  0.0 408986176   1344   ??  Ss    1:47PM   0:00.01 postgres: walwriter
		// tpaschalis       84823   0.0  0.0 408994368   1424   ??  Ss    1:47PM   0:00.02 postgres: background writer
		// tpaschalis       84822   0.0  0.0 409136704   1632   ??  Ss    1:47PM   0:00.00 postgres: checkpointer
		// tpaschalis       84809   0.0  0.0 408986496   4832   ??  S     1:47PM   0:00.03 /opt/homebrew/opt/postgresql@14/bin/postgres -D /opt/homebrew/var/postgresql@14
		//
		// The one that was started as a service has a parent PID of 1, so
		// let's roll with that for now.
		if p.Executable() == pg.binary && p.PPid() == 1 {
			postgresPID = p.Pid()
			break
		}
	}
	if postgresPID == 0 {
		return nil, fmt.Errorf("no running postgresql instance was found")
	}

	// Postgres is running, so we'll try to return _something_.
	res := &autodiscovery.Result{}
	lsof := autodiscovery.LSOF{}

	fns, err := autodiscovery.GetOpenFilenames(lsof, postgresPID, pg.ext...)
	if err != nil {
		return nil, err
	}
	for fn, _ := range fns {
		res.LogfileTargets = append(res.LogfileTargets,
			discovery.Target{"__path__": fn, "component": "postgres"},
		)
	}

	for _, dsn := range pg.dsn {
		conn, err := pgx.Connect(context.Background(), dsn)
		if err != nil {
			continue
		} else {
			defer conn.Close(context.Background())
			res.RiverConfig = fmt.Sprintf(`prometheus.exporter.postgres "default" {
	data_source_names = ["%s"]
	}`, dsn)
			res.MetricsExport = "prometheus.exporter.postgres.default.targets"
			return res, nil
		}
	}

	// Our predefined configurations didn't work; but Postgres is running.
	// Let's return a Flow component template for the user to fill out.
	res.RiverConfig = `prometheus.exporter.postgres "default" {
    // NOTE: Agent Autodiscovery could not automatically configure a Postgres exporter.
    // To set up a Consul exporter, please either set "data_source_names" explicitly
    // or set up the AGENT_POSTGRES_DSN environment variable and restart the Agent.
    data_source_names = [env("AGENT_POSTGRES_DSN")]
}`
	res.MetricsExport = "prometheus.exporter.postgres.default.targets"

	return res, nil
}
