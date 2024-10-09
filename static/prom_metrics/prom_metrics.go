package prom_metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/scrape"
)

var SDMetrics map[string]discovery.DiscovererMetrics
var ScrapeManagerMetrics *scrape.ScrapeMetrics

func init() {
	var err error
	SDMetrics, err = discovery.CreateAndRegisterSDMetrics(prometheus.DefaultRegisterer)
	if err != nil {
		panic(fmt.Sprintf("failed to create and register Prometheus SD metrics: %s", err))
	}

	ScrapeManagerMetrics, err = scrape.NewScrapeMetrics(prometheus.DefaultRegisterer)
	if err != nil {
		panic(fmt.Sprintf("failed to create Prometheus scrape manager metrics: %s", err))
	}
}
