// Copyright 2020 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collectors

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"google.golang.org/api/monitoring/v3"

	"github.com/prometheus-community/stackdriver_exporter/utils"
)

const namespace = "stackdriver"

type MetricFilter struct {
	Prefix   string
	Modifier string
}

type MonitoringCollector struct {
	projectID                       string
	metricsTypePrefixes             []string
	metricsFilters                  []MetricFilter
	metricsInterval                 time.Duration
	metricsOffset                   time.Duration
	metricsIngestDelay              bool
	monitoringService               *monitoring.Service
	apiCallsTotalMetric             prometheus.Counter
	scrapesTotalMetric              prometheus.Counter
	scrapeErrorsTotalMetric         prometheus.Counter
	lastScrapeErrorMetric           prometheus.Gauge
	lastScrapeTimestampMetric       prometheus.Gauge
	lastScrapeDurationSecondsMetric prometheus.Gauge
	collectorFillMissingLabels      bool
	monitoringDropDelegatedProjects bool
	logger                          log.Logger
	deltaCounterStore               DeltaCounterStore
	deltaDistributionStore          DeltaDistributionStore
	aggregateDeltas                 bool
}

type MonitoringCollectorOptions struct {
	// MetricTypePrefixes are the Google Monitoring (ex-Stackdriver) metric type prefixes that the collector
	// will be querying.
	MetricTypePrefixes []string
	// ExtraFilters is a list of criteria to apply to each corresponding metric prefix query. If one or more are
	// applicable to a given metric type prefix, they will be 'AND' concatenated.
	ExtraFilters []MetricFilter
	// RequestInterval is the time interval used in each request to get metrics. If there are many data points returned
	// during this interval, only the latest will be reported.
	RequestInterval time.Duration
	// RequestOffset is used to offset the requested interval into the past.
	RequestOffset time.Duration
	// IngestDelay decides if the ingestion delay specified in the metrics metadata is used when calculating the
	// request time interval.
	IngestDelay bool
	// FillMissingLabels decides if metric labels should be added with empty string to prevent failures due to label inconsistency on metrics.
	FillMissingLabels bool
	// DropDelegatedProjects decides if only metrics matching the collector's projectID should be retrieved.
	DropDelegatedProjects bool
	// AggregateDeltas decides if DELTA metrics should be treated as a counter using the provided counterStore/distributionStore or a gauge
	AggregateDeltas bool
}

func NewMonitoringCollector(projectID string, monitoringService *monitoring.Service, opts MonitoringCollectorOptions, logger log.Logger, counterStore DeltaCounterStore, distributionStore DeltaDistributionStore) (*MonitoringCollector, error) {
	const subsystem = "monitoring"

	apiCallsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "api_calls_total",
			Help:        "Total number of Google Stackdriver Monitoring API calls made.",
			ConstLabels: prometheus.Labels{"project_id": projectID},
		},
	)

	scrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "scrapes_total",
			Help:        "Total number of Google Stackdriver Monitoring metrics scrapes.",
			ConstLabels: prometheus.Labels{"project_id": projectID},
		},
	)

	scrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "scrape_errors_total",
			Help:        "Total number of Google Stackdriver Monitoring metrics scrape errors.",
			ConstLabels: prometheus.Labels{"project_id": projectID},
		},
	)

	lastScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "last_scrape_error",
			Help:        "Whether the last metrics scrape from Google Stackdriver Monitoring resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"project_id": projectID},
		},
	)

	lastScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "last_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last metrics scrape from Google Stackdriver Monitoring.",
			ConstLabels: prometheus.Labels{"project_id": projectID},
		},
	)

	lastScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "last_scrape_duration_seconds",
			Help:        "Duration of the last metrics scrape from Google Stackdriver Monitoring.",
			ConstLabels: prometheus.Labels{"project_id": projectID},
		},
	)

	monitoringCollector := &MonitoringCollector{
		projectID:                       projectID,
		metricsTypePrefixes:             opts.MetricTypePrefixes,
		metricsFilters:                  opts.ExtraFilters,
		metricsInterval:                 opts.RequestInterval,
		metricsOffset:                   opts.RequestOffset,
		metricsIngestDelay:              opts.IngestDelay,
		monitoringService:               monitoringService,
		apiCallsTotalMetric:             apiCallsTotalMetric,
		scrapesTotalMetric:              scrapesTotalMetric,
		scrapeErrorsTotalMetric:         scrapeErrorsTotalMetric,
		lastScrapeErrorMetric:           lastScrapeErrorMetric,
		lastScrapeTimestampMetric:       lastScrapeTimestampMetric,
		lastScrapeDurationSecondsMetric: lastScrapeDurationSecondsMetric,
		collectorFillMissingLabels:      opts.FillMissingLabels,
		monitoringDropDelegatedProjects: opts.DropDelegatedProjects,
		logger:                          logger,
		deltaCounterStore:               counterStore,
		deltaDistributionStore:          distributionStore,
		aggregateDeltas:                 opts.AggregateDeltas,
	}

	return monitoringCollector, nil
}

func (c *MonitoringCollector) Describe(ch chan<- *prometheus.Desc) {
	c.apiCallsTotalMetric.Describe(ch)
	c.scrapesTotalMetric.Describe(ch)
	c.scrapeErrorsTotalMetric.Describe(ch)
	c.lastScrapeErrorMetric.Describe(ch)
	c.lastScrapeTimestampMetric.Describe(ch)
	c.lastScrapeDurationSecondsMetric.Describe(ch)
}

func (c *MonitoringCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	if err := c.reportMonitoringMetrics(ch, begun); err != nil {
		errorMetric = float64(1)
		c.scrapeErrorsTotalMetric.Inc()
		level.Error(c.logger).Log("msg", "Error while getting Google Stackdriver Monitoring metrics", "err", err)
	}
	c.scrapeErrorsTotalMetric.Collect(ch)

	c.apiCallsTotalMetric.Collect(ch)

	c.scrapesTotalMetric.Inc()
	c.scrapesTotalMetric.Collect(ch)

	c.lastScrapeErrorMetric.Set(errorMetric)
	c.lastScrapeErrorMetric.Collect(ch)

	c.lastScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastScrapeTimestampMetric.Collect(ch)

	c.lastScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastScrapeDurationSecondsMetric.Collect(ch)
}

func (c *MonitoringCollector) reportMonitoringMetrics(ch chan<- prometheus.Metric, begun time.Time) error {
	metricDescriptorsFunction := func(page *monitoring.ListMetricDescriptorsResponse) error {
		var wg = &sync.WaitGroup{}

		c.apiCallsTotalMetric.Inc()

		// It has been noticed that the same metric descriptor can be obtained from different GCP
		// projects. When that happens, metrics are fetched twice and it provokes the error:
		//     "collected metric xxx was collected before with the same name and label values"
		//
		// Metric descriptor project is irrelevant when it comes to fetch metrics, as they will be
		// fetched from all the delegated projects filtering by metric type. Considering that, we
		// can filter descriptors to keep just one per type.
		//
		// The following makes sure metric descriptors are unique to avoid fetching more than once
		uniqueDescriptors := make(map[string]*monitoring.MetricDescriptor)
		for _, descriptor := range page.MetricDescriptors {
			uniqueDescriptors[descriptor.Type] = descriptor
		}

		errChannel := make(chan error, len(uniqueDescriptors))

		endTime := time.Now().UTC().Add(c.metricsOffset * -1)
		startTime := endTime.Add(c.metricsInterval * -1)

		for _, metricDescriptor := range uniqueDescriptors {
			wg.Add(1)
			go func(metricDescriptor *monitoring.MetricDescriptor, ch chan<- prometheus.Metric, startTime, endTime time.Time) {
				defer wg.Done()
				level.Debug(c.logger).Log("msg", "retrieving Google Stackdriver Monitoring metrics for descriptor", "descriptor", metricDescriptor.Type)
				filter := fmt.Sprintf("metric.type=\"%s\"", metricDescriptor.Type)
				if c.monitoringDropDelegatedProjects {
					filter = fmt.Sprintf(
						"project=\"%s\" AND metric.type=\"%s\"",
						c.projectID,
						metricDescriptor.Type)
				}

				if c.metricsIngestDelay &&
					metricDescriptor.Metadata != nil &&
					metricDescriptor.Metadata.IngestDelay != "" {
					ingestDelay := metricDescriptor.Metadata.IngestDelay
					ingestDelayDuration, err := time.ParseDuration(ingestDelay)
					if err != nil {
						level.Error(c.logger).Log("msg", "error parsing ingest delay from metric metadata", "descriptor", metricDescriptor.Type, "err", err, "delay", ingestDelay)
						errChannel <- err
						return
					}
					level.Debug(c.logger).Log("msg", "adding ingest delay", "descriptor", metricDescriptor.Type, "delay", ingestDelay)
					endTime = endTime.Add(ingestDelayDuration * -1)
					startTime = startTime.Add(ingestDelayDuration * -1)
				}

				for _, ef := range c.metricsFilters {
					if strings.Contains(metricDescriptor.Type, ef.Prefix) {
						filter = fmt.Sprintf("%s AND (%s)", filter, ef.Modifier)
					}
				}

				level.Debug(c.logger).Log("msg", "retrieving Google Stackdriver Monitoring metrics with filter", "filter", filter)

				timeSeriesListCall := c.monitoringService.Projects.TimeSeries.List(utils.ProjectResource(c.projectID)).
					Filter(filter).
					IntervalStartTime(startTime.Format(time.RFC3339Nano)).
					IntervalEndTime(endTime.Format(time.RFC3339Nano))

				for {
					c.apiCallsTotalMetric.Inc()
					page, err := timeSeriesListCall.Do()
					if err != nil {
						level.Error(c.logger).Log("msg", "error retrieving Time Series metrics for descriptor", "descriptor", metricDescriptor.Type, "err", err)
						errChannel <- err
						break
					}
					if page == nil {
						break
					}
					if err := c.reportTimeSeriesMetrics(page, metricDescriptor, ch, begun); err != nil {
						level.Error(c.logger).Log("msg", "error reporting Time Series metrics for descriptor", "descriptor", metricDescriptor.Type, "err", err)
						errChannel <- err
						break
					}
					if page.NextPageToken == "" {
						break
					}
					timeSeriesListCall.PageToken(page.NextPageToken)
				}
			}(metricDescriptor, ch, startTime, endTime)
		}

		wg.Wait()
		close(errChannel)

		return <-errChannel
	}

	var wg = &sync.WaitGroup{}

	errChannel := make(chan error, len(c.metricsTypePrefixes))

	for _, metricsTypePrefix := range c.metricsTypePrefixes {
		wg.Add(1)
		go func(metricsTypePrefix string) {
			defer wg.Done()
			level.Debug(c.logger).Log("msg", "listing Google Stackdriver Monitoring metric descriptors starting with", "prefix", metricsTypePrefix)
			ctx := context.Background()
			filter := fmt.Sprintf("metric.type = starts_with(\"%s\")", metricsTypePrefix)
			if c.monitoringDropDelegatedProjects {
				filter = fmt.Sprintf(
					"project = \"%s\" AND metric.type = starts_with(\"%s\")",
					c.projectID,
					metricsTypePrefix)
			}
			if err := c.monitoringService.Projects.MetricDescriptors.List(utils.ProjectResource(c.projectID)).
				Filter(filter).
				Pages(ctx, metricDescriptorsFunction); err != nil {
				errChannel <- err
			}
		}(metricsTypePrefix)
	}

	wg.Wait()
	close(errChannel)

	level.Debug(c.logger).Log("msg", "Done reporting monitoring metrics")
	return <-errChannel
}

func (c *MonitoringCollector) reportTimeSeriesMetrics(
	page *monitoring.ListTimeSeriesResponse,
	metricDescriptor *monitoring.MetricDescriptor,
	ch chan<- prometheus.Metric,
	begun time.Time,
) error {
	var metricValue float64
	var metricValueType prometheus.ValueType
	var newestTSPoint *monitoring.Point

	timeSeriesMetrics, err := NewTimeSeriesMetrics(metricDescriptor,
		ch,
		c.collectorFillMissingLabels,
		c.deltaCounterStore,
		c.deltaDistributionStore,
		c.aggregateDeltas,
	)
	if err != nil {
		return fmt.Errorf("error creating the TimeSeriesMetrics %v", err)
	}
	for _, timeSeries := range page.TimeSeries {
		newestEndTime := time.Unix(0, 0)
		for _, point := range timeSeries.Points {
			endTime, err := time.Parse(time.RFC3339Nano, point.Interval.EndTime)
			if err != nil {
				return fmt.Errorf("Error parsing TimeSeries Point interval end time `%s`: %s", point.Interval.EndTime, err)
			}
			if endTime.After(newestEndTime) {
				newestEndTime = endTime
				newestTSPoint = point
			}
		}
		labelKeys := []string{"unit"}
		labelValues := []string{metricDescriptor.Unit}

		// Add the metric labels
		// @see https://cloud.google.com/monitoring/api/metrics
		for key, value := range timeSeries.Metric.Labels {
			if !c.keyExists(labelKeys, key) {
				labelKeys = append(labelKeys, key)
				labelValues = append(labelValues, value)
			}
		}

		// Add the monitored resource labels
		// @see https://cloud.google.com/monitoring/api/resources
		for key, value := range timeSeries.Resource.Labels {
			if !c.keyExists(labelKeys, key) {
				labelKeys = append(labelKeys, key)
				labelValues = append(labelValues, value)
			}
		}

		if c.monitoringDropDelegatedProjects {
			dropDelegatedProject := false

			for idx, val := range labelKeys {
				if val == "project_id" && labelValues[idx] != c.projectID {
					dropDelegatedProject = true
					break
				}
			}

			if dropDelegatedProject {
				continue
			}
		}

		switch timeSeries.MetricKind {
		case "GAUGE":
			metricValueType = prometheus.GaugeValue
		case "DELTA":
			if c.aggregateDeltas {
				metricValueType = prometheus.CounterValue
			} else {
				metricValueType = prometheus.GaugeValue
			}
		case "CUMULATIVE":
			metricValueType = prometheus.CounterValue
		default:
			continue
		}

		switch timeSeries.ValueType {
		case "BOOL":
			metricValue = 0
			if *newestTSPoint.Value.BoolValue {
				metricValue = 1
			}
		case "INT64":
			metricValue = float64(*newestTSPoint.Value.Int64Value)
		case "DOUBLE":
			metricValue = *newestTSPoint.Value.DoubleValue
		case "DISTRIBUTION":
			dist := newestTSPoint.Value.DistributionValue
			buckets, err := c.generateHistogramBuckets(dist)

			if err == nil {
				timeSeriesMetrics.CollectNewConstHistogram(timeSeries, newestEndTime, labelKeys, dist, buckets, labelValues, timeSeries.MetricKind)
			} else {
				level.Debug(c.logger).Log("msg", "discarding", "resource", timeSeries.Resource.Type, "metric",
					timeSeries.Metric.Type, "err", err)
			}
			continue
		default:
			level.Debug(c.logger).Log("msg", "discarding", "value_type", timeSeries.ValueType, "metric", timeSeries)
			continue
		}

		timeSeriesMetrics.CollectNewConstMetric(timeSeries, newestEndTime, labelKeys, metricValueType, metricValue, labelValues, timeSeries.MetricKind)
	}
	timeSeriesMetrics.Complete(begun)
	return nil
}

func (c *MonitoringCollector) generateHistogramBuckets(
	dist *monitoring.Distribution,
) (map[float64]uint64, error) {
	opts := dist.BucketOptions
	var bucketKeys []float64
	switch {
	case opts.ExplicitBuckets != nil:
		// @see https://cloud.google.com/monitoring/api/ref_v3/rest/v3/TypedValue#explicit
		bucketKeys = make([]float64, len(opts.ExplicitBuckets.Bounds)+1)
		copy(bucketKeys, opts.ExplicitBuckets.Bounds)
	case opts.LinearBuckets != nil:
		// @see https://cloud.google.com/monitoring/api/ref_v3/rest/v3/TypedValue#linear
		// NumFiniteBuckets is inclusive so bucket count is num+2
		num := int(opts.LinearBuckets.NumFiniteBuckets)
		bucketKeys = make([]float64, num+2)
		for i := 0; i <= num; i++ {
			bucketKeys[i] = opts.LinearBuckets.Offset + (float64(i) * opts.LinearBuckets.Width)
		}
	case opts.ExponentialBuckets != nil:
		// @see https://cloud.google.com/monitoring/api/ref_v3/rest/v3/TypedValue#exponential
		// NumFiniteBuckets is inclusive so bucket count is num+2
		num := int(opts.ExponentialBuckets.NumFiniteBuckets)
		bucketKeys = make([]float64, num+2)
		for i := 0; i <= num; i++ {
			bucketKeys[i] = opts.ExponentialBuckets.Scale * math.Pow(opts.ExponentialBuckets.GrowthFactor, float64(i))
		}
	default:
		return nil, errors.New("Unknown distribution buckets")
	}
	// The last bucket is always infinity
	// @see https://cloud.google.com/monitoring/api/ref_v3/rest/v3/TypedValue#bucketoptions
	bucketKeys[len(bucketKeys)-1] = math.Inf(1)

	// Prometheus expects each bucket to have a lower bound of 0, but Google
	// sends a bucket with a lower bound of the previous bucket's upper bound, so
	// we need to store the last bucket and add it to the next bucket to make it
	// 0-bound.
	// Any remaining keys without data have a value of 0
	buckets := map[float64]uint64{}
	var last uint64
	for i, b := range bucketKeys {
		if len(dist.BucketCounts) > i {
			buckets[b] = uint64(dist.BucketCounts[i]) + last
			last = buckets[b]
		} else {
			buckets[b] = last
		}
	}
	return buckets, nil
}

func (c *MonitoringCollector) keyExists(labelKeys []string, key string) bool {
	for _, item := range labelKeys {
		if item == key {
			level.Debug(c.logger).Log("msg", "Found duplicate label key", "key", key)
			return true
		}
	}
	return false
}
