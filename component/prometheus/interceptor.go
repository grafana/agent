package prometheus

import (
	"context"
	"fmt"
	"sync"

	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/storage"
)

type intercept func(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, labels.Labels, int64, float64, error)

// Interceptor supports the concept of an appendable/appender that you can add a func to be called before
// the values are sent to the child appendable/append.
type Interceptor struct {
	mut sync.RWMutex
	// intercept allows one to intercept the series before it fans out to make any changes. If labels.Labels returns nil the series is not propagated.
	// Intercept shouuld be thread safe and can be called across appenders.
	intercept func(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, labels.Labels, int64, float64, error)

	// child is where to fan out.
	child storage.Appendable

	// ComponentID is what component this belongs to.
	componentID string
}

// NewInterceptor creates a interceptor appendable.
func NewInterceptor(inter intercept, child storage.Appendable, componentID string) *Interceptor {
	return &Interceptor{
		intercept:   inter,
		child:       child,
		componentID: componentID,
	}
}

// UpdateChild allows changing of the child of the interceptor.
func (f *Interceptor) UpdateChild(child storage.Appendable) {
	f.mut.Lock()
	defer f.mut.Unlock()

	f.child = child
}

// Appender satisfies the Appendable interface.
func (f *Interceptor) Appender(ctx context.Context) storage.Appender {
	f.mut.RLock()
	defer f.mut.RUnlock()

	app := &interceptappender{
		intercept:   f.intercept,
		componentID: f.componentID,
	}
	if f.child != nil {
		app.child = f.child.Appender(ctx)
	}
	return app
}

var _ storage.Appender = (*appender)(nil)

type interceptappender struct {
	child       storage.Appender
	componentID string
	intercept   func(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, labels.Labels, int64, float64, error)
}

// Append satisfies the Appender interface.
func (a *interceptappender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	if ref == 0 {
		ref = storage.SeriesRef(GlobalRefMapping.GetOrAddGlobalRefID(l))
	}
	newRef := ref
	newLabels := l
	newTimestamp := t
	newValue := v
	if a.intercept != nil {
		var err error
		newRef, newLabels, newTimestamp, newValue, err = a.intercept(ref, l, t, v)
		if err != nil {
			return 0, err
		}
	}
	if newLabels == nil {
		return ref, nil
	}
	if a.child == nil {
		return ref, nil
	}
	return a.child.Append(newRef, newLabels, newTimestamp, newValue)
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

	return 0, fmt.Errorf("appendExemplar not supported yet")
}

// UpdateMetadata satisifies the Appender interface.
func (a *interceptappender) UpdateMetadata(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata) (storage.SeriesRef, error) {
	return 0, fmt.Errorf("updateMetadata not supported yet")
}
