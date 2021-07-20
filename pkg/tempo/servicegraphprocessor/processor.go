package servicegraphprocessor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/tempo/contextkeys"
	"github.com/hashicorp/go-multierror"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchpersignal"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/translator/conventions"
	"go.uber.org/atomic"
)

var (
	ErrNoServiceName = errors.New("failed to find service name")
	ErrTooManyEdges  = errors.New("too many edges in memory")
)

type edge struct {
	serverService, clientService string
	serverLatency, clientLatency time.Duration

	failed bool
}

func (e *edge) complete() bool {
	return len(e.clientService) != 0 && len(e.serverService) != 0
}

var _ component.TracesProcessor = (*processor)(nil)

type processor struct {
	nextConsumer consumer.Traces
	reg          prometheus.Registerer

	store    *cache.Cache
	maxItems int

	serviceGraphRequestTotal       *prometheus.CounterVec
	serviceGraphRequestFailedTotal *prometheus.CounterVec
	serviceGraphRequestHistogram   *prometheus.HistogramVec

	closed atomic.Bool

	logger log.Logger
}

func newProcessor(nextConsumer consumer.Traces, cfg *Config) (*processor, error) {
	if cfg.wait == 0 {
		cfg.wait = defaultWait
	}
	if cfg.maxEdges == 0 {
		cfg.maxEdges = defaultMaxEdges
	}

	p := &processor{
		nextConsumer: nextConsumer,
		store:        cache.New(cfg.wait, cfg.wait*2),
		maxItems:     defaultMaxEdges,
		closed:       atomic.Bool{},
	}

	return p, nil
}

func (p *processor) Start(ctx context.Context, _ component.Host) error {
	var reg prometheus.Registerer
	reg = prometheus.NewRegistry()
	if v, ok := ctx.Value(contextkeys.PrometheusRegisterer).(prometheus.Registerer); ok {
		reg = v
	}
	p.reg = reg

	return p.registerMetrics()
}

func (p *processor) registerMetrics() error {
	p.serviceGraphRequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tempo_service_graph_request_total",
		Help: "Total count of requests between two nodes",
	}, []string{"client", "server"})
	p.serviceGraphRequestFailedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tempo_service_graph_request_failed_total",
		Help: "Total count of failed requests between two nodes",
	}, []string{"client", "server"})
	p.serviceGraphRequestHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tempo_service_graph_request_seconds",
		Help:    "Time for a request between two nodes",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 12),
	}, []string{"client", "server"})

	cs := []prometheus.Collector{
		p.serviceGraphRequestTotal,
		p.serviceGraphRequestFailedTotal,
		p.serviceGraphRequestHistogram,
	}

	for _, c := range cs {
		if err := p.reg.Register(c); err != nil {
			return err
		}
	}

	return nil
}

func (p *processor) Shutdown(context.Context) error {
	p.closed.Store(true)
	p.unregisterMetrics()
	return nil
}

func (p *processor) unregisterMetrics() {
	cs := []prometheus.Collector{
		p.serviceGraphRequestTotal,
		p.serviceGraphRequestFailedTotal,
		p.serviceGraphRequestHistogram,
	}

	for _, c := range cs {
		p.reg.Unregister(c)
	}
}

func (p *processor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

func (p *processor) ConsumeTraces(ctx context.Context, td pdata.Traces) error {
	if p.closed.Load() {
		return nil
	}

	if p.store.ItemCount() >= p.maxItems {
		return ErrTooManyEdges
	}

	var errs error
	for _, trace := range batchpersignal.SplitTraces(td) {
		if err := p.consume(trace); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	if errs != nil {
		return errs
	}

	p.collectMetrics()

	return p.nextConsumer.ConsumeTraces(ctx, td)
}

func (p *processor) collectMetrics() {
	for _, v := range p.store.Items() {
		e := v.Object.(edge)
		if e.complete() {
			p.serviceGraphRequestTotal.WithLabelValues(e.clientService, e.serverService).Inc()
			if e.failed {
				p.serviceGraphRequestFailedTotal.WithLabelValues(e.clientService, e.serverService).Inc()
			}
			p.serviceGraphRequestHistogram.WithLabelValues(e.clientService, e.serverService).Observe(e.serverLatency.Seconds())
		}
	}
}

func (p *processor) consume(trace pdata.Traces) error {
	rSpansSlice := trace.ResourceSpans()
	for i := 0; i < rSpansSlice.Len(); i++ {
		rSpan := rSpansSlice.At(i)

		svc, ok := rSpan.Resource().Attributes().Get(conventions.AttributeServiceName)
		if !ok {
			return ErrNoServiceName
		}

		ilsSlice := rSpan.InstrumentationLibrarySpans()
		for j := 0; j < ilsSlice.Len(); j++ {
			ils := ilsSlice.At(j)

			for k := 0; k < ils.Spans().Len(); k++ {
				span := ils.Spans().At(k)

				switch span.Kind() {
				case pdata.SpanKindClient:
					k := key(span.TraceID().HexString(), span.SpanID().HexString())

					var e edge
					if v, ok := p.store.Get(k); ok {
						e = v.(edge)
					}
					e.clientService = svc.StringVal()
					e.clientLatency = spanDuration(span)
					p.store.SetDefault(k, e)

				case pdata.SpanKindServer:
					k := key(span.TraceID().HexString(), span.ParentSpanID().HexString())

					var e edge
					if v, ok := p.store.Get(k); ok {
						e = v.(edge)
					}

					e.serverService = svc.StringVal()
					e.serverLatency = spanDuration(span)
					p.store.SetDefault(k, e)
				}
			}
		}
	}
	return nil
}

func spanDuration(span pdata.Span) time.Duration {
	return span.EndTimestamp().AsTime().Sub(span.StartTimestamp().AsTime())
}

func key(k1, k2 string) string {
	return fmt.Sprintf("%s-%s", k1, k2)
}
