package consul

import (
	"fmt"
	"os"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/autodiscovery"
	"github.com/grafana/agent/pkg/river"
	consul_api "github.com/hashicorp/consul/api"
	"github.com/mitchellh/go-ps"
)

// Config holds the configuration we will try to connect to a running Consul
// instance with.
type Config struct {
	Binary     string   `river:"binary,attr"`
	Servers    []string `river:"servers,attr,optional"`
	Extensions []string `river:"ext,attr,optional"`
}

// Consul is an autodiscovery mechanism for a Consul database.
type Consul struct {
	binary  string
	servers []string
	ext     []string
}

// New creates a new auto-discovery Consul mechanism instance.
func New() (*Consul, error) {
	bb, err := os.ReadFile("pkg/autodiscovery/consul/consul.river")
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = river.Unmarshal(bb, &cfg)
	if err != nil {
		return nil, err
	}

	return &Consul{
		binary:  cfg.Binary,
		servers: cfg.Servers,
		ext:     cfg.Extensions,
	}, nil
}

// Run check whether a Consul instance is running, and if so, returns a
// `prometheus.exporter.consul` component that can read metrics from it.
func (c *Consul) Run() (*autodiscovery.Result, error) {
	procs, err := ps.Processes()
	if err != nil {
		return nil, fmt.Errorf("could not read processes from host system: %w", err)
	}

	var consulPID int
	for _, p := range procs {
		if p.Executable() == c.binary && p.PPid() == 1 {
			consulPID = p.Pid()
			break
		}
	}
	if consulPID == 0 {
		return nil, fmt.Errorf("no running consul instance was found")
	}

	// Consul is running, so we'll try to return _something_.
	res := &autodiscovery.Result{}
	lsof := autodiscovery.LSOF{}

	fns, err := autodiscovery.GetOpenFilenames(lsof, consulPID, c.ext...)
	if err != nil {
		return nil, err
	}
	for fn, _ := range fns {
		res.LogfileTargets = append(res.LogfileTargets,
			discovery.Target{"__path__": fn, "component": "consul"},
		)
	}

	for _, srv := range c.servers {
		config := &consul_api.Config{}
		config.Address = srv
		_, err := consul_api.NewClient(config)
		if err != nil {
			continue
		} else {
			res.RiverConfig = fmt.Sprintf(` prometheus.exporter.consul "default" {
  server = "%s"
}`, srv)
			res.MetricsExport = "prometheus.exporter.consul.default.targets"
			return res, nil
		}
	}

	// Our predefined configurations didn't work; but Postgres is running.
	// Let's return a Flow component template for the user to fill out.
	res.RiverConfig = `prometheus.exporter.consul "default" {
  // NOTE: Agent Autodiscovery could not automatically configure a Consul exporter.
  // To set up a Consul exporter, please either set "server" explicitly
  // or set up the AGENT_CONSUL_SERVER environment variable and restart the Agent.
  server = env("AGENT_CONSUL_SERVER")
}`
	res.MetricsExport = "prometheus.exporter.consul.default.targets"

	return res, nil
}
