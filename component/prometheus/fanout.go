package prometheus

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"

	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/scrape"

	"github.com/prometheus/prometheus/storage"
)

var _ storage.Appendable = (*Fanout)(nil)

// Fanout supports the default Flow style of appendables since it can go to multiple outputs. It also allows the intercepting of appends.
type Fanout struct {
	mut sync.RWMutex
	// children is where to fan out.
	children []storage.Appendable
	// ComponentID is what component this belongs to.
	componentID string
}

// NewFanout creates a fanout appendable.
func NewFanout(children []storage.Appendable, componentID string) *Fanout {
	return &Fanout{
		children:    children,
		componentID: componentID,
	}
}

// UpdateChildren allows changing of the children of the fanout.
func (f *Fanout) UpdateChildren(children []storage.Appendable) {
	f.mut.Lock()
	defer f.mut.Unlock()
	f.children = children
}

// Appender satisfies the Appendable interface.
func (f *Fanout) Appender(ctx context.Context) storage.Appender {
	f.mut.RLock()
	defer f.mut.RUnlock()

	ctx = scrape.ContextWithMetricMetadataStore(ctx, noopMetadataStore{})
	ctx = scrape.ContextWithTarget(ctx, &scrape.Target{})

	app := &appender{
		children:    make([]storage.Appender, 0),
		componentID: f.componentID,
	}
	for _, x := range f.children {
		if x == nil {
			continue
		}
		app.children = append(app.children, x.Appender(ctx))
	}
	return app
}

var _ storage.Appender = (*appender)(nil)

type appender struct {
	children    []storage.Appender
	componentID string
}

// Append satisfies the Appender interface.
func (a *appender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	if ref == 0 {
		ref = storage.SeriesRef(GlobalRefMapping.GetOrAddGlobalRefID(l))
	}
	var multiErr error
	for _, x := range a.children {
		_, err := x.Append(ref, l, t, v)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return ref, multiErr
}

// Commit satisfies the Appender interface.
func (a *appender) Commit() error {
	var multiErr error
	for _, x := range a.children {
		err := x.Commit()
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return multiErr
}

// Rollback satisifies the Appender interface.
func (a *appender) Rollback() error {
	var multiErr error
	for _, x := range a.children {
		err := x.Rollback()
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return multiErr
}

// AppendExemplar satisfies the Appender interface.
func (a *appender) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	return 0, fmt.Errorf("appendExemplar not supported yet")
}

// UpdateMetadata satisifies the Appender interface.
func (a *appender) UpdateMetadata(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata) (storage.SeriesRef, error) {
	return 0, fmt.Errorf("updateMetadata not supported yet")
}

type noopMetadataStore map[string]scrape.MetricMetadata

func (ms noopMetadataStore) GetMetadata(familyName string) (scrape.MetricMetadata, bool) {
	return scrape.MetricMetadata{}, false
}
func (ms noopMetadataStore) ListMetadata() []scrape.MetricMetadata { return nil }
func (ms noopMetadataStore) SizeMetadata() int                     { return 0 }
func (ms noopMetadataStore) LengthMetadata() int                   { return 0 }
