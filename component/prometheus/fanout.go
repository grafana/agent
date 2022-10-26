package prometheus

import (
	"context"
	"sync"

	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"

	"github.com/prometheus/prometheus/storage"
)

var _ storage.Appendable = (*Fanout)(nil)

type intercept func(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, labels.Labels, int64, float64, error)

// Fanout supports the default Flow style of appendables since it can go to multiple outputs. It also allows the intercepting of appends.
type Fanout struct {
	mut sync.RWMutex
	// intercept allows one to intercept the series before it fans out to make any changes. If labels.Labels returns nil the series is not propagated.
	// Intercept shouuld be thread safe and can be called across appenders.
	intercept func(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, labels.Labels, int64, float64, error)

	// children is where to fan out.
	children []storage.Appendable

	// ComponentID is what component this belongs to.
	componentID string
}

// NewFanout creates a fanout appendable.
func NewFanout(inter intercept, children []storage.Appendable, componentID string) *Fanout {
	return &Fanout{
		intercept:   inter,
		children:    children,
		componentID: componentID,
	}
}

func (f *Fanout) UpdateChildren(children []storage.Appendable) {
	f.mut.Lock()
	defer f.mut.Unlock()
	f.children = children
}

// Appender satisfies the Appendable interface.
func (f *Fanout) Appender(ctx context.Context) storage.Appender {
	f.mut.RLock()
	defer f.mut.RUnlock()

	app := &appender{
		children:    make([]storage.Appender, len(f.children)),
		intercept:   f.intercept,
		componentID: f.componentID,
	}
	for i, x := range f.children {
		if x == nil {
			continue
		}
		app.children[i] = x.Appender(ctx)
	}
	return app
}

var _ storage.Appender = (*appender)(nil)

type appender struct {
	children    []storage.Appender
	componentID string
	intercept   func(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, labels.Labels, int64, float64, error)
}

// Append satisfies the Appender interface.
func (a *appender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	if ref == 0 {
		ref = storage.SeriesRef(GlobalRefMapping.GetOrAddGlobalRefID(l))
	}
	nRef := ref
	nL := l
	nT := t
	nV := v
	if a.intercept != nil {
		var err error
		nRef, nL, nT, nV, err = a.intercept(ref, l, t, v)
		if err != nil {
			return 0, err
		}
	}
	for _, x := range a.children {
		if x == nil || nL == nil {
			continue
		}
		_, _ = x.Append(nRef, nL, nT, nV)
	}
	return ref, nil
}

// Commit satisfies the Appender interface.
func (a *appender) Commit() error {
	for _, x := range a.children {
		if x == nil {
			continue
		}
		_ = x.Commit()
	}
	return nil
}

// Rollback satisifies the Appender interface.
func (a *appender) Rollback() error {
	for _, x := range a.children {
		_, _ = x, a.Rollback()
	}
	return nil
}

// AppendExemplar satisfies the Appender interface.
func (a *appender) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	// TODO implement me
	panic("implement me")
}

// UpdateMetadata satisifies the Appender interface.
func (a *appender) UpdateMetadata(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata) (storage.SeriesRef, error) {
	// TODO implement me
	panic("implement me")
}
