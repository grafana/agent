package labelstore

import (
	"context"

	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/storage"
)

type LabelStore interface {
	// ConvertToSeries This will convert a prometheus series to a labelstore series. Any series here must be later passed on to remove add/remove staleness markers
	ConvertToSeries(ts int64, val float64, lbls labels.Labels) *Series

	// GetOrAddLink returns the global id for the values, if none found one will be created based on the lbls.
	GetOrAddLink(componentID string, localRefID uint64, lbls labels.Labels) uint64

	// GetLocalRefID gets the mapping from global to local id specific to a component. Returns 0 if nothing found.
	GetLocalRefID(componentID string, globalRefID uint64) uint64

	// HandleStaleMarkers will remove or add staleness markers as needed.
	HandleStaleMarkers(series []*Series)
}

// Series should be treated as immutable and only created via ConvertToSeries.
type Series struct {
	GlobalID uint64
	Lbls     labels.Labels
	Hash     uint64
	Ts       int64
	Value    float64
}
type Appendable interface {
	Appender(ctx context.Context) Appender
}
type Appender interface {
	Append(s *Series) (storage.SeriesRef, error)
	AppendHistogram(s *Series, h *histogram.Histogram, fh *histogram.FloatHistogram) (storage.SeriesRef, error)
	UpdateMetadata(s *Series, m metadata.Metadata) (storage.SeriesRef, error)
	AppendExemplar(s *Series, e exemplar.Exemplar) (storage.SeriesRef, error)
	Commit() error
	Rollback() error
}
