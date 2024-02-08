package prometheus

import (
	"context"

	"github.com/grafana/agent/service/labelstore"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/storage"
)

// Interceptor is a storage.Appendable which invokes callback functions upon
// getting data. Interceptor should not be modified once created. All callback
// fields are optional.
type Interceptor struct {
	onAppend          func(series *labelstore.Series, next labelstore.Appender) (storage.SeriesRef, error)
	onAppendExemplar  func(series *labelstore.Series, e exemplar.Exemplar, next labelstore.Appender) (storage.SeriesRef, error)
	onUpdateMetadata  func(series *labelstore.Series, m metadata.Metadata, next labelstore.Appender) (storage.SeriesRef, error)
	onAppendHistogram func(series *labelstore.Series, h *histogram.Histogram, fh *histogram.FloatHistogram, next labelstore.Appender) (storage.SeriesRef, error)

	// next is the next appendable to pass in the chain.
	next labelstore.Appendable

	ls labelstore.LabelStore
}

var _ labelstore.Appendable = (*Interceptor)(nil)

// NewInterceptor creates a new Interceptor storage.Appendable. Options can be
// provided to NewInterceptor to install custom hooks for different methods.
func NewInterceptor(next labelstore.Appendable, ls labelstore.LabelStore, opts ...InterceptorOption) *Interceptor {
	i := &Interceptor{
		next: next,
		ls:   ls,
	}
	for _, opt := range opts {
		opt(i)
	}
	return i
}

// InterceptorOption is an option argument passed to NewInterceptor.
type InterceptorOption func(*Interceptor)

// WithAppendHook returns an InterceptorOption which hooks into calls to
// Append.
func WithAppendHook(f func(series *labelstore.Series, next labelstore.Appender) (storage.SeriesRef, error)) InterceptorOption {
	return func(i *Interceptor) {
		i.onAppend = f
	}
}

// WithExemplarHook returns an InterceptorOption which hooks into calls to
// AppendExemplar.
func WithExemplarHook(f func(series *labelstore.Series, e exemplar.Exemplar, next labelstore.Appender) (storage.SeriesRef, error)) InterceptorOption {
	return func(i *Interceptor) {
		i.onAppendExemplar = f
	}
}

// WithMetadataHook returns an InterceptorOption which hooks into calls to
// UpdateMetadata.
func WithMetadataHook(f func(series *labelstore.Series, m metadata.Metadata, next labelstore.Appender) (storage.SeriesRef, error)) InterceptorOption {
	return func(i *Interceptor) {
		i.onUpdateMetadata = f
	}
}

// WithHistogramHook returns an InterceptorOption which hooks into calls to
// AppendHistogram.
func WithHistogramHook(f func(series *labelstore.Series, h *histogram.Histogram, fh *histogram.FloatHistogram, next labelstore.Appender) (storage.SeriesRef, error)) InterceptorOption {
	return func(i *Interceptor) {
		i.onAppendHistogram = f
	}
}

// Appender satisfies the Appendable interface.
func (f *Interceptor) Appender(ctx context.Context) labelstore.Appender {
	app := &interceptappender{
		interceptor: f,
		ls:          f.ls,
		series:      make([]*labelstore.Series, 0),
	}
	if f.next != nil {
		app.child = f.next.Appender(ctx)
	}
	return app
}

type interceptappender struct {
	interceptor *Interceptor
	child       labelstore.Appender
	ls          labelstore.LabelStore
	series      []*labelstore.Series
}

var _ labelstore.Appender = (*interceptappender)(nil)

// Append satisfies the Appender interface.
func (a *interceptappender) Append(series *labelstore.Series) (storage.SeriesRef, error) {
	a.series = append(a.series, series)
	if a.interceptor.onAppend != nil {
		return a.interceptor.onAppend(series, a.child)
	}
	if a.child == nil {
		return 0, nil
	}
	return a.child.Append(series)
}

// Commit satisfies the Appender interface.
func (a *interceptappender) Commit() error {
	a.ls.HandleStaleMarkers(a.series)
	if a.child == nil {
		return nil
	}
	return a.child.Commit()
}

// Rollback satisfies the Appender interface.
func (a *interceptappender) Rollback() error {
	a.ls.HandleStaleMarkers(a.series)
	if a.child == nil {
		return nil
	}
	return a.child.Rollback()
}

// AppendExemplar satisfies the Appender interface.
func (a *interceptappender) AppendExemplar(s *labelstore.Series, e exemplar.Exemplar) (storage.SeriesRef, error) {
	a.series = append(a.series, s)
	if a.interceptor.onAppendExemplar != nil {
		return a.interceptor.onAppendExemplar(s, e, a.child)
	}
	if a.child == nil {
		return 0, nil
	}
	return a.child.AppendExemplar(s, e)
}

// UpdateMetadata satisfies the Appender interface.
func (a *interceptappender) UpdateMetadata(s *labelstore.Series, m metadata.Metadata) (storage.SeriesRef, error) {
	a.series = append(a.series, s)
	if a.interceptor.onUpdateMetadata != nil {
		return a.interceptor.onUpdateMetadata(s, m, a.child)
	}
	if a.child == nil {
		return 0, nil
	}
	return a.child.UpdateMetadata(s, m)
}

func (a *interceptappender) AppendHistogram(s *labelstore.Series, h *histogram.Histogram, fh *histogram.FloatHistogram) (storage.SeriesRef, error) {
	a.series = append(a.series, s)
	if a.interceptor.onAppendHistogram != nil {
		return a.interceptor.onAppendHistogram(s, h, fh, a.child)
	}
	if a.child == nil {
		return 0, nil
	}
	return a.child.AppendHistogram(s, h, fh)
}
