package exporter

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/config"
	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/job"
	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/logger"
	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/model"
	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/promutil"
	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/services"
	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/session"
)

// Metrics is a slice of prometheus metrics specific to the scraping process such API call counters
var Metrics = []prometheus.Collector{
	promutil.CloudwatchAPICounter,
	promutil.CloudwatchAPIErrorCounter,
	promutil.CloudwatchGetMetricDataAPICounter,
	promutil.CloudwatchGetMetricStatisticsAPICounter,
	promutil.ResourceGroupTaggingAPICounter,
	promutil.AutoScalingAPICounter,
	promutil.TargetGroupsAPICounter,
	promutil.APIGatewayAPICounter,
	promutil.Ec2APICounter,
	promutil.DmsAPICounter,
	promutil.StoragegatewayAPICounter,
}

// UpdateMetrics can be used to scrape metrics from AWS on demand using the provided parameters. Scraped metrics will be added to the provided registry and
// any labels discovered during the scrape will be added to observedMetricLabels with their metric name as the key. Any errors encountered are not returned but
// will be logged and will either fail the scrape or a partial metric result will be added to the registry.
func UpdateMetrics(
	ctx context.Context,
	config config.ScrapeConf,
	registry *prometheus.Registry,
	metricsPerQuery int,
	labelsSnakeCase bool,
	cloudwatchSemaphore, tagSemaphore chan struct{},
	cache session.SessionCache,
	observedMetricLabels map[string]model.LabelSet,
	logger logger.Logger,
) {
	tagsData, cloudwatchData := job.ScrapeAwsData(
		ctx,
		config,
		metricsPerQuery,
		cloudwatchSemaphore,
		tagSemaphore,
		cache,
		logger,
	)

	metrics, observedMetricLabels, err := job.MigrateCloudwatchToPrometheus(cloudwatchData, labelsSnakeCase, observedMetricLabels, logger)
	if err != nil {
		logger.Error(err, "Error migrating cloudwatch metrics to prometheus metrics")
		return
	}
	metrics = job.EnsureLabelConsistencyForMetrics(metrics, observedMetricLabels)

	metrics = append(metrics, services.MigrateTagsToPrometheus(tagsData, labelsSnakeCase, logger)...)

	registry.MustRegister(promutil.NewPrometheusCollector(metrics))
}
