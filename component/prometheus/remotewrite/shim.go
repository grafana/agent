package remotewrite

import (
	"context"

	"github.com/grafana/agent/service/labelstore"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/storage"
)

var _ labelstore.Appendable = (*shim)(nil)

type shim struct {
	id   string
	ls   labelstore.LabelStore
	next storage.Appendable
}

func (s *shim) Appender(ctx context.Context) labelstore.Appender {
	return &appender{
		id:   s.id,
		ls:   s.ls,
		next: s.next.Appender(ctx),
	}
}

var _ labelstore.Appender = (*appender)(nil)

type appender struct {
	ls   labelstore.LabelStore
	id   string
	next storage.Appender
}

func (a *appender) Append(s *labelstore.Series) (storage.SeriesRef, error) {
	localID := a.ls.GetLocalRefID(a.id, s.GlobalID)
	newRef, nextErr := a.next.Append(storage.SeriesRef(localID), s.Lbls, s.Ts, s.Value)
	if localID == 0 {
		a.ls.GetOrAddLink(a.id, uint64(newRef), s.Lbls)
	}
	return storage.SeriesRef(s.GlobalID), nextErr
}

func (a *appender) AppendHistogram(s *labelstore.Series, h *histogram.Histogram, fh *histogram.FloatHistogram) (_ storage.SeriesRef, _ error) {
	localID := a.ls.GetLocalRefID(a.id, s.GlobalID)
	newRef, nextErr := a.next.AppendHistogram(storage.SeriesRef(localID), s.Lbls, s.Ts, h, fh)
	if localID == 0 {
		a.ls.GetOrAddLink(a.id, uint64(newRef), s.Lbls)
	}
	return storage.SeriesRef(s.GlobalID), nextErr
}

func (a *appender) UpdateMetadata(s *labelstore.Series, m metadata.Metadata) (_ storage.SeriesRef, _ error) {
	localID := a.ls.GetLocalRefID(a.id, s.GlobalID)
	newRef, nextErr := a.next.UpdateMetadata(storage.SeriesRef(localID), s.Lbls, m)
	if localID == 0 {
		a.ls.GetOrAddLink(a.id, uint64(newRef), s.Lbls)
	}
	return storage.SeriesRef(s.GlobalID), nextErr
}

func (a *appender) AppendExemplar(s *labelstore.Series, e exemplar.Exemplar) (_ storage.SeriesRef, _ error) {
	localID := a.ls.GetLocalRefID(a.id, s.GlobalID)
	newRef, nextErr := a.next.AppendExemplar(storage.SeriesRef(localID), s.Lbls, e)
	if localID == 0 {
		a.ls.GetOrAddLink(a.id, uint64(newRef), s.Lbls)
	}
	return storage.SeriesRef(s.GlobalID), nextErr
}

func (a *appender) Commit() error {
	return a.next.Commit()
}

func (a *appender) Rollback() (_ error) {
	return a.next.Rollback()
}
