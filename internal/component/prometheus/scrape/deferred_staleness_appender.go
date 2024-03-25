package scrape

import (
	"context"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/model/value"
	"github.com/prometheus/prometheus/storage"
)

type deferredStalenessAppendable struct {
	delegate storage.Appendable
	logger   log.Logger
}

func newDeferredStalenessAppendable(delegate storage.Appendable, logger log.Logger) storage.Appendable {
	return &deferredStalenessAppendable{
		delegate: delegate,
		logger:   logger,
	}
}

func (d *deferredStalenessAppendable) Appender(ctx context.Context) storage.Appender {
	app := d.delegate.Appender(ctx)
	return &deferredStalenessAppender{
		parent:   d,
		delegate: app,
	}
}

type deferredStalenessAppender struct {
	parent   *deferredStalenessAppendable
	delegate storage.Appender
}

func (d *deferredStalenessAppender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	if value.IsStaleNaN(v) {
		level.Warn(d.parent.logger).Log("msg", ">>> REJECTED STALE MARKER", "labels", l.String(), "timestamp", t)
		return 0, storage.ErrOutOfOrderSample
	}
	return d.delegate.Append(ref, l, t, v)
}

func (d *deferredStalenessAppender) Commit() error {
	return d.delegate.Commit()
}

func (d *deferredStalenessAppender) Rollback() error {
	return d.delegate.Rollback()
}

func (d *deferredStalenessAppender) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	return d.delegate.AppendExemplar(ref, l, e)
}

func (d *deferredStalenessAppender) AppendHistogram(ref storage.SeriesRef, l labels.Labels, t int64, h *histogram.Histogram, fh *histogram.FloatHistogram) (storage.SeriesRef, error) {
	return d.delegate.AppendHistogram(ref, l, t, h, fh)
}

func (d *deferredStalenessAppender) UpdateMetadata(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata) (storage.SeriesRef, error) {
	return d.delegate.UpdateMetadata(ref, l, m)
}
