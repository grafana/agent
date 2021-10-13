package servicegraphprocessor

import (
	"context"
	"errors"
	"fmt"
	"time"

	util "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/traces/contextkeys"
	"github.com/hashicorp/go-multierror"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchpersignal"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/model/pdata"
	semconv "go.opentelemetry.io/collector/model/semconv/v1.6.1"
	"google.golang.org/grpc/codes"
)

var (
	errTooManyItems = errors.New("too many items in store")
)

// edgeRequest is a request between two nodes in the graph
type edgeRequest struct {
	serverService, clientService string
	serverLatency, clientLatency time.Duration

	// If either the client or the server spans have status code error,
	// the request will be considered as failed.
	failed bool
}

// complete returns true if the corresponding client and server
// pair spans have been processed for the given request
func (e *edgeRequest) complete() bool {
	return len(e.clientService) != 0 && len(e.serverService) != 0
}

var _ component.TracesProcessor = (*processor)(nil)

type processor struct {
	nextConsumer consumer.Traces
	reg          prometheus.Registerer

	// store is a local storage for request between graphs nodes
	store    *cache.Cache
	maxItems int

	serviceGraphRequestTotal           *prometheus.CounterVec
	serviceGraphRequestFailedTotal     *prometheus.CounterVec
	serviceGraphRequestServerHistogram *prometheus.HistogramVec
	serviceGraphRequestClientHistogram *prometheus.HistogramVec
	serviceGraphUnpairedSpansTotal     *prometheus.CounterVec
	serviceGraphDroppedSpansTotal      *prometheus.CounterVec

	httpSuccessCode map[int]struct{}
	grpcSuccessCode map[int]struct{}

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

	var (
		httpSuccessCode = make(map[int]struct{})
		grpcSuccessCode = make(map[int]struct{})
	)
	if cfg.SuccessCodes != nil {
		for _, sc := range cfg.SuccessCodes.http {
			httpSuccessCode[int(sc)] = struct{}{}
		}
		for _, sc := range cfg.SuccessCodes.grpc {
			grpcSuccessCode[int(sc)] = struct{}{}
		}
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
	p.serviceGraphRequestServerHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tempo_service_graph_request_server_seconds",
		Help:    "Time for a request between two nodes as seen from the server",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 12),
	}, []string{"client", "server"})
	p.serviceGraphRequestClientHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tempo_service_graph_request_client_seconds",
		Help:    "Time for a request between two nodes as seen from the client",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 12),
	}, []string{"client", "server"})
	p.serviceGraphUnpairedSpansTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tempo_service_graph_unpaired_spans_total",
		Help: "Total count of unpaired spans",
	}, []string{"client", "server"})
	p.serviceGraphDroppedSpansTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tempo_service_graph_dropped_spans_total",
		Help: "Total count of dropped spans",
	}, []string{"service"})

	cs := []prometheus.Collector{
		p.serviceGraphRequestTotal,
		p.serviceGraphRequestFailedTotal,
		p.serviceGraphRequestServerHistogram,
		p.serviceGraphRequestClientHistogram,
		p.serviceGraphUnpairedSpansTotal,
		p.serviceGraphDroppedSpansTotal,
	}

	for _, c := range cs {
		if err := p.reg.Register(c); err != nil {
			return err
		}
	}

	// Collect unpaired spans when evicting items from the store during
	// periodic cleanup
	p.store.OnEvicted(func(s string, i interface{}) {
		e := i.(edgeRequest)
		if !e.complete() {
			p.serviceGraphUnpairedSpansTotal.WithLabelValues(e.clientService, e.serverService).Inc()
		}
	})

	return nil
}

func (p *processor) Shutdown(context.Context) error {
	p.unregisterMetrics()
	p.store.Flush()
	return nil
}

func (p *processor) unregisterMetrics() {
	cs := []prometheus.Collector{
		p.serviceGraphRequestTotal,
		p.serviceGraphRequestFailedTotal,
		p.serviceGraphRequestServerHistogram,
		p.serviceGraphRequestClientHistogram,
		p.serviceGraphUnpairedSpansTotal,
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
				level.Info(p.logger).Log("msg", "skipped processing of spans", "maxItems", p.maxItems, "err", errTooManyItems)
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
		e := v.Object.(edgeRequest)
		if e.complete() {
			p.serviceGraphRequestTotal.WithLabelValues(e.clientService, e.serverService).Inc()
			if e.failed {
				p.serviceGraphRequestFailedTotal.WithLabelValues(e.clientService, e.serverService).Inc()
			}
			p.serviceGraphRequestServerHistogram.WithLabelValues(e.clientService, e.serverService).Observe(e.serverLatency.Seconds())
			p.serviceGraphRequestClientHistogram.WithLabelValues(e.clientService, e.serverService).Observe(e.clientLatency.Seconds())
			p.store.Delete(k)
		}
	}
}

func (p *processor) consume(trace pdata.Traces) error {
	rSpansSlice := trace.ResourceSpans()
	for i := 0; i < rSpansSlice.Len(); i++ {
		rSpan := rSpansSlice.At(i)

		svc, ok := rSpan.Resource().Attributes().Get(semconv.AttributeServiceName)
		if !ok {
			continue
		}

		ilsSlice := rSpan.InstrumentationLibrarySpans()
		for j := 0; j < ilsSlice.Len(); j++ {
			ils := ilsSlice.At(j)

			for k := 0; k < ils.Spans().Len(); k++ {

				if p.store.ItemCount() >= p.maxItems {
					remainingSpans := float64(ils.Spans().Len() - k)
					p.serviceGraphDroppedSpansTotal.WithLabelValues(svc.StringVal()).Add(remainingSpans)

					return errTooManyItems
				}

				span := ils.Spans().At(k)

				switch span.Kind() {
				case pdata.SpanKindClient:
					k := key(span.TraceID().HexString(), span.SpanID().HexString())

					var r edgeRequest
					if v, ok := p.store.Get(k); ok {
						r = v.(edgeRequest)
					}
					r.clientService = svc.StringVal()
					r.clientLatency = spanDuration(span)
					r.failed = p.spanFailed(span)
					p.store.SetDefault(k, r)

				case pdata.SpanKindServer:
					k := key(span.TraceID().HexString(), span.ParentSpanID().HexString())

					var r edgeRequest
					if v, ok := p.store.Get(k); ok {
						r = v.(edgeRequest)
					}

					r.serverService = svc.StringVal()
					r.serverLatency = spanDuration(span)
					r.failed = p.spanFailed(span)
					p.store.SetDefault(k, r)

				default:
				}
			}
		}
	}
	return nil
}

func (p *processor) spanFailed(span pdata.Span) bool {
	// Request considered failed if status is not 2XX or added as a successful status code
	if statusCode, ok := span.Attributes().Get("http.status_code"); ok {
		sc := int(statusCode.IntVal())
		if _, ok := p.httpSuccessCode[sc]; !ok || sc/100 != 2 {
			return true
		}
	}

	// Request considered failed if status is not OK or added as a successful status code
	if statusCode, ok := span.Attributes().Get("grpc.status_code"); ok {
		sc := int(statusCode.IntVal())
		if _, ok := p.grpcSuccessCode[sc]; !ok || sc != int(codes.OK) {
			return true
		}
	}

	return span.Status().Code() == pdata.StatusCodeError
}

func spanDuration(span pdata.Span) time.Duration {
	return span.EndTimestamp().AsTime().Sub(span.StartTimestamp().AsTime())
}

func key(k1, k2 string) string {
	return fmt.Sprintf("%s-%s", k1, k2)
}
