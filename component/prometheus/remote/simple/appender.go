package simple

import (
	"github.com/grafana/agent/component/prometheus"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/storage"
)

// appender is used to transfer from incoming samples to the PebbleDB.
type appender struct {
	parent          *Simple
	metrics         []prometheus.Sample
	exemplars       []prometheus.Exemplar
	histogram       []prometheus.Histogram
	floatHistograms []prometheus.FloatHistogram
	metadata        []prometheus.Metadata
}

func newAppender(parent *Simple) *appender {
	return &appender{
		parent:          parent,
		metrics:         make([]prometheus.Sample, 0),
		exemplars:       make([]prometheus.Exemplar, 0),
		histogram:       make([]prometheus.Histogram, 0),
		floatHistograms: make([]prometheus.FloatHistogram, 0),
		metadata:        make([]prometheus.Metadata, 0),
	}
}

// Append metric
func (a *appender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	a.metrics = append(a.metrics, prometheus.Sample{
		L:         l,
		Timestamp: t,
		Value:     v,
	})
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
	a.exemplars = append(a.exemplars, prometheus.Exemplar{
		Sample: prometheus.Sample{
			L:         l,
			Timestamp: e.Ts,
			Value:     e.Value,
		},
		L: l,
	})
	return ref, nil
}

// AppendHistogram appends histogram
func (a *appender) AppendHistogram(ref storage.SeriesRef, l labels.Labels, t int64, h *histogram.Histogram, fh *histogram.FloatHistogram) (_ storage.SeriesRef, _ error) {
	if h != nil {
		a.histogram = append(a.histogram, prometheus.Histogram{
			L:         l,
			Timestamp: t,
			Value:     h,
		})
	} else {
		a.floatHistograms = append(a.floatHistograms, prometheus.FloatHistogram{
			L:         l,
			Timestamp: t,
			Value:     fh,
		})
	}
	return ref, nil
}

// UpdateMetadata updates metadata.
func (a *appender) UpdateMetadata(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata) (_ storage.SeriesRef, _ error) {
	a.metadata = append(a.metadata, prometheus.Metadata{
		L:    l,
		Meta: m,
	})
	return ref, nil
}

var _ storage.Appender = (*appender)(nil)
