package main

import (
	"context"

	"github.com/cortexproject/cortex/pkg/util"
	"github.com/go-kit/kit/log"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/scrape"
)

func main() {
	ctxScrape, cancelScrape := context.WithCancel(context.Background())
	discoveryManagerScrape := discovery.NewManager(ctxScrape, log.With(util.Logger, "component", "discovery manager scrape"), discovery.Name("scrape"))

	wal := agent.NewWALStorage()

	scrapeManager := scrape.NewManager(log.With(util.Logger, "component", "scrape manager"), wal)
}
