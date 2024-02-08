package labelstore

import (
	"context"

	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/storage"
)

var _ storage.Appendable = (*Shim)(nil)

// Shim provides a translation layer between the native prometheus and flow series.
type Shim struct {
	ls   LabelStore
	next Appendable
}

func NewShim(ls LabelStore, next Appendable) *Shim {
	return &Shim{
		ls:   ls,
		next: next,
	}
}

// Appender returns a new appender for the storage. The implementation
// can choose whether or not to use the context, for deadlines or to check
// for errors.
func (f Shim) Appender(ctx context.Context) storage.Appender {
	return &shimappender{
		ls:     f.ls,
		next:   f.next.Appender(ctx),
		series: make([]*Series, 0),
	}
}

var _ storage.Appender = (*shimappender)(nil)

type shimappender struct {
	ls     LabelStore
	series []*Series
	next   Appender
}

// Append adds a sample pair for the given series.
// An optional series reference can be provided to accelerate calls.
// A series reference number is returned which can be used to add further
// samples to the given series in the same or later transactions.
// Returned reference numbers are ephemeral and may be rejected in calls
// to Append() at any point. Adding the sample via Append() returns a new
// reference number.
// If the reference is 0 it must not be used for caching.
func (sa *shimappender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	series := sa.ls.ConvertToSeries(t, v, l)
	sa.series = append(sa.series, series)
	return sa.next.Append(series)
}

// Commit submits the collected samples and purges the batch. If Commit
// returns a non-nil error, it also rolls back all modifications made in
// the appender so far, as Rollback would do. In any case, an Appender
// must not be used anymore after Commit has been called.
func (sa *shimappender) Commit() error {
	sa.ls.HandleStaleMarkers(sa.series)
	return sa.next.Commit()
}

// Rollback rolls back all modifications made in the appender so far.
// Appender has to be discarded after rollback.
func (sa *shimappender) Rollback() error {
	// Even in a rollback we should handle this since they have still been added to the cache.
	sa.ls.HandleStaleMarkers(sa.series)
	return sa.next.Rollback()
}

// AppendExemplar adds an exemplar for the given series labels.
// An optional reference number can be provided to accelerate calls.
// A reference number is returned which can be used to add further
// exemplars in the same or later transactions.
// Returned reference numbers are ephemeral and may be rejected in calls
// to Append() at any point. Adding the sample via Append() returns a new
// reference number.
// If the reference is 0 it must not be used for caching.
// Note that in our current implementation of Prometheus' exemplar storage
// calls to Append should generate the reference numbers, AppendExemplar
// generating a new reference number should be considered possible erroneous behaviour and be logged.
func (sa *shimappender) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	series := sa.ls.ConvertToSeries(0, 0, l)
	sa.series = append(sa.series, series)
	return sa.next.AppendExemplar(series, e)
}

// AppendHistogram adds a histogram for the given series labels. An
// optional reference number can be provided to accelerate calls. A
// reference number is returned which can be used to add further
// histograms in the same or later transactions. Returned reference
// numbers are ephemeral and may be rejected in calls to Append() at any
// point. Adding the sample via Append() returns a new reference number.
// If the reference is 0, it must not be used for caching.
//
// For efficiency reasons, the histogram is passed as a
// pointer. AppendHistogram won't mutate the histogram, but in turn
// depends on the caller to not mutate it either.
func (sa *shimappender) AppendHistogram(ref storage.SeriesRef, l labels.Labels, t int64, h *histogram.Histogram, fh *histogram.FloatHistogram) (storage.SeriesRef, error) {
	series := sa.ls.ConvertToSeries(0, 0, l)
	sa.series = append(sa.series, series)
	return sa.next.AppendHistogram(series, h, fh)
}

// UpdateMetadata updates a metadata entry for the given series and labels.
// A series reference number is returned which can be used to modify the
// metadata of the given series in the same or later transactions.
// Returned reference numbers are ephemeral and may be rejected in calls
// to UpdateMetadata() at any point. If the series does not exist,
// UpdateMetadata returns an error.
// If the reference is 0 it must not be used for caching.
func (sa *shimappender) UpdateMetadata(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata) (storage.SeriesRef, error) {
	series := sa.ls.ConvertToSeries(0, 0, l)
	sa.series = append(sa.series, series)
	return sa.next.UpdateMetadata(series, m)
}
