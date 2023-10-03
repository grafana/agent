package prometheusconvert

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/exp/maps"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/scrape"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	prom_config "github.com/prometheus/prometheus/config"
	prom_discovery "github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/storage"
)

func appendPrometheusScrape(pb *prometheusBlocks, scrapeConfig *prom_config.ScrapeConfig, forwardTo []storage.Appendable, targets []discovery.Target, label string) {
	scrapeArgs := toScrapeArguments(scrapeConfig, forwardTo, targets)
	name := []string{"prometheus", "scrape"}
	block := common.NewBlockWithOverride(name, label, scrapeArgs)
	summary := fmt.Sprintf("Converted scrape_configs job_name %q into...", scrapeConfig.JobName)
	detail := fmt.Sprintf("	A %s.%s component", strings.Join(name, "."), label)
	pb.prometheusScrapeBlocks = append(pb.prometheusScrapeBlocks, newPrometheusBlock(block, name, label, summary, detail))
}

func validatePrometheusScrape(scrapeConfig *prom_config.ScrapeConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	if scrapeConfig.NativeHistogramBucketLimit != 0 {
		diags.Add(diag.SeverityLevelError, "unsupported native_histogram_bucket_limit for scrape_configs")
	}

	diags.AddAll(ValidateHttpClientConfig(&scrapeConfig.HTTPClientConfig))

	return diags
}

func toScrapeArguments(scrapeConfig *prom_config.ScrapeConfig, forwardTo []storage.Appendable, targets []discovery.Target) *scrape.Arguments {
	if scrapeConfig == nil {
		return nil
	}

	return &scrape.Arguments{
		Targets:                   targets,
		ForwardTo:                 forwardTo,
		JobName:                   scrapeConfig.JobName,
		HonorLabels:               scrapeConfig.HonorLabels,
		HonorTimestamps:           scrapeConfig.HonorTimestamps,
		Params:                    scrapeConfig.Params,
		ScrapeClassicHistograms:   scrapeConfig.ScrapeClassicHistograms,
		ScrapeInterval:            time.Duration(scrapeConfig.ScrapeInterval),
		ScrapeTimeout:             time.Duration(scrapeConfig.ScrapeTimeout),
		MetricsPath:               scrapeConfig.MetricsPath,
		Scheme:                    scrapeConfig.Scheme,
		BodySizeLimit:             scrapeConfig.BodySizeLimit,
		SampleLimit:               scrapeConfig.SampleLimit,
		TargetLimit:               scrapeConfig.TargetLimit,
		LabelLimit:                scrapeConfig.LabelLimit,
		LabelNameLengthLimit:      scrapeConfig.LabelNameLengthLimit,
		LabelValueLengthLimit:     scrapeConfig.LabelValueLengthLimit,
		HTTPClientConfig:          *ToHttpClientConfig(&scrapeConfig.HTTPClientConfig),
		ExtraMetrics:              false,
		EnableProtobufNegotiation: false,
		Clustering:                scrape.Clustering{Enabled: false},
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
				newMap := map[string]string{}
				maps.Copy(newMap, targetMap)
				targets = append(targets, newMap)
			}
		}
	}

	return targets
}

func validateScrapeTargets(staticConfig prom_discovery.StaticConfig) diag.Diagnostics {
	return make(diag.Diagnostics, 0)
}
