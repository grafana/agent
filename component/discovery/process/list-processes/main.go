//go:build linux

package main

import (
	"os"
	"sort"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	analCache "github.com/grafana/agent/component/discovery/process/analyze/cache"

	"github.com/grafana/agent/component/discovery/process"
)

var logger = log.NewLogfmtLogger(os.Stderr)

func run() error {
	cache := analCache.New(logger)
	processes, err := process.Discover(logger, &process.DiscoverConfig{}, cache)
	if err != nil {
		return err
	}

	var (
		keys       = make([]string, 16)
		attributes = make([]interface{}, 16)
	)

	for _, p := range processes {
		attributes = attributes[:4]
		attributes[0] = "msg"
		attributes[1] = "found process"
		attributes[2] = "pid"
		attributes[3] = p.PID

		keys = keys[:0]
		for k := range p.Analysis.Labels {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			attributes = append(attributes, k, p.Analysis.Labels[k])
		}

		level.Info(logger).Log(attributes...)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		level.Error(logger).Log("msg", "failed to discover processes", "err", err)
	}
}
