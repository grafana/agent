package redis

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/autodiscovery"
	"github.com/grafana/agent/pkg/river"
	"github.com/mitchellh/go-ps"
)

type Config struct {
	Binary         string   `river:"binary,attr"`
	RedisAddresses []string `river:"redis_addresses,attr,optional"`
	Extensions     []string `river:"ext,attr,optional"`
}

type Redis struct {
	binary         string
	redisAddresses []string
	ext            []string
}

func New() (*Redis, error) {
	bb, err := os.ReadFile("pkg/autodiscovery/redis/redis.river")
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = river.Unmarshal(bb, &cfg)
	if err != nil {
		return nil, err
	}

	return &Redis{
		binary:         cfg.Binary,
		redisAddresses: cfg.RedisAddresses,
		ext:            cfg.Extensions,
	}, nil
}

// Run check whether a Redis instance is running, and if so, returns a
// `prometheus.exporter.redis` component that can read metrics from it.
func (m *Redis) Run() (*autodiscovery.Result, error) {
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

	// Redis is running on the host system, so we'll try to return _something_.
	res := &autodiscovery.Result{}
	var lsof autodiscovery.LSOF

	fns, err := autodiscovery.GetOpenFilenames(lsof, pid, m.ext...)
	if err != nil {
		return nil, err
	}
	for fn, _ := range fns {
		res.LogfileTargets = append(res.LogfileTargets,
			discovery.Target{"__path__": fn, "component": "redis"},
		)
	}

	// Let's try to use the configuration to connect using predefined URIs.
	for _, uri := range m.redisAddresses {
		rdb := redis.NewClient(&redis.Options{
			Addr: uri,
		})

		//TODO: What should be the scope of ctx?
		var ctx = context.Background()
		redisStatus := rdb.Ping(ctx)

		if redisStatus.Err() != nil {
			continue
		}

		res.RiverConfig = fmt.Sprintf(`prometheus.exporter.redis "default" {
    redis_addr = "%s"
}`, uri)
		res.MetricsExport = "prometheus.exporter.redis.default.targets"

		return res, nil
	}

	// Our predefined configurations didn't work; but MySQL is running.
	// Let's return a Flow component template for the user to fill out.
	res.RiverConfig = `prometheus.exporter.redis "default" {
  // NOTE: Agent Autodiscovery could not automatically configure a Redis exporter.
  // To set up a Consul exporter, please either set "redis_addr" explicitly
  // or set up the REDIS_SERVER_ADDRESS environment variable and restart the Agent.
  redis_addr = env("REDIS_SERVER_ADDRESS")
}`
	res.MetricsExport = "prometheus.exporter.redis.default.targets"

	return res, nil
}

func isRealServerStatusPage(httpResp *http.Response) bool {
	return false
}
