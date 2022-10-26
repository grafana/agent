package remotewrite

import (
	"fmt"

	"github.com/grafana/agent/component/prometheus"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/storage"
	"golang.org/x/net/context"
)

var _ storage.Appendable = (*appendable)(nil)
var _ storage.Appender = (*appender)(nil)

type appendable struct {
	inner       storage.Appendable
	componentID string
}

// Appender satisfies the Appendable interface.
func (a *appendable) Appender(ctx context.Context) storage.Appender {
	app := &appender{
		child:       a.inner.Appender(ctx),
		componentID: a.componentID,
	}
	return app
}

type appender struct {
	child       storage.Appender
	componentID string
}

// Append satisfies the Appender interface.
func (a *appender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	// Conversion is needed because remote_writes assume they own all the IDs, so if you have two remote_writes they will
	// both assume they have only one scraper attached. In flow that is not true, so we need to translate from a global id
	// to a local (remote_write) id.
	localID := prometheus.GlobalRefMapping.GetLocalRefID(a.componentID, uint64(ref))
	newref, err := a.child.Append(storage.SeriesRef(localID), l, t, v)
	// If there was no local id we need to propagate it.
	if localID == 0 {
		prometheus.GlobalRefMapping.GetOrAddLink(a.componentID, uint64(newref), l)
	}
	return ref, err
}

// Commit satisfies the Appender interface.
func (a *appender) Commit() error {
	return a.child.Commit()
}

// Rollback satisfies the Appender interface.
func (a *appender) Rollback() error {
	return a.child.Rollback()
}

// AppendExemplar satisfies the Appender interface.
func (a *appender) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	return 0, fmt.Errorf("appendExemplar not supported yet")
}

// UpdateMetadata satisfies the Appender interface.
func (a *appender) UpdateMetadata(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata) (storage.SeriesRef, error) {
	return 0, fmt.Errorf("updateMetadata not supported yet")
}
