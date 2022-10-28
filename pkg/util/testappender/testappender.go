// Package testappender exposes utilities to test code which writes to
// Prometheus storage.Appenders.
package testappender

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/model/textparse"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/utils/pointer"
)

// Appender implements storage.Appender. It keeps track of samples, metadata,
// and exemplars written to it.
//
// When Commit is called, the written data will be converted into a slice of
// *dto.MetricFamily, when can then be used for asserting against expectations
// in tests.
//
// The zero value of Appender is ready for use. Appender is only intended for
// test code, and is not optimized for production.
//
// Appender is not safe for concurrent use.
type Appender struct {
	commitCalled, rollbackCalled bool

	samples   map[string]sample            // metric labels -> sample
	exemplars map[string]seriesExemplar    // metric labels -> series exemplar
	metadata  map[string]metadata.Metadata // metric family name -> metadata

	families []*dto.MetricFamily
}

type seriesExemplar struct {
	// Labels is the labels of the series exposing the exemplar, not the labels
	// on the exemplar itself.
	Labels   labels.Labels
	Exemplar exemplar.Exemplar
}

type sample struct {
	Labels    labels.Labels
	Timestamp int64
	Value     float64
}

var _ storage.Appender = (*Appender)(nil)

func (app *Appender) init() {
	if app.samples == nil {
		app.samples = make(map[string]sample)
	}
	if app.exemplars == nil {
		app.exemplars = make(map[string]seriesExemplar)
	}
	if app.metadata == nil {
		app.metadata = make(map[string]metadata.Metadata)
	}
}

// Append adds or updates a sample for a given metric, identified by labels. l
// must not be empty. If Append is called twice for the same metric, older
// samples are discarded.
//
// Upon calling Commit, a MetricFamily is created for each unique `__name__`
// label. If UpdateMetadata is not called for the named series, the series will
// be treated as untyped. The timestamp of the metric will be reported with the
// value denoted by t.
//
// The ref field is ignored, and Append always returns 0 for the resulting
// storage.SeriesRef.
func (app *Appender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	if app.commitCalled || app.rollbackCalled {
		return 0, fmt.Errorf("appender is closed")
	}
	app.init()

	l = l.WithoutEmpty()
	if len(l) == 0 {
		return 0, fmt.Errorf("empty labelset: %w", tsdb.ErrInvalidSample)
	}
	if lbl, dup := l.HasDuplicateLabelNames(); dup {
		return 0, fmt.Errorf("label name %q is not unique: %w", lbl, tsdb.ErrInvalidSample)
	}

	app.samples[l.String()] = sample{
		Labels:    l,
		Timestamp: t,
		Value:     v,
	}
	return 0, nil
}

// AppendExemplar adds an exemplar for a given metric, identified by lablels. l
// must not be empty.
//
// Upon calling Commit, exemplars are injected into the resulting Metrics for
// any Counter or Histogram.
//
// The ref field is ignored, and AppendExemplar always returns 0 for the
// resulting storage.SeriesRef.
func (app *Appender) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	if app.commitCalled || app.rollbackCalled {
		return 0, fmt.Errorf("appender is closed")
	}
	app.init()

	l = l.WithoutEmpty()
	if len(l) == 0 {
		return 0, fmt.Errorf("empty labelset: %w", tsdb.ErrInvalidSample)
	}
	if lbl, dup := l.HasDuplicateLabelNames(); dup {
		return 0, fmt.Errorf("label name %q is not unique: %w", lbl, tsdb.ErrInvalidSample)
	}

	app.exemplars[l.String()] = seriesExemplar{
		Labels:   l,
		Exemplar: e,
	}
	return 0, nil
}

// UpdateMetadata associates metadata for a given named metric. l must not be
// empty. Only the `__name__` label is used from the label set; other labels
// are ignored.
//
// Upon calling Commit, the metadata will be injected into the associated
// MetricFamily. If m represents a histogram, metrics suffixed with `_bucket`,
// `_sum`, and `_count` will be brought into the same MetricFamily.
func (app *Appender) UpdateMetadata(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata) (storage.SeriesRef, error) {
	if app.commitCalled || app.rollbackCalled {
		return 0, fmt.Errorf("appender is closed")
	}
	app.init()

	l = l.WithoutEmpty()
	if len(l) == 0 {
		return 0, fmt.Errorf("empty labelset: %w", tsdb.ErrInvalidSample)
	}
	if lbl, dup := l.HasDuplicateLabelNames(); dup {
		return 0, fmt.Errorf("label name %q is not unique: %w", lbl, tsdb.ErrInvalidSample)
	}

	// Metadata is associated with just the metric family, retrieved by the
	// __name__ label. All metrics for that same name always have the same
	// metadata.
	familyName := l.Get(model.MetricNameLabel)
	if familyName == "" {
		return 0, fmt.Errorf("__name__ label missing: %w", tsdb.ErrInvalidSample)
	}
	app.metadata[familyName] = m
	return 0, nil
}

// Commit commits pending samples, exemplars, and metadata, converting them
// into a slice of *dto.MetricsFamily. Call MetricFamlies to get the resulting
// data.
//
// After calling Commit, no other methods except MetricFamlies may be called.
func (app *Appender) Commit() error {
	if app.commitCalled || app.rollbackCalled {
		return fmt.Errorf("appender is closed")
	}

	var (
		families     []*dto.MetricFamily
		familyLookup = make(map[string]*dto.MetricFamily)
	)

	// First, iterate over the metadata to prepopulate the MetricFamily list.
	// This will help us determine what inner types to create for samples.
	for familyName, m := range app.metadata {
		var mt dto.MetricType
		switch m.Type {
		case textparse.MetricTypeCounter:
			mt = dto.MetricType_COUNTER
		case textparse.MetricTypeGauge:
			mt = dto.MetricType_GAUGE
		case textparse.MetricTypeHistogram:
			mt = dto.MetricType_HISTOGRAM
		case textparse.MetricTypeSummary:
			mt = dto.MetricType_SUMMARY
		default:
			// There are other values for m.Type, but they're all
			// OpenMetrics-specific and we're only converting into the Prometheus
			// exposition format.
			mt = dto.MetricType_UNTYPED
		}

		mf := &dto.MetricFamily{
			Name: pointer.String(familyName),
			Type: &mt,
		}

		if m.Help != "" {
			mf.Help = &m.Help
		}

		families = append(families, mf)

		// Determine how to populate the lookup table.
		switch mt {
		case dto.MetricType_SUMMARY:
			// Summaries include metrics with the family name (for quantiles),
			// followed by _sum and _count suffixes.
			familyLookup[familyName] = mf
			familyLookup[familyName+"_sum"] = mf
			familyLookup[familyName+"_count"] = mf
		case dto.MetricType_HISTOGRAM:
			// Histograms include metrics for _bucket, _sum, and _count suffixes.
			familyLookup[familyName+"_bucket"] = mf
			familyLookup[familyName+"_sum"] = mf
			familyLookup[familyName+"_count"] = mf
		default:
			// Everything else matches the family name exactly.
			familyLookup[familyName] = mf
		}
	}

	// Next, iterate over all samples and add them to metrics in the appropriate
	// MetricFamily. A new MetricFamily is created if one doesn't exist.
	for _, sample := range app.samples {
		// Get or create the metric family.
		metricName := sample.Labels.Get(model.MetricNameLabel)

		mf, ok := familyLookup[metricName]
		if !ok {
			mt := dto.MetricType_UNTYPED
			mf = &dto.MetricFamily{
				Name: &metricName,
				Type: &mt,
			}
			families = append(families, mf)
			familyLookup[metricName] = mf
		}

		// Retrieve the sample based on labels.
		m := getOrCreateMetric(mf, sample.Labels)
		m.TimestampMs = pointer.Int64(sample.Timestamp)

		switch familyType(mf) {
		case dto.MetricType_COUNTER:
			m.Counter = &dto.Counter{
				Value: pointer.Float64(sample.Value),
			}

		case dto.MetricType_GAUGE:
			m.Gauge = &dto.Gauge{
				Value: pointer.Float64(sample.Value),
			}

		case dto.MetricType_SUMMARY:
			if m.Summary == nil {
				m.Summary = &dto.Summary{}
			}

			switch {
			case metricName == mf.GetName()+"_count":
				val := uint64(sample.Value)
				m.Summary.SampleCount = &val
			case metricName == mf.GetName()+"_sum":
				m.Summary.SampleSum = pointer.Float64(sample.Value)
			case metricName == mf.GetName():
				quantile, err := strconv.ParseFloat(sample.Labels.Get(model.QuantileLabel), 64)
				if err != nil {
					continue
				}

				m.Summary.Quantile = append(m.Summary.Quantile, &dto.Quantile{
					Quantile: &quantile,
					Value:    pointer.Float64(sample.Value),
				})
			}

		case dto.MetricType_UNTYPED:
			m.Untyped = &dto.Untyped{
				Value: pointer.Float64(sample.Value),
			}

		case dto.MetricType_HISTOGRAM:
			if m.Histogram == nil {
				m.Histogram = &dto.Histogram{}
			}

			switch {
			case metricName == mf.GetName()+"_count":
				val := uint64(sample.Value)
				m.Histogram.SampleCount = &val
			case metricName == mf.GetName()+"_sum":
				m.Histogram.SampleSum = pointer.Float64(sample.Value)
			case metricName == mf.GetName()+"_bucket":
				boundary, err := strconv.ParseFloat(sample.Labels.Get(model.BucketLabel), 64)
				if err != nil {
					continue
				}

				count := uint64(sample.Value)

				m.Histogram.Bucket = append(m.Histogram.Bucket, &dto.Bucket{
					UpperBound:      &boundary,
					CumulativeCount: &count,
				})
			}
		}
	}

	// Finally, iterate through exemplars and attach them where appropriate.
	for _, e := range app.exemplars {
		// Get or create the metric family.
		exemplarName := e.Labels.Get(model.MetricNameLabel)

		mf, ok := familyLookup[exemplarName]
		if !ok {
			// No metric family (which means no corresponding sample; ignore)
			continue
		}

		m := getMetric(mf, e.Labels)
		if m == nil {
			continue
		}

		switch familyType(mf) {
		case dto.MetricType_COUNTER:
			if m.Counter == nil {
				// Sample never added?
				continue
			}
			m.Counter.Exemplar = convertExemplar(dto.MetricType_COUNTER, e.Exemplar)
		case dto.MetricType_HISTOGRAM:
			if m.Histogram == nil {
				// Sample never added?
				continue
			}

			switch {
			case exemplarName == mf.GetName()+"_bucket":
				boundary, err := strconv.ParseFloat(e.Labels.Get(model.BucketLabel), 64)
				if err != nil {
					continue
				}

				bucket := findBucket(m.Histogram, boundary)
				if bucket == nil {
					continue
				}
				bucket.Exemplar = convertExemplar(dto.MetricType_HISTOGRAM, e.Exemplar)
			default:
				// Exemplars only supported on buckets
				continue
			}

		default:
			// Exemplars not supported for other types; ignore.
			continue
		}
	}

	// Finally, sort all the DTO data.
	sortMetricFamilies(families)

	app.commitCalled = true
	app.families = families
	return nil
}

func familyType(mf *dto.MetricFamily) dto.MetricType {
	ty := mf.Type
	if ty == nil {
		return dto.MetricType_UNTYPED
	}
	return *ty
}

func getMetric(mf *dto.MetricFamily, l labels.Labels) *dto.Metric {
	metricLabels := toLabelPairs(familyType(mf), l)

	for _, check := range mf.Metric {
		if labelPairsEqual(check.Label, metricLabels) {
			return check
		}
	}

	return nil
}

func getOrCreateMetric(mf *dto.MetricFamily, l labels.Labels) *dto.Metric {
	metricLabels := toLabelPairs(familyType(mf), l)

	for _, check := range mf.Metric {
		if labelPairsEqual(check.Label, metricLabels) {
			return check
		}
	}

	m := &dto.Metric{
		Label: metricLabels,
	}
	mf.Metric = append(mf.Metric, m)
	return m
}

func convertExemplar(mt dto.MetricType, e exemplar.Exemplar) *dto.Exemplar {
	res := &dto.Exemplar{
		Label: toLabelPairs(mt, e.Labels),
		Value: &e.Value,
	}
	if e.HasTs {
		res.Timestamp = timestamppb.New(time.UnixMilli(e.Ts))
	}
	return res
}

// toLabelPairs converts labels.Labels into []*dto.LabelPair. The __name__
// label is always dropped, since the metric name is retrieved from the family
// name instead.
//
// The quantile label is dropped for summaries, and the le label is dropped for
// histograms.
func toLabelPairs(mt dto.MetricType, ls labels.Labels) []*dto.LabelPair {
	res := make([]*dto.LabelPair, 0, len(ls))
	for _, l := range ls {
		if l.Name == model.MetricNameLabel {
			continue
		} else if l.Name == model.QuantileLabel && mt == dto.MetricType_SUMMARY {
			continue
		} else if l.Name == model.BucketLabel && mt == dto.MetricType_HISTOGRAM {
			continue
		}

		res = append(res, &dto.LabelPair{
			Name:  pointer.String(l.Name),
			Value: pointer.String(l.Value),
		})
	}

	sort.Slice(res, func(i, j int) bool {
		switch {
		case *res[i].Name < *res[j].Name:
			return true
		case *res[i].Value < *res[j].Value:
			return true
		default:
			return false
		}
	})
	return res
}

func labelPairsEqual(a, b []*dto.LabelPair) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if *a[i].Name != *b[i].Name || *a[i].Value != *b[i].Value {
			return false
		}
	}

	return true
}

// Rollback discards pending samples, exemplars, and metadata.
//
// After calling Rollback, no other methods may be called and the Appender must
// be discarded.
func (app *Appender) Rollback() error {
	if app.commitCalled || app.rollbackCalled {
		return fmt.Errorf("appender is closed")
	}

	app.rollbackCalled = true
	return nil
}

// MetricFamilies returns the generated slice of *dto.MetricsFamily.
// MetricFamilies returns an error unless Commit was called.
//
// MetricFamilies always returns a non-nil slice. If no data was appended, the
// resulting slice has a length of zero.
func (app *Appender) MetricFamilies() ([]*dto.MetricFamily, error) {
	if !app.commitCalled {
		return nil, fmt.Errorf("MetricFamilies is not ready")
	} else if app.rollbackCalled {
		return nil, fmt.Errorf("appender is closed")
	}

	return app.families, nil
}

func findBucket(h *dto.Histogram, bound float64) *dto.Bucket {
	for _, b := range h.GetBucket() {
		// If it's close enough, use the bucket.
		if math.Abs(b.GetUpperBound()-bound) < 1e-9 {
			return b
		}
	}

	return nil
}
