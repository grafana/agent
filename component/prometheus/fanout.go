package prometheus

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/hashicorp/go-multierror"

	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/histogram"
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
	componentID    string
	writeLatency   prometheus.Histogram
	samplesCounter prometheus.Counter
}

// NewFanout creates a fanout appendable.
func NewFanout(children []storage.Appendable, componentID string, register prometheus.Registerer) *Fanout {
	wl := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "agent_prometheus_fanout_latency",
		Help: "Write latency for sending to direct and indirect components",
	})
	_ = register.Register(wl)

	s := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "agent_prometheus_forwarded_samples_total",
		Help: "Total number of samples sent to downstream components.",
	})
	_ = register.Register(s)

	return &Fanout{
		children:       children,
		componentID:    componentID,
		writeLatency:   wl,
		samplesCounter: s,
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

	// TODO(@tpaschalis): The `otelcol.receiver.prometheus` component reuses
	// code from the prometheusreceiver which expects the Appender context to
	// be contain both a scrape target and a metadata store, and fails the
	// conversion if they are missing. We should find a way around this as both
	// Targets and Metadata will be handled in a different way in Flow.
	ctx = scrape.ContextWithTarget(ctx, &scrape.Target{})
	ctx = scrape.ContextWithMetricMetadataStore(ctx, NoopMetadataStore{})

	app := &appender{
		children:       make([]storage.Appender, 0),
		componentID:    f.componentID,
		writeLatency:   f.writeLatency,
		samplesCounter: f.samplesCounter,
	}

	for _, x := range f.children {
		if x == nil {
			continue
		}
		app.children = append(app.children, x.Appender(ctx))
	}
	return app
}

type appender struct {
	children       []storage.Appender
	componentID    string
	writeLatency   prometheus.Histogram
	samplesCounter prometheus.Counter
	start          time.Time
}

var _ storage.Appender = (*appender)(nil)

// Append satisfies the Appender interface.
func (a *appender) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	if a.start.IsZero() {
		a.start = time.Now()
	}
	if ref == 0 {
		ref = storage.SeriesRef(GlobalRefMapping.GetOrAddGlobalRefID(l))
	}
	var multiErr error
	updated := false
	for _, x := range a.children {
		_, err := x.Append(ref, l, t, v)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		} else {
			updated = true
		}
	}
	if updated {
		a.samplesCounter.Inc()
	}
	return ref, multiErr
}

// Commit satisfies the Appender interface.
func (a *appender) Commit() error {
	defer a.recordLatency()
	var multiErr error
	for _, x := range a.children {
		err := x.Commit()
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return multiErr
}

// Rollback satisfies the Appender interface.
func (a *appender) Rollback() error {
	defer a.recordLatency()
	var multiErr error
	for _, x := range a.children {
		err := x.Rollback()
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return multiErr
}

func (a *appender) recordLatency() {
	if a.start.IsZero() {
		return
	}
	duration := time.Since(a.start)
	a.writeLatency.Observe(duration.Seconds())
}

// AppendExemplar satisfies the Appender interface.
func (a *appender) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	if a.start.IsZero() {
		a.start = time.Now()
	}
	if ref == 0 {
		ref = storage.SeriesRef(GlobalRefMapping.GetOrAddGlobalRefID(l))
	}
	var multiErr error
	for _, x := range a.children {
		_, err := x.AppendExemplar(ref, l, e)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return ref, multiErr
}

// UpdateMetadata satisfies the Appender interface.
func (a *appender) UpdateMetadata(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata) (storage.SeriesRef, error) {
	if a.start.IsZero() {
		a.start = time.Now()
	}
	if ref == 0 {
		ref = storage.SeriesRef(GlobalRefMapping.GetOrAddGlobalRefID(l))
	}
	var multiErr error
	for _, x := range a.children {
		_, err := x.UpdateMetadata(ref, l, m)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return ref, multiErr
}

func (a *appender) AppendHistogram(ref storage.SeriesRef, l labels.Labels, t int64, h *histogram.Histogram, fh *histogram.FloatHistogram) (storage.SeriesRef, error) {
	if a.start.IsZero() {
		a.start = time.Now()
	}
	if ref == 0 {
		ref = storage.SeriesRef(GlobalRefMapping.GetOrAddGlobalRefID(l))
	}
	var multiErr error
	for _, x := range a.children {
		_, err := x.AppendHistogram(ref, l, t, h, fh)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return ref, multiErr
}

// NoopMetadataStore implements the MetricMetadataStore interface.
type NoopMetadataStore map[string]scrape.MetricMetadata

// GetMetadata implements the MetricMetadataStore interface.
func (ms NoopMetadataStore) GetMetadata(familyName string) (scrape.MetricMetadata, bool) {
	return scrape.MetricMetadata{}, false
}

// ListMetadata implements the MetricMetadataStore interface.
func (ms NoopMetadataStore) ListMetadata() []scrape.MetricMetadata { return nil }

// SizeMetadata implements the MetricMetadataStore interface.
func (ms NoopMetadataStore) SizeMetadata() int { return 0 }

// LengthMetadata implements the MetricMetadataStore interface.
func (ms NoopMetadataStore) LengthMetadata() int { return 0 }
