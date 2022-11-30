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

// Intercept func allows interceptor owners to inject custom behavior.
type Intercept func(ref storage.SeriesRef, l labels.Labels, t int64, v float64, next storage.Appender) (storage.SeriesRef, error)

// Interceptor supports the concept of an appendable/appender that you can add a func to be called before
// the values are sent to the child appendable/append.
type Interceptor struct {
	mut sync.RWMutex
	// intercept allows one to intercept the series before it fans out to make any changes. If labels.Labels returns nil the series is not propagated.
	// Intercept shouuld be thread safe and can be called across appenders.
	intercept Intercept
	// next is where to send the next command.
	next storage.Appendable

	// ComponentID is what component this belongs to.
	componentID string
}

// NewInterceptor creates a interceptor appendable.
func NewInterceptor(inter Intercept, next storage.Appendable, componentID string) (*Interceptor, error) {
	if inter == nil {
		return nil, fmt.Errorf("intercept cannot be null for component %s", componentID)
	}
	return &Interceptor{
		intercept:   inter,
		next:        next,
		componentID: componentID,
	}, nil
}

// UpdateChild allows changing of the child of the interceptor.
func (f *Interceptor) UpdateChild(child storage.Appendable) {
	f.mut.Lock()
	defer f.mut.Unlock()

	f.next = child
}

// Appender satisfies the Appendable interface.
func (f *Interceptor) Appender(ctx context.Context) storage.Appender {
	f.mut.RLock()
	defer f.mut.RUnlock()

	app := &interceptappender{
		intercept:   f.intercept,
		componentID: f.componentID,
	}
	if f.next != nil {
		app.child = f.next.Appender(ctx)
	}
	return app
}

var _ storage.Appender = (*appender)(nil)

type interceptappender struct {
	child       storage.Appender
	componentID string
	intercept   Intercept
}

// Append satisfies the Appender interface.
func (a *interceptappender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	if ref == 0 {
		ref = storage.SeriesRef(GlobalRefMapping.GetOrAddGlobalRefID(l))
	}
	return a.intercept(ref, l, t, v, a.child)
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
func (a *interceptappender) UpdateMetadata(
	ref storage.SeriesRef,
	l labels.Labels,
	m metadata.Metadata,
) (storage.SeriesRef, error) {

	return 0, fmt.Errorf("updateMetadata not supported yet")
}
