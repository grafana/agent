package servicegraphprocessor

import (
	"context"
	"errors"
	"fmt"
	"time"

	util "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/tempo/contextkeys"
	"github.com/hashicorp/go-multierror"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchpersignal"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/translator/conventions"
)

var (
	errTooManyItems = errors.New("too many items in store")
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
	serviceGraphUnpairedSpansTotal *prometheus.CounterVec
	serviceGraphUntaggedSpansTotal *prometheus.CounterVec

	logger log.Logger
}

func newProcessor(nextConsumer consumer.Traces, cfg *Config) *processor {
	logger := log.With(util.Logger, "component", "tempo service graphs")

	if cfg.Wait == 0 {
		cfg.Wait = DefaultWait
	}
	if cfg.MaxItems == 0 {
		cfg.MaxItems = DefaultMaxItems
	}

	// TODO(mapno): Add support for an external cache (e.g. memcached)
	p := &processor{
		nextConsumer: nextConsumer,
		// Cleanup period is hardcoded to twice the waiting time for simplicity
		// Most likely not ideal in every scenario
		store:    cache.New(cfg.Wait, cfg.Wait*2),
		maxItems: cfg.MaxItems,
		logger:   logger,
	}

	return p
}

func (p *processor) Start(ctx context.Context, _ component.Host) error {
	reg, ok := ctx.Value(contextkeys.PrometheusRegisterer).(prometheus.Registerer)
	if !ok || reg == nil {
		return fmt.Errorf("key does not contain a prometheus registerer")
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
	p.serviceGraphUnpairedSpansTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tempo_service_graph_unpaired_spans_total",
		Help: "Total count of requests between two nodes",
	}, []string{"client", "server"})
	p.serviceGraphUntaggedSpansTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tempo_service_graph_untagged_spans_total",
		Help: "Total count of spans processed that were not tagged with span.kind",
	}, []string{"span_kind"})

	cs := []prometheus.Collector{
		p.serviceGraphRequestTotal,
		p.serviceGraphRequestFailedTotal,
		p.serviceGraphRequestHistogram,
		p.serviceGraphUnpairedSpansTotal,
		p.serviceGraphUntaggedSpansTotal,
	}

	for _, c := range cs {
		if err := p.reg.Register(c); err != nil {
			return err
		}
	}

	// Collect unpaired spans when evicting items from the store during
	// periodic cleanup
	p.store.OnEvicted(func(s string, i interface{}) {
		e := i.(edge)
		if !e.complete() {
			p.serviceGraphUnpairedSpansTotal.WithLabelValues(e.clientService, e.serverService).Inc()
		}
	})

	return nil
}

func (p *processor) Shutdown(context.Context) error {
	p.unregisterMetrics()
	return nil
}

func (p *processor) unregisterMetrics() {
	cs := []prometheus.Collector{
		p.serviceGraphRequestTotal,
		p.serviceGraphRequestFailedTotal,
		p.serviceGraphRequestHistogram,
		p.serviceGraphUnpairedSpansTotal,
		p.serviceGraphUntaggedSpansTotal,
	}

	for _, c := range cs {
		p.reg.Unregister(c)
	}
}

func (p *processor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

func (p *processor) ConsumeTraces(ctx context.Context, td pdata.Traces) error {
	level.Debug(p.logger).Log("msg", "consuming traces")

	var errs error
	for _, trace := range batchpersignal.SplitTraces(td) {
		if err := p.consume(trace); err != nil {
			if errors.Is(err, errTooManyItems) {
				level.Warn(p.logger).Log("msg", "skipped processing of spans", "maxItems", p.maxItems, "err", errTooManyItems)
				break
			}
			errs = multierror.Append(errs, err)
		}
	}
	if errs != nil {
		level.Error(p.logger).Log("msg", "failed consuming traces", "err", errs)
	}

	p.collectMetrics()

	return p.nextConsumer.ConsumeTraces(ctx, td)
}

func (p *processor) collectMetrics() {
	for k, v := range p.store.Items() {
		e := v.Object.(edge)
		if e.complete() {
			p.serviceGraphRequestTotal.WithLabelValues(e.clientService, e.serverService).Inc()
			if e.failed {
				p.serviceGraphRequestFailedTotal.WithLabelValues(e.clientService, e.serverService).Inc()
			}
			p.serviceGraphRequestHistogram.WithLabelValues(e.clientService, e.serverService).Observe(e.serverLatency.Seconds())
			p.store.Delete(k)
		}
	}
}

func (p *processor) consume(trace pdata.Traces) error {
	rSpansSlice := trace.ResourceSpans()
	for i := 0; i < rSpansSlice.Len(); i++ {
		rSpan := rSpansSlice.At(i)

		svc, ok := rSpan.Resource().Attributes().Get(conventions.AttributeServiceName)
		if !ok {
			continue
		}

		ilsSlice := rSpan.InstrumentationLibrarySpans()
		for j := 0; j < ilsSlice.Len(); j++ {
			ils := ilsSlice.At(j)

			for k := 0; k < ils.Spans().Len(); k++ {
				span := ils.Spans().At(k)

				if p.store.ItemCount() >= p.maxItems {
					return errTooManyItems
				}

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
				default:
					p.serviceGraphUntaggedSpansTotal.WithLabelValues(span.Kind().String()).Inc()
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
