package prometheus

import (
	"context"

	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/storage"
)

// Interceptor is a storage.Appendable which invokes callback functions upon
// getting data. Interceptor should not be modified once created. All callback
// fields are optional.
type Interceptor struct {
	OnAppend         func(ref storage.SeriesRef, l labels.Labels, t int64, v float64, next storage.Appender) (storage.SeriesRef, error)
	OnAppendExemplar func(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar, next storage.Appender) (storage.SeriesRef, error)
	OnUpdateMetadata func(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata, next storage.Appender) (storage.SeriesRef, error)

	// Next is the next appendable to pass in the chain.
	Next storage.Appendable
}

var _ storage.Appendable = (*Interceptor)(nil)

// Appender satisfies the Appendable interface.
func (f *Interceptor) Appender(ctx context.Context) storage.Appender {
	app := &interceptappender{
		interceptor: f,
	}
	if f.Next != nil {
		app.child = f.Next.Appender(ctx)
	}
	return app
}

type interceptappender struct {
	interceptor *Interceptor
	child       storage.Appender
}

var _ storage.Appender = (*interceptappender)(nil)

// Append satisfies the Appender interface.
func (a *interceptappender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	if ref == 0 {
		ref = storage.SeriesRef(GlobalRefMapping.GetOrAddGlobalRefID(l))
	}

	if a.interceptor.OnAppend != nil {
		return a.interceptor.OnAppend(ref, l, t, v, a.child)
	}
	return a.child.Append(ref, l, t, v)
}

// Commit satisfies the Appender interface.
func (a *interceptappender) Commit() error {
	if a.child == nil {
		return nil
	}
	return a.child.Commit()
}

// Rollback satisifies the Appender interface.
func (a *interceptappender) Rollback() error {
	if a.child == nil {
		return nil
	}
	return a.child.Rollback()
}

// AppendExemplar satisfies the Appender interface.
func (a *interceptappender) AppendExemplar(
	ref storage.SeriesRef,
	l labels.Labels,
	e exemplar.Exemplar,
) (storage.SeriesRef, error) {

	if ref == 0 {
		ref = storage.SeriesRef(GlobalRefMapping.GetOrAddGlobalRefID(l))
	}

	if a.interceptor.OnAppendExemplar != nil {
		return a.interceptor.OnAppendExemplar(ref, l, e, a.child)
	}
	return a.child.AppendExemplar(ref, l, e)
}

// UpdateMetadata satisifies the Appender interface.
func (a *interceptappender) UpdateMetadata(
	ref storage.SeriesRef,
	l labels.Labels,
	m metadata.Metadata,
) (storage.SeriesRef, error) {

	if ref == 0 {
		ref = storage.SeriesRef(GlobalRefMapping.GetOrAddGlobalRefID(l))
	}

	if a.interceptor.OnUpdateMetadata != nil {
		return a.interceptor.OnUpdateMetadata(ref, l, m, a.child)
	}
	return a.child.UpdateMetadata(ref, l, m)
}
