package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/scrape"
	"github.com/grafana/agent/converter/internal/common"
	prom_config "github.com/prometheus/prometheus/config"
	prom_discovery "github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/storage"
)

func appendPrometheusScrape(pb *prometheusBlocks, scrapeConfig *prom_config.ScrapeConfig, forwardTo []storage.Appendable, targets []discovery.Target, label string) {
	scrapeArgs := toScrapeArguments(scrapeConfig, forwardTo, targets)
	block := common.NewBlockWithOverride([]string{"prometheus", "scrape"}, label, scrapeArgs)
	pb.prometheusScrapeBlocks = append(pb.prometheusScrapeBlocks, block)
}

func toScrapeArguments(scrapeConfig *prom_config.ScrapeConfig, forwardTo []storage.Appendable, targets []discovery.Target) *scrape.Arguments {
	if scrapeConfig == nil {
		return nil
	}

	return &scrape.Arguments{
		Targets:               targets,
		ForwardTo:             forwardTo,
		JobName:               scrapeConfig.JobName,
		HonorLabels:           scrapeConfig.HonorLabels,
		HonorTimestamps:       scrapeConfig.HonorTimestamps,
		Params:                scrapeConfig.Params,
		ScrapeInterval:        time.Duration(scrapeConfig.ScrapeInterval),
		ScrapeTimeout:         time.Duration(scrapeConfig.ScrapeTimeout),
		MetricsPath:           scrapeConfig.MetricsPath,
		Scheme:                scrapeConfig.Scheme,
		BodySizeLimit:         scrapeConfig.BodySizeLimit,
		SampleLimit:           scrapeConfig.SampleLimit,
		TargetLimit:           scrapeConfig.TargetLimit,
		LabelLimit:            scrapeConfig.LabelLimit,
		LabelNameLengthLimit:  scrapeConfig.LabelNameLengthLimit,
		LabelValueLengthLimit: scrapeConfig.LabelValueLengthLimit,
		HTTPClientConfig:      *toHttpClientConfig(&scrapeConfig.HTTPClientConfig),
		ExtraMetrics:          false,
		Clustering:            scrape.Clustering{Enabled: false},
	}
}

func getScrapeTargets(staticConfig prom_discovery.StaticConfig) []discovery.Target {
	targets := []discovery.Target{}

	for _, target := range staticConfig {
		targetMap := map[string]string{}

		for labelName, labelValue := range target.Labels {
			targetMap[string(labelName)] = string(labelValue)
		}

		for _, labelSet := range target.Targets {
			for labelName, labelValue := range labelSet {
				targetMap[string(labelName)] = string(labelValue)
				targets = append(targets, targetMap)
			}
		}
	}

	return targets
}
