package main

import (
	"context"

	"github.com/cortexproject/cortex/pkg/util"
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/wal"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/scrape"
)

func main() {
	ctxScrape, cancelScrape := context.WithCancel(context.Background())
	defer cancelScrape()

	discoveryManagerScrape := discovery.NewManager(ctxScrape, log.With(util.Logger, "component", "discovery manager scrape"), discovery.Name("scrape"))
	_ = discoveryManagerScrape

	wstore, err := wal.NewStorage(util.Logger, nil, ".walPath")
	if err != nil {
		panic(err)
	}

	scrapeManager := scrape.NewManager(log.With(util.Logger, "component", "scrape manager"), wstore)
	_ = scrapeManager
}
