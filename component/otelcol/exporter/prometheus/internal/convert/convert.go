// Package convert implements conversion utilities to convert between
// OpenTelemetry Collector data and Prometheus data.
//
// It follows the [OpenTelemetry Metrics Data Model] for implementing the
// conversion.
//
// [OpenTelemetry Metrics Data Model]: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/metrics/data-model.md
package convert

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/model/textparse"
	"github.com/prometheus/prometheus/model/value"
	"github.com/prometheus/prometheus/storage"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	semconv "go.opentelemetry.io/collector/semconv/v1.6.1"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus"
)

var (
	scopeNameLabel    = "otel_scope_name"
	scopeVersionLabel = "otel_scope_version"
)

// TODO(rfratto): Exemplars are not currently supported.

// Converter implements consumer.Metrics and converts received metrics
// into Prometheus-compatible metrics.
type Converter struct {
	log log.Logger

	optsMut sync.RWMutex
	opts    Options

	seriesCache   sync.Map // Cache of active series.
	metadataCache sync.Map // Cache of active metadata entries.

	next storage.Appendable // Location to write converted metrics.
}

// Options configure a Converter.
type Options struct {
	// IncludeTargetInfo includes the target_info metric.
	IncludeTargetInfo bool
	// IncludeScopeInfo includes the otel_scope_info metric and adds
	// otel_scope_name and otel_scope_version labels to data points.
	IncludeScopeInfo bool
}

var _ consumer.Metrics = (*Converter)(nil)

// New returns a new Converter. Converted metrics are passed to the provided
// storage.Appendable implementation.
func New(l log.Logger, next storage.Appendable, opts Options) *Converter {
	if l == nil {
		l = log.NewNopLogger()
	}
	return &Converter{log: l, next: next, opts: opts}
}

// UpdateOptions updates the options for the Converter.
func (conv *Converter) UpdateOptions(opts Options) {
	conv.optsMut.Lock()
	defer conv.optsMut.Unlock()
	conv.opts = opts
}

// getOpts gets a copy of the current options for the Converter.
func (conv *Converter) getOpts() Options {
	conv.optsMut.RLock()
	defer conv.optsMut.RUnlock()
	return conv.opts
}

// Capabilities implements consumer.Metrics.
func (conv *Converter) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{
		MutatesData: false,
	}
}

// ConsumeMetrics converts the provided OpenTelemetry Collector-formatted
// metrics into Prometheus-compatible metrics. Each call to ConsumeMetrics
// requests a storage.Appender and will commit generated metrics to it at the
// end of the call.
//
// Metrics are tracked in memory over time. Call [*Converter.GC] to clean up
// stale metrics.
func (conv *Converter) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	// NOTE(rfratto): OpenTelemetry Collector doesn't have any equivalent concept
	// of storage.SeriesRef from Prometheus. This adds some extra CPU overhead in
	// converting pmetric.Metrics to Prometheus data, since we'll always have to
	// build a key to uniquely represent each data point.
	//
	// To reduce CPU and allocations as much as possible, each datapoint is
	// tracked as an "active series." See memorySeries for information on what is
	// cached.

	app := conv.next.Appender(ctx)

	for rcount := 0; rcount < md.ResourceMetrics().Len(); rcount++ {
		rm := md.ResourceMetrics().At(rcount)
		conv.consumeResourceMetrics(app, rm)
	}

	return app.Commit()
}

func (conv *Converter) consumeResourceMetrics(app storage.Appender, rm pmetric.ResourceMetrics) {
	resourceMD := conv.createOrUpdateMetadata("target_info", metadata.Metadata{
		Type: textparse.MetricTypeGauge,
		Help: "Target metadata",
	})
	memResource := conv.getOrCreateResource(rm.Resource())

	if conv.getOpts().IncludeTargetInfo {
		if err := resourceMD.WriteTo(app, time.Now()); err != nil {
			level.Warn(conv.log).Log("msg", "failed to write target_info metadata", "err", err)
		}
		if err := memResource.WriteTo(app, time.Now()); err != nil {
			level.Error(conv.log).Log("msg", "failed to write target_info metric", "err", err)
		}
	}

	for smcount := 0; smcount < rm.ScopeMetrics().Len(); smcount++ {
		sm := rm.ScopeMetrics().At(smcount)
		conv.consumeScopeMetrics(app, memResource, sm)
	}
}

func (conv *Converter) createOrUpdateMetadata(name string, md metadata.Metadata) *memoryMetadata {
	entry := &memoryMetadata{
		Name: name,
	}
	if actual, loaded := conv.metadataCache.LoadOrStore(name, entry); loaded {
		entry = actual.(*memoryMetadata)
	}

	entry.Update(md)
	return entry
}

// getOrCreateResource gets or creates a [*memorySeries] from the provided
// res. The LastSeen field of the *memorySeries is updated before returning.
func (conv *Converter) getOrCreateResource(res pcommon.Resource) *memorySeries {
	targetInfoLabels := labels.FromStrings(model.MetricNameLabel, "target_info")

	var (
		// There is no need to sort the attributes here.
		// The call to lb.Labels below will sort them.
		attrs = res.Attributes()

		jobLabel      string
		instanceLabel string
	)

	if serviceName, ok := attrs.Get(semconv.AttributeServiceName); ok {
		if serviceNamespace, ok := attrs.Get(semconv.AttributeServiceNamespace); ok {
			jobLabel = fmt.Sprintf("%s/%s", serviceNamespace.AsString(), serviceName.AsString())
		} else {
			jobLabel = serviceName.AsString()
		}
	}

	if instanceID, ok := attrs.Get(semconv.AttributeServiceInstanceID); ok {
		instanceLabel = instanceID.AsString()
	}

	lb := labels.NewBuilder(targetInfoLabels)
	lb.Set(model.JobLabel, jobLabel)
	lb.Set(model.InstanceLabel, instanceLabel)

	attrs.Range(func(k string, v pcommon.Value) bool {
		// Skip attributes that we used for determining the job and instance
		// labels.
		if k == semconv.AttributeServiceName ||
			k == semconv.AttributeServiceNamespace ||
			k == semconv.AttributeServiceInstanceID {

			return true
		}

		lb.Set(prometheus.NormalizeLabel(k), v.AsString())
		return true
	})

	labels := lb.Labels(nil)

	entry := newMemorySeries(map[string]string{
		model.JobLabel:      jobLabel,
		model.InstanceLabel: instanceLabel,
	}, labels)
	if actual, loaded := conv.seriesCache.LoadOrStore(labels.String(), entry); loaded {
		entry = actual.(*memorySeries)
	}

	entry.SetValue(1)
	entry.Ping()
	return entry
}

func (conv *Converter) consumeScopeMetrics(app storage.Appender, memResource *memorySeries, sm pmetric.ScopeMetrics) {
	scopeMD := conv.createOrUpdateMetadata("otel_scope_info", metadata.Metadata{
		Type: textparse.MetricTypeGauge,
	})
	memScope := conv.getOrCreateScope(memResource, sm.Scope())

	if conv.getOpts().IncludeScopeInfo {
		if err := scopeMD.WriteTo(app, time.Now()); err != nil {
			level.Warn(conv.log).Log("msg", "failed to write otel_scope_info metadata", "err", err)
		}
		if err := memScope.WriteTo(app, time.Now()); err != nil {
			level.Error(conv.log).Log("msg", "failed to write otel_scope_info metric", "err", err)
		}
	}

	for mcount := 0; mcount < sm.Metrics().Len(); mcount++ {
		m := sm.Metrics().At(mcount)
		conv.consumeMetric(app, memResource, memScope, m)
	}
}

// getOrCreateScope gets or creates a [*memorySeries] from the provided scope.
// The LastSeen field of the *memorySeries is updated before returning.
func (conv *Converter) getOrCreateScope(res *memorySeries, scope pcommon.InstrumentationScope) *memorySeries {
	scopeInfoLabels := labels.FromStrings(
		model.MetricNameLabel, "otel_scope_info",
		model.JobLabel, res.metadata[model.JobLabel],
		model.InstanceLabel, res.metadata[model.InstanceLabel],
		"name", scope.Name(),
		"version", scope.Version(),
	)

	lb := labels.NewBuilder(scopeInfoLabels)
	// There is no need to sort the attributes here.
	// The call to lb.Labels below will sort them.
	scope.Attributes().Range(func(k string, v pcommon.Value) bool {
		lb.Set(prometheus.NormalizeLabel(k), v.AsString())
		return true
	})

	labels := lb.Labels(nil)

	entry := newMemorySeries(map[string]string{
		scopeNameLabel:    scope.Name(),
		scopeVersionLabel: scope.Version(),
	}, labels)
	if actual, loaded := conv.seriesCache.LoadOrStore(labels.String(), entry); loaded {
		entry = actual.(*memorySeries)
	}

	entry.SetValue(1)
	entry.Ping()
	return entry
}

func (conv *Converter) consumeMetric(app storage.Appender, memResource *memorySeries, memScope *memorySeries, m pmetric.Metric) {
	switch m.Type() {
	case pmetric.MetricTypeGauge:
		conv.consumeGauge(app, memResource, memScope, m)
	case pmetric.MetricTypeSum:
		conv.consumeSum(app, memResource, memScope, m)
	case pmetric.MetricTypeHistogram:
		conv.consumeHistogram(app, memResource, memScope, m)
	case pmetric.MetricTypeSummary:
		conv.consumeSummary(app, memResource, memScope, m)
	}
}

func (conv *Converter) consumeGauge(app storage.Appender, memResource *memorySeries, memScope *memorySeries, m pmetric.Metric) {
	metricName := prometheus.BuildPromCompliantName(m, "")

	metricMD := conv.createOrUpdateMetadata(metricName, metadata.Metadata{
		Type: textparse.MetricTypeGauge,
		Unit: m.Unit(),
		Help: m.Description(),
	})
	if err := metricMD.WriteTo(app, time.Now()); err != nil {
		level.Warn(conv.log).Log("msg", "failed to write metric family metadata", "err", err)
	}

	for dpcount := 0; dpcount < m.Gauge().DataPoints().Len(); dpcount++ {
		dp := m.Gauge().DataPoints().At(dpcount)

		memSeries := conv.getOrCreateSeries(memResource, memScope, metricName, dp.Attributes())
		if err := writeSeries(app, memSeries, dp, getNumberDataPointValue(dp)); err != nil {
			level.Error(conv.log).Log("msg", "failed to write metric sample", "err", err)
		}
	}
}

type otelcolDataPoint interface {
	Timestamp() pcommon.Timestamp
	Flags() pmetric.DataPointFlags
}

func writeSeries(app storage.Appender, series *memorySeries, dp otelcolDataPoint, val float64) error {
	ts := dp.Timestamp().AsTime()
	if ts.Before(series.Timestamp()) {
		// Out-of-order; skip.
		return nil
	}
	series.SetTimestamp(ts)

	if dp.Flags().NoRecordedValue() {
		val = float64(value.StaleNaN)
	}
	series.SetValue(val)

	return series.WriteTo(app, ts)
}

// getOrCreateSeries gets or creates a [*memorySeries] from the provided
// resource, scope, metric, and attributes. The LastSeen field of the
// *memorySeries is updated before returning.
func (conv *Converter) getOrCreateSeries(res *memorySeries, scope *memorySeries, name string, attrs pcommon.Map, extraLabels ...labels.Label) *memorySeries {
	seriesBaseLabels := labels.FromStrings(
		model.MetricNameLabel, name,
		model.JobLabel, res.metadata[model.JobLabel],
		model.InstanceLabel, res.metadata[model.InstanceLabel],
	)

	lb := labels.NewBuilder(seriesBaseLabels)
	for _, extraLabel := range extraLabels {
		lb.Set(extraLabel.Name, extraLabel.Value)
	}

	if conv.getOpts().IncludeScopeInfo {
		lb.Set("otel_scope_name", scope.metadata[scopeNameLabel])
		lb.Set("otel_scope_version", scope.metadata[scopeVersionLabel])
	}

	// There is no need to sort the attributes here.
	// The call to lb.Labels below will sort them.
	attrs.Range(func(k string, v pcommon.Value) bool {
		lb.Set(prometheus.NormalizeLabel(k), v.AsString())
		return true
	})

	labels := lb.Labels(nil)

	entry := newMemorySeries(nil, labels)
	if actual, loaded := conv.seriesCache.LoadOrStore(labels.String(), entry); loaded {
		entry = actual.(*memorySeries)
	}

	entry.Ping()
	return entry
}

func getNumberDataPointValue(dp pmetric.NumberDataPoint) float64 {
	switch dp.ValueType() {
	case pmetric.NumberDataPointValueTypeDouble:
		return dp.DoubleValue()
	case pmetric.NumberDataPointValueTypeInt:
		return float64(dp.IntValue())
	}

	return 0
}

func (conv *Converter) consumeSum(app storage.Appender, memResource *memorySeries, memScope *memorySeries, m pmetric.Metric) {
	metricName := prometheus.BuildPromCompliantName(m, "")

	// Excerpt from the spec:
	//
	// * If the aggregation temporarlity is cumulative and sum is monotonic, it
	//   MUST be converted to a Prometheus Counter.
	// * If the aggregation temporarlity is cumulative and sum is non-monotonic,
	//   it MUST be converted to a Prometheus Gauge.
	// * If the aggregation temporarlity is delta and the sum is monotonic, it
	//   SHOULD be converted to a cumulative temporarlity and become a Prometheus
	//   Sum.
	// * Otherwise, it MUST be dropped.
	var convType textparse.MetricType
	switch {
	case m.Sum().AggregationTemporality() == pmetric.AggregationTemporalityCumulative && m.Sum().IsMonotonic():
		convType = textparse.MetricTypeCounter
	case m.Sum().AggregationTemporality() == pmetric.AggregationTemporalityCumulative && !m.Sum().IsMonotonic():
		convType = textparse.MetricTypeGauge
	case m.Sum().AggregationTemporality() == pmetric.AggregationTemporalityDelta && m.Sum().IsMonotonic():
		// Drop non-cumulative summaries for now, which is permitted by the spec.
		//
		// TODO(rfratto): implement delta-to-cumulative for sums.
		return
	default:
		// Drop the metric.
		return
	}

	metricMD := conv.createOrUpdateMetadata(metricName, metadata.Metadata{
		Type: convType,
		Unit: m.Unit(),
		Help: m.Description(),
	})
	if err := metricMD.WriteTo(app, time.Now()); err != nil {
		level.Warn(conv.log).Log("msg", "failed to write metric family metadata", "err", err)
	}

	for dpcount := 0; dpcount < m.Sum().DataPoints().Len(); dpcount++ {
		dp := m.Sum().DataPoints().At(dpcount)

		memSeries := conv.getOrCreateSeries(memResource, memScope, metricName, dp.Attributes())

		val := getNumberDataPointValue(dp)
		if err := writeSeries(app, memSeries, dp, val); err != nil {
			level.Error(conv.log).Log("msg", "failed to write metric sample", "err", err)
		}
	}
}

func (conv *Converter) consumeHistogram(app storage.Appender, memResource *memorySeries, memScope *memorySeries, m pmetric.Metric) {
	metricName := prometheus.BuildPromCompliantName(m, "")

	if m.Histogram().AggregationTemporality() != pmetric.AggregationTemporalityCumulative {
		// Drop non-cumulative histograms for now, which is permitted by the spec.
		//
		// TODO(rfratto): implement delta-to-cumulative for histograms.
		return
	}

	metricMD := conv.createOrUpdateMetadata(metricName, metadata.Metadata{
		Type: textparse.MetricTypeHistogram,
		Unit: m.Unit(),
		Help: m.Description(),
	})
	if err := metricMD.WriteTo(app, time.Now()); err != nil {
		level.Warn(conv.log).Log("msg", "failed to write metric family metadata", "err", err)
	}

	for dpcount := 0; dpcount < m.Histogram().DataPoints().Len(); dpcount++ {
		dp := m.Histogram().DataPoints().At(dpcount)

		// Sum metric
		if dp.HasSum() {
			sumMetric := conv.getOrCreateSeries(memResource, memScope, metricName+"_sum", dp.Attributes())
			sumMetricVal := dp.Sum()

			if err := writeSeries(app, sumMetric, dp, sumMetricVal); err != nil {
				level.Error(conv.log).Log("msg", "failed to write histogram sum sample", "err", err)
			}
		}

		// Count metric
		{
			countMetric := conv.getOrCreateSeries(memResource, memScope, metricName+"_count", dp.Attributes())
			countMetricVal := float64(dp.Count())

			if err := writeSeries(app, countMetric, dp, countMetricVal); err != nil {
				level.Error(conv.log).Log("msg", "failed to write histogram count sample", "err", err)
			}
		}

		// Process the boundaries. The number of buckets = number of explicit
		// bounds + 1.
		for i := 0; i < dp.ExplicitBounds().Len() && i < dp.BucketCounts().Len(); i++ {
			bound := dp.ExplicitBounds().At(i)
			count := dp.BucketCounts().At(i)

			bucketLabel := labels.Label{
				Name:  model.BucketLabel,
				Value: strconv.FormatFloat(bound, 'f', -1, 64),
			}

			bucket := conv.getOrCreateSeries(memResource, memScope, metricName+"_bucket", dp.Attributes(), bucketLabel)
			bucketVal := float64(count)

			if err := writeSeries(app, bucket, dp, bucketVal); err != nil {
				level.Error(conv.log).Log("msg", "failed to write histogram bucket sample", "bucket", bucketLabel.Value, "err", err)
			}
		}

		// Add le=+Inf bucket. All values are <= +Inf, so the value is the same as
		// the count of the datapoint.
		{
			bucketLabel := labels.Label{
				Name:  model.BucketLabel,
				Value: "+Inf",
			}

			infBucket := conv.getOrCreateSeries(memResource, memScope, metricName+"_bucket", dp.Attributes(), bucketLabel)
			infBucketVal := float64(dp.Count())

			if err := writeSeries(app, infBucket, dp, infBucketVal); err != nil {
				level.Error(conv.log).Log("msg", "failed to write histogram bucket sample", "bucket", bucketLabel.Value, "err", err)
			}
		}
	}
}

func (conv *Converter) consumeSummary(app storage.Appender, memResource *memorySeries, memScope *memorySeries, m pmetric.Metric) {
	metricName := prometheus.BuildPromCompliantName(m, "")

	metricMD := conv.createOrUpdateMetadata(metricName, metadata.Metadata{
		Type: textparse.MetricTypeSummary,
		Unit: m.Unit(),
		Help: m.Description(),
	})
	if err := metricMD.WriteTo(app, time.Now()); err != nil {
		level.Warn(conv.log).Log("msg", "failed to write metric family metadata", "err", err)
	}

	for dpcount := 0; dpcount < m.Summary().DataPoints().Len(); dpcount++ {
		dp := m.Summary().DataPoints().At(dpcount)

		// Sum metric
		{
			sumMetric := conv.getOrCreateSeries(memResource, memScope, metricName+"_sum", dp.Attributes())
			sumMetricVal := dp.Sum()

			if err := writeSeries(app, sumMetric, dp, sumMetricVal); err != nil {
				level.Error(conv.log).Log("msg", "failed to write summary sum sample", "err", err)
			}
		}

		// Count metric
		{
			countMetric := conv.getOrCreateSeries(memResource, memScope, metricName+"_count", dp.Attributes())
			countMetricVal := float64(dp.Count())

			if err := writeSeries(app, countMetric, dp, countMetricVal); err != nil {
				level.Error(conv.log).Log("msg", "failed to write histogram count sample", "err", err)
			}
		}

		// Quantiles
		for i := 0; i < dp.QuantileValues().Len(); i++ {
			qp := dp.QuantileValues().At(i)

			quantileLabel := labels.Label{
				Name:  model.QuantileLabel,
				Value: strconv.FormatFloat(qp.Quantile(), 'f', -1, 64),
			}

			quantile := conv.getOrCreateSeries(memResource, memScope, metricName, dp.Attributes(), quantileLabel)
			quantileVal := qp.Value()

			if err := writeSeries(app, quantile, dp, quantileVal); err != nil {
				level.Error(conv.log).Log("msg", "failed to write histogram quantile sample", "quantile", quantileLabel.Value, "err", err)
			}
		}
	}
}

// GC cleans up stale metrics which have not been updated in the time specified
// by staleTime.
func (conv *Converter) GC(staleTime time.Duration) {
	now := time.Now()

	// In the code below, we use TryLock as a small performance optimization.
	//
	// The garbage collector doesn't bother to wait for locks for anything in the
	// cache; the lock being unavailable implies that the cached resource is
	// still active.

	conv.seriesCache.Range(func(key, value any) bool {
		series := value.(*memorySeries)
		if !series.TryLock() {
			return true
		}
		defer series.Unlock()

		if now.Sub(series.lastSeen) > staleTime {
			conv.seriesCache.Delete(key)
		}
		return true
	})

	conv.metadataCache.Range(func(key, value any) bool {
		series := value.(*memoryMetadata)
		if !series.TryLock() {
			return true
		}
		defer series.Unlock()

		if now.Sub(series.lastSeen) > staleTime {
			conv.seriesCache.Delete(key)
		}
		return true
	})
}

// FlushMetadata empties out the metadata cache, forcing metadata to get
// rewritten.
func (conv *Converter) FlushMetadata() {
	// TODO(rfratto): this is fairly inefficient since it'll require rebuilding
	// all of the metadata for every active series. However, it's the easiest
	// thing to do for now.
	conv.metadataCache.Range(func(key, _ any) bool {
		conv.metadataCache.Delete(key)
		return true
	})
}
