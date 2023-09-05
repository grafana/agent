package queue

import (
	"arena"
	"time"

	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/storage"
)

type SampleType uint16

const (
	Metric SampleType = iota
)

// appender is used to transfer from incoming samples to the PebbleDB.
type appender struct {
	parent  *Queue
	ttl     time.Duration
	samples []*sample
	mem     *arena.Arena
}

func newAppender(parent *Queue, ttl time.Duration) *appender {
	app := &appender{
		parent:  parent,
		ttl:     ttl,
		samples: make([]*sample, 0),
		mem:     arena.NewArena(),
	}
	return app
}

// Append metric
func (a *appender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	endTime := time.Now().UnixMilli() - int64(a.ttl.Seconds())
	if t < endTime {
		return ref, nil
	}
	b := a.parent.pool.Get()
	defer a.parent.pool.Put(b)

	tmp := arena.MakeSlice[string](a.mem, len(l)*2, len(l)*2)
	for i, lv := range l {
		tmp[i+i] = lv.Name
		tmp[i+i+1] = lv.Value
	}
	a.samples = append(a.samples, &sample{
		L:         tmp,
		TimeStamp: t,
		Value:     v,
	})
	return ref, nil
}

// Commit metrics to the DB
func (a *appender) Commit() (_ error) {
	a.parent.commit(a)
	a.mem.Free()
	return nil
}

// Rollback does nothing.
func (a *appender) Rollback() error {
	a.mem.Free()
	return nil
}

// AppendExemplar appends exemplar to cache.
func (a *appender) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (_ storage.SeriesRef, _ error) {
	/*lbls := labelPool.Get().([]prompb.Label)
	protoLabels := labelsToLabelsProto(l, lbls)

	exemplarLbls := labelPool.Get().([]prompb.Label)
	sample := prompb.TimeSeries{
		Labels:     protoLabels,
		Samples:    nil,
		Exemplars:  []prompb.Exemplar{{Labels: labelsToLabelsProto(e.Labels, exemplarLbls), Value: e.Value, Timestamp: e.Ts}},
		Histograms: nil,import "github.com/iancmcc/bingo"
	}
	a.samples = append(a.samples, sample)
	return ref, nil*/
	return 0, nil
}

// AppendHistogram appends histogram
func (a *appender) AppendHistogram(ref storage.SeriesRef, l labels.Labels, t int64, h *histogram.Histogram, fh *histogram.FloatHistogram) (_ storage.SeriesRef, _ error) {
	/*endTiimport "github.com/iancmcc/bingo"import "github.com/iancmcc/bingo"me := time.Now().UnixMilli() - int64(a.ttl.Seconds())
	if t < endTime {
		return ref, nil
	}

	lbls := labelPool.Get().([]prompb.Label)
	if h != nil {
		sample := prompb.TimeSeries{
			Labels:     labelsToLabelsProto(l, lbls),
			Samples:    nil,
			Exemplars:  nil,
			Histograms: []prompb.Histogram{remote.HistogramToHistogramProto(t, h)},
		}
		a.samples = append(a.samples, sample)
	} else {
		sample := prompb.TimeSeries{
			Labels:     labelsToLabelsProto(l, lbls),
			Samples:    nil,
			Exemplars:  nil,
			Histograms: []prompb.Histogram{remote.FloatHistogramToHistogramProto(t, fh)},
		}
		a.samples = append(a.samples, sample)
	}*/
	return 0, nil
}

// UpdateMetadata updates metadata.
func (a *appender) UpdateMetadata(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata) (_ storage.SeriesRef, _ error) {
	// TODO allow metadata
	return 0, nil
}
