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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/api/monitoring/v3"

	"sort"

	"github.com/prometheus-community/stackdriver_exporter/utils"
)

func buildFQName(timeSeries *monitoring.TimeSeries) string {
	// The metric name to report is composed by the 3 parts:
	// 1. namespace is a constant prefix (stackdriver)
	// 2. subsystem is the monitored resource type (ie gce_instance)
	// 3. name is the metric type (ie compute.googleapis.com/instance/cpu/usage_time)
	return prometheus.BuildFQName(namespace, utils.NormalizeMetricName(timeSeries.Resource.Type), utils.NormalizeMetricName(timeSeries.Metric.Type))
}

type timeSeriesMetrics struct {
	metricDescriptor *monitoring.MetricDescriptor

	ch chan<- prometheus.Metric

	fillMissingLabels bool
	constMetrics      map[string][]*ConstMetric
	histogramMetrics  map[string][]*HistogramMetric

	deltaCounterStore      DeltaCounterStore
	deltaDistributionStore DeltaDistributionStore
	aggregateDeltas        bool
}

func NewTimeSeriesMetrics(descriptor *monitoring.MetricDescriptor,
	ch chan<- prometheus.Metric,
	fillMissingLabels bool,
	deltaCounterStore DeltaCounterStore,
	deltaDistributionStore DeltaDistributionStore,
	aggregateDeltas bool) (*timeSeriesMetrics, error) {

	return &timeSeriesMetrics{
		metricDescriptor:       descriptor,
		ch:                     ch,
		fillMissingLabels:      fillMissingLabels,
		constMetrics:           make(map[string][]*ConstMetric),
		histogramMetrics:       make(map[string][]*HistogramMetric),
		deltaCounterStore:      deltaCounterStore,
		deltaDistributionStore: deltaDistributionStore,
		aggregateDeltas:        aggregateDeltas,
	}, nil
}

func (t *timeSeriesMetrics) newMetricDesc(fqName string, labelKeys []string) *prometheus.Desc {
	return prometheus.NewDesc(
		fqName,
		t.metricDescriptor.Description,
		labelKeys,
		prometheus.Labels{},
	)
}

type ConstMetric struct {
	fqName      string
	labelKeys   []string
	valueType   prometheus.ValueType
	value       float64
	labelValues []string
	reportTime  time.Time

	keysHash uint64
}

type HistogramMetric struct {
	fqName      string
	labelKeys   []string
	dist        *monitoring.Distribution
	buckets     map[float64]uint64
	labelValues []string
	reportTime  time.Time

	keysHash uint64
}

func (t *timeSeriesMetrics) CollectNewConstHistogram(timeSeries *monitoring.TimeSeries, reportTime time.Time, labelKeys []string, dist *monitoring.Distribution, buckets map[float64]uint64, labelValues []string, metricKind string) {
	fqName := buildFQName(timeSeries)

	var v HistogramMetric
	if t.fillMissingLabels || (metricKind == "DELTA" && t.aggregateDeltas) {
		v = HistogramMetric{
			fqName:      fqName,
			labelKeys:   labelKeys,
			dist:        dist,
			buckets:     buckets,
			labelValues: labelValues,
			reportTime:  reportTime,

			keysHash: hashLabelKeys(labelKeys),
		}
	}

	if metricKind == "DELTA" && t.aggregateDeltas {
		t.deltaDistributionStore.Increment(t.metricDescriptor, &v)
		return
	}

	if t.fillMissingLabels {
		vs, ok := t.histogramMetrics[fqName]
		if !ok {
			vs = make([]*HistogramMetric, 0)
		}
		t.histogramMetrics[fqName] = append(vs, &v)
		return
	}

	t.ch <- t.newConstHistogram(fqName, reportTime, labelKeys, dist, buckets, labelValues)
}

func (t *timeSeriesMetrics) newConstHistogram(fqName string, reportTime time.Time, labelKeys []string, dist *monitoring.Distribution, buckets map[float64]uint64, labelValues []string) prometheus.Metric {
	return prometheus.NewMetricWithTimestamp(
		reportTime,
		prometheus.MustNewConstHistogram(
			t.newMetricDesc(fqName, labelKeys),
			uint64(dist.Count),
			dist.Mean*float64(dist.Count), // Stackdriver does not provide the sum, but we can fake it
			buckets,
			labelValues...,
		),
	)
}

func (t *timeSeriesMetrics) CollectNewConstMetric(timeSeries *monitoring.TimeSeries, reportTime time.Time, labelKeys []string, metricValueType prometheus.ValueType, metricValue float64, labelValues []string, metricKind string) {
	fqName := buildFQName(timeSeries)

	var v ConstMetric
	if t.fillMissingLabels || (metricKind == "DELTA" && t.aggregateDeltas) {
		v = ConstMetric{
			fqName:      fqName,
			labelKeys:   labelKeys,
			valueType:   metricValueType,
			value:       metricValue,
			labelValues: labelValues,
			reportTime:  reportTime,

			keysHash: hashLabelKeys(labelKeys),
		}
	}

	if metricKind == "DELTA" && t.aggregateDeltas {
		t.deltaCounterStore.Increment(t.metricDescriptor, &v)
		return
	}

	if t.fillMissingLabels {
		vs, ok := t.constMetrics[fqName]
		if !ok {
			vs = make([]*ConstMetric, 0)
		}
		t.constMetrics[fqName] = append(vs, &v)
		return
	}

	t.ch <- t.newConstMetric(fqName, reportTime, labelKeys, metricValueType, metricValue, labelValues)
}

func (t *timeSeriesMetrics) newConstMetric(fqName string, reportTime time.Time, labelKeys []string, metricValueType prometheus.ValueType, metricValue float64, labelValues []string) prometheus.Metric {
	return prometheus.NewMetricWithTimestamp(
		reportTime,
		prometheus.MustNewConstMetric(
			t.newMetricDesc(fqName, labelKeys),
			metricValueType,
			metricValue,
			labelValues...,
		),
	)
}

func hashLabelKeys(labelKeys []string) uint64 {
	dh := hashNew()
	sortedKeys := make([]string, len(labelKeys))
	copy(sortedKeys, labelKeys)
	sort.Strings(sortedKeys)
	for _, key := range sortedKeys {
		dh = hashAdd(dh, key)
		dh = hashAddByte(dh, separatorByte)
	}
	return dh
}

func (t *timeSeriesMetrics) Complete(reportingStartTime time.Time) {
	t.completeDeltaConstMetrics(reportingStartTime)
	t.completeDeltaHistogramMetrics(reportingStartTime)
	t.completeConstMetrics(t.constMetrics)
	t.completeHistogramMetrics(t.histogramMetrics)
}

func (t *timeSeriesMetrics) completeConstMetrics(constMetrics map[string][]*ConstMetric) {
	for _, vs := range constMetrics {
		if len(vs) > 1 {
			var needFill bool
			for i := 1; i < len(vs); i++ {
				if vs[0].keysHash != vs[i].keysHash {
					needFill = true
				}
			}
			if needFill {
				vs = fillConstMetricsLabels(vs)
			}
		}

		for _, v := range vs {
			t.ch <- t.newConstMetric(v.fqName, v.reportTime, v.labelKeys, v.valueType, v.value, v.labelValues)
		}
	}
}

func (t *timeSeriesMetrics) completeHistogramMetrics(histograms map[string][]*HistogramMetric) {
	for _, vs := range histograms {
		if len(vs) > 1 {
			var needFill bool
			for i := 1; i < len(vs); i++ {
				if vs[0].keysHash != vs[i].keysHash {
					needFill = true
				}
			}
			if needFill {
				vs = fillHistogramMetricsLabels(vs)
			}
		}
		for _, v := range vs {
			t.ch <- t.newConstHistogram(v.fqName, v.reportTime, v.labelKeys, v.dist, v.buckets, v.labelValues)
		}
	}
}

func (t *timeSeriesMetrics) completeDeltaConstMetrics(reportingStartTime time.Time) {
	descriptorMetrics := t.deltaCounterStore.ListMetrics(t.metricDescriptor.Name)
	now := time.Now().Truncate(time.Minute)

	constMetrics := map[string][]*ConstMetric{}
	for _, metrics := range descriptorMetrics {
		for _, collected := range metrics {
			// If the metric wasn't collected we should still export it at the next sample time to avoid staleness
			if reportingStartTime.After(collected.lastCollectedAt) {
				// Ideally we could use monitoring.MetricDescriptorMetadata.SamplePeriod to determine how many
				// samples were missed to adjust this but monitoring.MetricDescriptorMetadata is viewed as optional
				// for a monitoring.MetricDescriptor
				reportingLag := collected.lastCollectedAt.Sub(collected.metric.reportTime).Truncate(time.Minute)
				collected.metric.reportTime = now.Add(-reportingLag)
			}
			if t.fillMissingLabels {
				if _, exists := constMetrics[collected.metric.fqName]; !exists {
					constMetrics[collected.metric.fqName] = []*ConstMetric{}
				}
				constMetrics[collected.metric.fqName] = append(constMetrics[collected.metric.fqName], collected.metric)
			} else {
				t.ch <- t.newConstMetric(
					collected.metric.fqName,
					collected.metric.reportTime,
					collected.metric.labelKeys,
					collected.metric.valueType,
					collected.metric.value,
					collected.metric.labelValues,
				)
			}
		}
	}

	if t.fillMissingLabels {
		t.completeConstMetrics(constMetrics)
	}
}

func (t *timeSeriesMetrics) completeDeltaHistogramMetrics(reportingStartTime time.Time) {
	descriptorMetrics := t.deltaDistributionStore.ListMetrics(t.metricDescriptor.Name)
	now := time.Now().Truncate(time.Minute)

	histograms := map[string][]*HistogramMetric{}
	for _, metrics := range descriptorMetrics {
		for _, collected := range metrics {
			// If the histogram wasn't collected we should still export it at the next sample time to avoid staleness
			if reportingStartTime.After(collected.lastCollectedAt) {
				// Ideally we could use monitoring.MetricDescriptorMetadata.SamplePeriod to determine how many
				// samples were missed to adjust this but monitoring.MetricDescriptorMetadata is viewed as optional
				// for a monitoring.MetricDescriptor
				reportingLag := collected.lastCollectedAt.Sub(collected.histogram.reportTime).Truncate(time.Minute)
				collected.histogram.reportTime = now.Add(-reportingLag)
			}
			if t.fillMissingLabels {
				if _, exists := histograms[collected.histogram.fqName]; !exists {
					histograms[collected.histogram.fqName] = []*HistogramMetric{}
				}
				histograms[collected.histogram.fqName] = append(histograms[collected.histogram.fqName], collected.histogram)
			} else {
				t.ch <- t.newConstHistogram(
					collected.histogram.fqName,
					collected.histogram.reportTime,
					collected.histogram.labelKeys,
					collected.histogram.dist,
					collected.histogram.buckets,
					collected.histogram.labelValues,
				)
			}
		}
	}

	if t.fillMissingLabels {
		t.completeHistogramMetrics(histograms)
	}
}

func fillConstMetricsLabels(metrics []*ConstMetric) []*ConstMetric {
	allKeys := make(map[string]struct{})
	for _, metric := range metrics {
		for _, key := range metric.labelKeys {
			allKeys[key] = struct{}{}
		}
	}

	for _, metric := range metrics {
		if len(metric.labelKeys) != len(allKeys) {
			metricKeys := make(map[string]struct{})
			for _, key := range metric.labelKeys {
				metricKeys[key] = struct{}{}
			}
			for key := range allKeys {
				if _, ok := metricKeys[key]; !ok {
					metric.labelKeys = append(metric.labelKeys, key)
					metric.labelValues = append(metric.labelValues, "")
				}
			}
		}
	}

	return metrics
}

func fillHistogramMetricsLabels(metrics []*HistogramMetric) []*HistogramMetric {
	allKeys := make(map[string]struct{})
	for _, metric := range metrics {
		for _, key := range metric.labelKeys {
			allKeys[key] = struct{}{}
		}
	}

	for _, metric := range metrics {
		if len(metric.labelKeys) != len(allKeys) {
			metricKeys := make(map[string]struct{})
			for _, key := range metric.labelKeys {
				metricKeys[key] = struct{}{}
			}
			for key := range allKeys {
				if _, ok := metricKeys[key]; !ok {
					metric.labelKeys = append(metric.labelKeys, key)
					metric.labelValues = append(metric.labelValues, "")
				}
			}
		}
	}

	return metrics
}
