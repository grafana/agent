package remotewrite

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/grafana/agent/component/prometheus"
	"github.com/prometheus/prometheus/model/textparse"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/scrape"
	"strings"

	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/storage"
)

// Interceptor is a storage.Appendable which invokes callback functions upon
// getting data. Interceptor should not be modified once created. All callback
// fields are optional.
type Interceptor struct {
	onAppend          func(ref storage.SeriesRef, l labels.Labels, t int64, v float64, next storage.Appender) (storage.SeriesRef, error)
	onAppendExemplar  func(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar, next storage.Appender) (storage.SeriesRef, error)
	onUpdateMetadata  func(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata, next storage.Appender) (storage.SeriesRef, error)
	onAppendHistogram func(ref storage.SeriesRef, l labels.Labels, t int64, h *histogram.Histogram, fh *histogram.FloatHistogram, next storage.Appender) (storage.SeriesRef, error)
	// next is the next appendable to pass in the chain.
	next storage.Appendable
	send func(ctx context.Context, metadata []prompb.MetricMetadata, pBuf *proto.Buffer) error
}

var _ storage.Appendable = (*Interceptor)(nil)

// NewInterceptor creates a new Interceptor storage.Appendable. Options can be
// provided to NewInterceptor to install custom hooks for different methods.
func NewInterceptor(next storage.Appendable, send func(ctx context.Context, metadata []prompb.MetricMetadata, pBuf *proto.Buffer) error, opts ...InterceptorOption) *Interceptor {
	i := &Interceptor{next: next, send: send}
	for _, opt := range opts {
		opt(i)
	}
	return i
}

// InterceptorOption is an option argument passed to NewInterceptor.
type InterceptorOption func(*Interceptor)

// WithAppendHook returns an InterceptorOption which hooks into calls to
// Append.
func WithAppendHook(f func(ref storage.SeriesRef, l labels.Labels, t int64, v float64, next storage.Appender) (storage.SeriesRef, error)) InterceptorOption {
	return func(i *Interceptor) {
		i.onAppend = f
	}
}

// WithExemplarHook returns an InterceptorOption which hooks into calls to
// AppendExemplar.
func WithExemplarHook(f func(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar, next storage.Appender) (storage.SeriesRef, error)) InterceptorOption {
	return func(i *Interceptor) {
		i.onAppendExemplar = f
	}
}

// WithMetadataHook returns an InterceptorOption which hooks into calls to
// UpdateMetadata.
func WithMetadataHook(f func(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata, next storage.Appender) (storage.SeriesRef, error)) InterceptorOption {
	return func(i *Interceptor) {
		i.onUpdateMetadata = f
	}
}

// WithAppendHistogram returns an InterceptorOption which hooks into calls to
// AppendHistogram.
func WithAppendHistogram(f func(ref storage.SeriesRef, l labels.Labels, t int64, h *histogram.Histogram, fh *histogram.FloatHistogram, next storage.Appender) (storage.SeriesRef, error)) InterceptorOption {
	return func(i *Interceptor) {
		i.onAppendHistogram = f
	}
}

// Appender satisfies the Appendable interface.
func (f *Interceptor) Appender(ctx context.Context) storage.Appender {
	app := &interceptappender{
		interceptor: f,
		ctx:         ctx,
		lbls:        make([]labels.Labels, 0),
		send:        f.send,
	}
	if f.next != nil {
		app.child = f.next.Appender(ctx)
	}
	return app
}

type interceptappender struct {
	interceptor *Interceptor
	lbls        []labels.Labels
	child       storage.Appender
	ctx         context.Context
	send        func(ctx context.Context, metadata []prompb.MetricMetadata, pBuf *proto.Buffer) error
}

var _ storage.Appender = (*interceptappender)(nil)

// Append satisfies the Appender interface.
func (a *interceptappender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	if ref == 0 {
		ref = storage.SeriesRef(prometheus.GlobalRefMapping.GetOrAddGlobalRefID(l))
	}
	a.lbls = append(a.lbls, l)

	if a.interceptor.onAppend != nil {
		return a.interceptor.onAppend(ref, l, t, v, a.child)
	}
	return a.child.Append(ref, l, t, v)
}

// Commit satisfies the Appender interface.
func (a *interceptappender) Commit() error {
	if a.child == nil {
		return nil
	}
	if meta, found := scrape.MetricMetadataStoreFromContext(a.ctx); found {
		allmeta := make([]prompb.MetricMetadata, 0)
		for _, l := range a.lbls {
			name := l.Get("__name__")
			if arr, metaFound := meta.GetMetadata(name); metaFound {
				allmeta = append(allmeta, prompb.MetricMetadata{
					Type:             metricTypeToMetricTypeProto(arr.Type),
					Unit:             arr.Unit,
					Help:             arr.Help,
					MetricFamilyName: name,
				})
			}
		}
		pBuf := proto.NewBuffer(nil)
		a.send(a.ctx, allmeta, pBuf)
	}
	return a.child.Commit()
}

// Rollback satisfies the Appender interface.
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
		ref = storage.SeriesRef(prometheus.GlobalRefMapping.GetOrAddGlobalRefID(l))
	}

	if a.interceptor.onAppendExemplar != nil {
		return a.interceptor.onAppendExemplar(ref, l, e, a.child)
	}
	return a.child.AppendExemplar(ref, l, e)
}

// UpdateMetadata satisfies the Appender interface.
func (a *interceptappender) UpdateMetadata(
	ref storage.SeriesRef,
	l labels.Labels,
	m metadata.Metadata,
) (storage.SeriesRef, error) {

	if ref == 0 {
		ref = storage.SeriesRef(prometheus.GlobalRefMapping.GetOrAddGlobalRefID(l))
	}

	if a.interceptor.onUpdateMetadata != nil {
		return a.interceptor.onUpdateMetadata(ref, l, m, a.child)
	}
	return a.child.UpdateMetadata(ref, l, m)
}

func (a *interceptappender) AppendHistogram(
	ref storage.SeriesRef,
	l labels.Labels,
	t int64,
	h *histogram.Histogram,
	fh *histogram.FloatHistogram,
) (storage.SeriesRef, error) {

	if ref == 0 {
		ref = storage.SeriesRef(prometheus.GlobalRefMapping.GetOrAddGlobalRefID(l))
	}

	if a.interceptor.onAppendHistogram != nil {
		return a.interceptor.onAppendHistogram(ref, l, t, h, fh, a.child)
	}
	return a.child.AppendHistogram(ref, l, t, h, fh)
}

// metricTypeToMetricTypeProto transforms a Prometheus metricType into prompb metricType. Since the former is a string we need to transform it to an enum.
func metricTypeToMetricTypeProto(t textparse.MetricType) prompb.MetricMetadata_MetricType {
	mt := strings.ToUpper(string(t))
	v, ok := prompb.MetricMetadata_MetricType_value[mt]
	if !ok {
		return prompb.MetricMetadata_UNKNOWN
	}

	return prompb.MetricMetadata_MetricType(v)
}
