package scrape

import (
	"context"

	"github.com/grafana/agent/component/prometheus"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"

	"github.com/prometheus/prometheus/storage"
)

var _ storage.Appendable = (*fanout)(nil)

type fanout struct {
	children []storage.Appendable
}

func (f *fanout) Appender(ctx context.Context) storage.Appender {
	app := &appender{children: make([]storage.Appender, 0)}
	for _, x := range f.children {
		app.children = append(app.children, x.Appender(ctx))
	}
	return app

}

var _ storage.Appender = (*appender)(nil)

type appender struct {
	children []storage.Appender
}

func (a *appender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	if ref == 0 {
		ref = storage.SeriesRef(prometheus.GlobalRefMapping.GetOrAddGlobalRefID(l))
	}
	for _, x := range a.children {
		_, _ = x.Append(ref, l, t, v)
	}
	return ref, nil
}

func (a *appender) Commit() error {
	for _, x := range a.children {
		_ = x.Commit()
	}
	return nil
}

func (a *appender) Rollback() error {
	for _, x := range a.children {
		_, _ = x, a.Rollback()
	}
	return nil
}

func (a *appender) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	//TODO implement me
	panic("implement me")
}

func (a *appender) UpdateMetadata(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata) (storage.SeriesRef, error) {
	//TODO implement me
	panic("implement me")
}
