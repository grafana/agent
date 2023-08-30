package simple

import (
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/storage/remote"
	"time"
)

// appender is used to transfer from incoming samples to the PebbleDB.
type appender struct {
	parent *Simple
	/*metrics         []prometheus.Sample
	exemplars       []prometheus.Exemplar
	histogram       []prometheus.Histogram
	floatHistograms []prometheus.FloatHistogram
	metadata        []prometheus.Metadata*/
	samples []prompb.TimeSeries
	ttl     time.Duration
}

func newAppender(parent *Simple, ttl time.Duration) *appender {
	return &appender{
		parent: parent,
		/*metrics:         make([]prometheus.Sample, 0),
		exemplars:       make([]prometheus.Exemplar, 0),
		histogram:       make([]prometheus.Histogram, 0),
		floatHistograms: make([]prometheus.FloatHistogram, 0),
		metadata:        make([]prometheus.Metadata, 0),*/
		samples: make([]prompb.TimeSeries, 0),
		ttl:     ttl,
	}
}

// Append metric
func (a *appender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	endTime := time.Now().UnixMilli() - int64(a.ttl.Seconds())
	if t < endTime {
		return ref, nil
	}
	sample := prompb.TimeSeries{
		Labels:     labelsToLabelsProto(l),
		Samples:    []prompb.Sample{{Value: v, Timestamp: t}},
		Exemplars:  nil,
		Histograms: nil,
	}
	a.samples = append(a.samples, sample)
	return ref, nil
}

// Commit metrics to the DB
func (a *appender) Commit() (_ error) {
	a.parent.commit(a)
	return nil
}

// Rollback does nothing.
func (a *appender) Rollback() error {
	// Since nothing is committed we dont need to do any cleanup here.
	return nil
}

// AppendExemplar appends exemplar to cache.
func (a *appender) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (_ storage.SeriesRef, _ error) {
	protoLabels := labelsToLabelsProto(l)
	sample := prompb.TimeSeries{
		Labels:     protoLabels,
		Samples:    nil,
		Exemplars:  []prompb.Exemplar{{Labels: labelsToLabelsProto(e.Labels), Value: e.Value, Timestamp: e.Ts}},
		Histograms: nil,
	}
	a.samples = append(a.samples, sample)
	return ref, nil
}

// AppendHistogram appends histogram
func (a *appender) AppendHistogram(ref storage.SeriesRef, l labels.Labels, t int64, h *histogram.Histogram, fh *histogram.FloatHistogram) (_ storage.SeriesRef, _ error) {
	endTime := time.Now().UnixMilli() - int64(a.ttl.Seconds())
	if t < endTime {
		return ref, nil
	}

	if h != nil {
		sample := prompb.TimeSeries{
			Labels:     labelsToLabelsProto(l),
			Samples:    nil,
			Exemplars:  nil,
			Histograms: []prompb.Histogram{remote.HistogramToHistogramProto(t, h)},
		}
		a.samples = append(a.samples, sample)

	} else {
		sample := prompb.TimeSeries{
			Labels:     labelsToLabelsProto(l),
			Samples:    nil,
			Exemplars:  nil,
			Histograms: []prompb.Histogram{remote.FloatHistogramToHistogramProto(t, fh)},
		}
		a.samples = append(a.samples, sample)
	}
	return ref, nil
}

// UpdateMetadata updates metadata.
func (a *appender) UpdateMetadata(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata) (_ storage.SeriesRef, _ error) {
	// TODO allow metadata
	return 0, nil
}

var _ storage.Appender = (*appender)(nil)
