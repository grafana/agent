// Copyright 2013 The Prometheus Authors
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

package registry

import (
	"bytes"
	"fmt"
	"hash"
	"hash/fnv"
	"sort"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/statsd_exporter/pkg/clock"
	"github.com/prometheus/statsd_exporter/pkg/mapper"
	"github.com/prometheus/statsd_exporter/pkg/metrics"
)

// uncheckedCollector wraps a Collector but its Describe method yields no Desc.
// This allows incoming metrics to have inconsistent label sets
type uncheckedCollector struct {
	c prometheus.Collector
}

func (u uncheckedCollector) Describe(_ chan<- *prometheus.Desc) {}
func (u uncheckedCollector) Collect(c chan<- prometheus.Metric) {
	u.c.Collect(c)
}

type Registry struct {
	Registerer prometheus.Registerer
	Metrics    map[string]metrics.Metric
	Mapper     *mapper.MetricMapper
	// The below value and label variables are allocated in the registry struct
	// so that we don't have to allocate them every time have to compute a label
	// hash.
	ValueBuf, NameBuf bytes.Buffer
	Hasher            hash.Hash64
}

func NewRegistry(reg prometheus.Registerer, mapper *mapper.MetricMapper) *Registry {
	return &Registry{
		Registerer: reg,
		Metrics:    make(map[string]metrics.Metric),
		Mapper:     mapper,
		Hasher:     fnv.New64a(),
	}
}

func (r *Registry) MetricConflicts(metricName string, metricType metrics.MetricType) bool {
	vector, hasMetrics := r.Metrics[metricName]
	if !hasMetrics {
		// No metrics.Metric with this name exists
		return false
	}

	if vector.MetricType == metricType {
		// We've found a copy of this metrics.Metric with this type, but different
		// labels, so it's safe to create a new one.
		return false
	}

	// The metrics.Metric exists, but it's of a different type than we're trying to
	// create.
	return true
}

func (r *Registry) StoreCounter(metricName string, hash metrics.LabelHash, labels prometheus.Labels, vec *prometheus.CounterVec, c prometheus.Counter, ttl time.Duration) {
	r.Store(metricName, hash, labels, vec, c, metrics.CounterMetricType, ttl)
}

func (r *Registry) StoreGauge(metricName string, hash metrics.LabelHash, labels prometheus.Labels, vec *prometheus.GaugeVec, g prometheus.Gauge, ttl time.Duration) {
	r.Store(metricName, hash, labels, vec, g, metrics.GaugeMetricType, ttl)
}

func (r *Registry) StoreHistogram(metricName string, hash metrics.LabelHash, labels prometheus.Labels, vec *prometheus.HistogramVec, o prometheus.Observer, ttl time.Duration) {
	r.Store(metricName, hash, labels, vec, o, metrics.HistogramMetricType, ttl)
}

func (r *Registry) StoreSummary(metricName string, hash metrics.LabelHash, labels prometheus.Labels, vec *prometheus.SummaryVec, o prometheus.Observer, ttl time.Duration) {
	r.Store(metricName, hash, labels, vec, o, metrics.SummaryMetricType, ttl)
}

func (r *Registry) Store(metricName string, hash metrics.LabelHash, labels prometheus.Labels, vh metrics.VectorHolder, mh metrics.MetricHolder, metricType metrics.MetricType, ttl time.Duration) {
	metric, hasMetrics := r.Metrics[metricName]
	if !hasMetrics {
		metric.MetricType = metricType
		metric.Vectors = make(map[metrics.NameHash]*metrics.Vector)
		metric.Metrics = make(map[metrics.ValueHash]*metrics.RegisteredMetric)

		r.Metrics[metricName] = metric
	}

	v, ok := metric.Vectors[hash.Names]
	if !ok {
		v = &metrics.Vector{Holder: vh}
		metric.Vectors[hash.Names] = v
	}

	now := clock.Now()
	rm, ok := metric.Metrics[hash.Values]
	if !ok {
		rm = &metrics.RegisteredMetric{
			LastRegisteredAt: now,
			Labels:           labels,
			TTL:              ttl,
			Metric:           mh,
			VecKey:           hash.Names,
		}
		metric.Metrics[hash.Values] = rm
		v.RefCount++
		return
	}
	rm.LastRegisteredAt = now
	// Update ttl from mapping
	rm.TTL = ttl
}

func (r *Registry) Get(metricName string, hash metrics.LabelHash, metricType metrics.MetricType) (metrics.VectorHolder, metrics.MetricHolder) {
	metric, hasMetric := r.Metrics[metricName]

	if !hasMetric {
		return nil, nil
	}
	if metric.MetricType != metricType {
		return nil, nil
	}

	rm, ok := metric.Metrics[hash.Values]
	if ok {
		now := clock.Now()
		rm.LastRegisteredAt = now
		return metric.Vectors[hash.Names].Holder, rm.Metric
	}

	vector, ok := metric.Vectors[hash.Names]
	if ok {
		return vector.Holder, nil
	}

	return nil, nil
}

func (r *Registry) GetCounter(metricName string, labels prometheus.Labels, help string, mapping *mapper.MetricMapping, metricsCount *prometheus.GaugeVec) (prometheus.Counter, error) {
	hash, labelNames := r.HashLabels(labels)
	vh, mh := r.Get(metricName, hash, metrics.CounterMetricType)
	if mh != nil {
		return mh.(prometheus.Counter), nil
	}

	if r.MetricConflicts(metricName, metrics.CounterMetricType) {
		return nil, fmt.Errorf("metric with name %s is already registered", metricName)
	}

	var counterVec *prometheus.CounterVec
	if vh == nil {
		metricsCount.WithLabelValues("counter").Inc()
		counterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: metricName,
			Help: help,
		}, labelNames)

		if err := r.Registerer.Register(uncheckedCollector{counterVec}); err != nil {
			return nil, err
		}
	} else {
		counterVec = vh.(*prometheus.CounterVec)
	}

	var counter prometheus.Counter
	var err error
	if counter, err = counterVec.GetMetricWith(labels); err != nil {
		return nil, err
	}
	r.StoreCounter(metricName, hash, labels, counterVec, counter, mapping.Ttl)

	return counter, nil
}

func (r *Registry) GetGauge(metricName string, labels prometheus.Labels, help string, mapping *mapper.MetricMapping, metricsCount *prometheus.GaugeVec) (prometheus.Gauge, error) {
	hash, labelNames := r.HashLabels(labels)
	vh, mh := r.Get(metricName, hash, metrics.GaugeMetricType)
	if mh != nil {
		return mh.(prometheus.Gauge), nil
	}

	if r.MetricConflicts(metricName, metrics.GaugeMetricType) {
		return nil, fmt.Errorf("metrics.Metric with name %s is already registered", metricName)
	}

	var gaugeVec *prometheus.GaugeVec
	if vh == nil {
		metricsCount.WithLabelValues("gauge").Inc()
		gaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: metricName,
			Help: help,
		}, labelNames)

		if err := r.Registerer.Register(uncheckedCollector{gaugeVec}); err != nil {
			return nil, err
		}
	} else {
		gaugeVec = vh.(*prometheus.GaugeVec)
	}

	var gauge prometheus.Gauge
	var err error
	if gauge, err = gaugeVec.GetMetricWith(labels); err != nil {
		return nil, err
	}
	r.StoreGauge(metricName, hash, labels, gaugeVec, gauge, mapping.Ttl)

	return gauge, nil
}

func (r *Registry) GetHistogram(metricName string, labels prometheus.Labels, help string, mapping *mapper.MetricMapping, metricsCount *prometheus.GaugeVec) (prometheus.Observer, error) {
	hash, labelNames := r.HashLabels(labels)
	vh, mh := r.Get(metricName, hash, metrics.HistogramMetricType)
	if mh != nil {
		return mh.(prometheus.Observer), nil
	}

	if r.MetricConflicts(metricName, metrics.HistogramMetricType) {
		return nil, fmt.Errorf("metrics.Metric with name %s is already registered", metricName)
	}
	if r.MetricConflicts(metricName+"_sum", metrics.HistogramMetricType) {
		return nil, fmt.Errorf("metrics.Metric with name %s is already registered", metricName)
	}
	if r.MetricConflicts(metricName+"_count", metrics.HistogramMetricType) {
		return nil, fmt.Errorf("metrics.Metric with name %s is already registered", metricName)
	}
	if r.MetricConflicts(metricName+"_bucket", metrics.HistogramMetricType) {
		return nil, fmt.Errorf("metrics.Metric with name %s is already registered", metricName)
	}

	var histogramVec *prometheus.HistogramVec
	if vh == nil {
		metricsCount.WithLabelValues("histogram").Inc()
		buckets := r.Mapper.Defaults.Buckets
		if mapping.HistogramOptions != nil && len(mapping.HistogramOptions.Buckets) > 0 {
			buckets = mapping.HistogramOptions.Buckets
		}
		histogramVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    metricName,
			Help:    help,
			Buckets: buckets,
		}, labelNames)

		if err := prometheus.Register(uncheckedCollector{histogramVec}); err != nil {
			return nil, err
		}
	} else {
		histogramVec = vh.(*prometheus.HistogramVec)
	}

	var observer prometheus.Observer
	var err error
	if observer, err = histogramVec.GetMetricWith(labels); err != nil {
		return nil, err
	}
	r.StoreHistogram(metricName, hash, labels, histogramVec, observer, mapping.Ttl)

	return observer, nil
}

func (r *Registry) GetSummary(metricName string, labels prometheus.Labels, help string, mapping *mapper.MetricMapping, metricsCount *prometheus.GaugeVec) (prometheus.Observer, error) {
	hash, labelNames := r.HashLabels(labels)
	vh, mh := r.Get(metricName, hash, metrics.SummaryMetricType)
	if mh != nil {
		return mh.(prometheus.Observer), nil
	}

	if r.MetricConflicts(metricName, metrics.SummaryMetricType) {
		return nil, fmt.Errorf("metrics.Metric with name %s is already registered", metricName)
	}
	if r.MetricConflicts(metricName+"_sum", metrics.SummaryMetricType) {
		return nil, fmt.Errorf("metrics.Metric with name %s is already registered", metricName)
	}
	if r.MetricConflicts(metricName+"_count", metrics.SummaryMetricType) {
		return nil, fmt.Errorf("metrics.Metric with name %s is already registered", metricName)
	}

	var summaryVec *prometheus.SummaryVec
	if vh == nil {
		metricsCount.WithLabelValues("summary").Inc()
		quantiles := r.Mapper.Defaults.Quantiles
		if mapping != nil && mapping.SummaryOptions != nil && len(mapping.SummaryOptions.Quantiles) > 0 {
			quantiles = mapping.SummaryOptions.Quantiles
		}
		summaryOptions := mapper.SummaryOptions{}
		if mapping != nil && mapping.SummaryOptions != nil {
			summaryOptions = *mapping.SummaryOptions
		}
		objectives := make(map[float64]float64)
		for _, q := range quantiles {
			objectives[q.Quantile] = q.Error
		}
		// In the case of no mapping file, explicitly define the default quantiles
		if len(objectives) == 0 {
			objectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
		}
		summaryVec = prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name:       metricName,
			Help:       help,
			Objectives: objectives,
			MaxAge:     summaryOptions.MaxAge,
			AgeBuckets: summaryOptions.AgeBuckets,
			BufCap:     summaryOptions.BufCap,
		}, labelNames)

		if err := prometheus.Register(uncheckedCollector{summaryVec}); err != nil {
			return nil, err
		}
	} else {
		summaryVec = vh.(*prometheus.SummaryVec)
	}

	var observer prometheus.Observer
	var err error
	if observer, err = summaryVec.GetMetricWith(labels); err != nil {
		return nil, err
	}
	r.StoreSummary(metricName, hash, labels, summaryVec, observer, mapping.Ttl)

	return observer, nil
}

func (r *Registry) RemoveStaleMetrics() {
	now := clock.Now()
	// delete timeseries with expired ttl
	for _, metric := range r.Metrics {
		for hash, rm := range metric.Metrics {
			if rm.TTL == 0 {
				continue
			}
			if rm.LastRegisteredAt.Add(rm.TTL).Before(now) {
				metric.Vectors[rm.VecKey].Holder.Delete(rm.Labels)
				metric.Vectors[rm.VecKey].RefCount--
				delete(metric.Metrics, hash)
			}
		}
	}
}

// Calculates a hash of both the label names and the label names and values.
func (r *Registry) HashLabels(labels prometheus.Labels) (metrics.LabelHash, []string) {
	r.Hasher.Reset()
	r.NameBuf.Reset()
	r.ValueBuf.Reset()
	labelNames := make([]string, 0, len(labels))

	for labelName := range labels {
		labelNames = append(labelNames, labelName)
	}
	sort.Strings(labelNames)

	r.ValueBuf.WriteByte(model.SeparatorByte)
	for _, labelName := range labelNames {
		r.ValueBuf.WriteString(labels[labelName])
		r.ValueBuf.WriteByte(model.SeparatorByte)

		r.NameBuf.WriteString(labelName)
		r.NameBuf.WriteByte(model.SeparatorByte)
	}

	lh := metrics.LabelHash{}
	r.Hasher.Write(r.NameBuf.Bytes())
	lh.Names = metrics.NameHash(r.Hasher.Sum64())

	// Now add the values to the names we've already hashed.
	r.Hasher.Write(r.ValueBuf.Bytes())
	lh.Values = metrics.ValueHash(r.Hasher.Sum64())

	return lh, labelNames
}
