package promutil

import (
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
)

var (
	CloudwatchAPICounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "yace_cloudwatch_requests_total",
		Help: "Help is not implemented yet.",
	})
	CloudwatchAPIErrorCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "yace_cloudwatch_request_errors",
		Help: "Help is not implemented yet.",
	})
	CloudwatchGetMetricDataAPICounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "yace_cloudwatch_getmetricdata_requests_total",
		Help: "Help is not implemented yet.",
	})
	CloudwatchGetMetricStatisticsAPICounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "yace_cloudwatch_getmetricstatistics_requests_total",
		Help: "Help is not implemented yet.",
	})
	ResourceGroupTaggingAPICounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "yace_cloudwatch_resourcegrouptaggingapi_requests_total",
		Help: "Help is not implemented yet.",
	})
	AutoScalingAPICounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "yace_cloudwatch_autoscalingapi_requests_total",
		Help: "Help is not implemented yet.",
	})
	TargetGroupsAPICounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "yace_cloudwatch_targetgroupapi_requests_total",
		Help: "Help is not implemented yet.",
	})
	APIGatewayAPICounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "yace_cloudwatch_apigatewayapi_requests_total",
	})
	Ec2APICounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "yace_cloudwatch_ec2api_requests_total",
		Help: "Help is not implemented yet.",
	})
	ManagedPrometheusAPICounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "yace_cloudwatch_managedprometheusapi_requests_total",
		Help: "Help is not implemented yet.",
	})
	StoragegatewayAPICounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "yace_cloudwatch_storagegatewayapi_requests_total",
		Help: "Help is not implemented yet.",
	})
	DmsAPICounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "yace_cloudwatch_dmsapi_requests_total",
		Help: "Help is not implemented yet.",
	})
)

var replacer = strings.NewReplacer(
	" ", "_",
	",", "_",
	"\t", "_",
	"/", "_",
	"\\", "_",
	".", "_",
	"-", "_",
	":", "_",
	"=", "_",
	"“", "_",
	"@", "_",
	"<", "_",
	">", "_",
	"%", "_percent",
)
var splitRegexp = regexp.MustCompile(`([a-z0-9])([A-Z])`)

type PrometheusMetric struct {
	Name             *string
	Labels           map[string]string
	Value            *float64
	IncludeTimestamp bool
	Timestamp        time.Time
}

type PrometheusCollector struct {
	metrics []*PrometheusMetric
}

func NewPrometheusCollector(metrics []*PrometheusMetric) *PrometheusCollector {
	return &PrometheusCollector{
		metrics: removeDuplicatedMetrics(metrics),
	}
}

func (p *PrometheusCollector) Describe(descs chan<- *prometheus.Desc) {
	for _, metric := range p.metrics {
		descs <- createDesc(metric)
	}
}

func (p *PrometheusCollector) Collect(metrics chan<- prometheus.Metric) {
	for _, metric := range p.metrics {
		metrics <- createMetric(metric)
	}
}

func createDesc(metric *PrometheusMetric) *prometheus.Desc {
	return prometheus.NewDesc(
		*metric.Name,
		"Help is not implemented yet.",
		nil,
		metric.Labels,
	)
}

func createMetric(metric *PrometheusMetric) prometheus.Metric {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        *metric.Name,
		Help:        "Help is not implemented yet.",
		ConstLabels: metric.Labels,
	})

	gauge.Set(*metric.Value)

	if !metric.IncludeTimestamp {
		return gauge
	}

	return prometheus.NewMetricWithTimestamp(metric.Timestamp, gauge)
}

func removeDuplicatedMetrics(metrics []*PrometheusMetric) []*PrometheusMetric {
	keys := make(map[string]bool)
	filteredMetrics := []*PrometheusMetric{}
	for _, metric := range metrics {
		check := *metric.Name + combineLabels(metric.Labels)
		if _, value := keys[check]; !value {
			keys[check] = true
			filteredMetrics = append(filteredMetrics, metric)
		}
	}
	return filteredMetrics
}

func combineLabels(labels map[string]string) string {
	var combinedLabels string
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		combinedLabels += PromString(k) + PromString(labels[k])
	}
	return combinedLabels
}

func PromString(text string) string {
	text = splitString(text)
	return strings.ToLower(sanitize(text))
}

func PromStringTag(text string, labelsSnakeCase bool) (bool, string) {
	var s string
	if labelsSnakeCase {
		s = PromString(text)
	} else {
		s = sanitize(text)
	}
	return model.LabelName(s).IsValid(), s
}

func sanitize(text string) string {
	return replacer.Replace(text)
}

func splitString(text string) string {
	return splitRegexp.ReplaceAllString(text, `$1.$2`)
}
