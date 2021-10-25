package servicegraphprocessor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"runtime"

	util "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/traces/contextkeys"
	"github.com/hashicorp/go-multierror"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchpersignal"
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

// edge is a edge between two nodes in the graph
type edge struct {
	serverService, clientService string
	serverLatency, clientLatency time.Duration

	// If either the client or the server spans have status code error,
	// the edge will be considered as failed.
	failed bool

	// expiration is the time at which the edge expires, expressed as Unix time
	expiration int64
}

func newEdge(w time.Duration, expireFn func()) *edge {
	go expireFn()
	return &edge{
		expiration: time.Now().Add(w).Unix(),
	}
}

// completed returns true if the corresponding client and server
// pair spans have been processed for the given edge
func (e *edge) isCompleted() bool {
	return len(e.clientService) != 0 && len(e.serverService) != 0
}

func (e *edge) isExpired() bool {
	return time.Now().Unix() >= e.expiration
}

var _ component.TracesProcessor = (*processor)(nil)

type processor struct {
	nextConsumer consumer.Traces
	reg          prometheus.Registerer

	// store is a local storage for edge between graphs nodes
	store    map[string]*edge
	storeMtx sync.RWMutex
	maxItems int
	wait     time.Duration

	firesCh chan string

	closedCh chan struct{}

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

	p := &processor{
		nextConsumer: nextConsumer,
		logger:       logger,

		store:    map[string]*edge{},
		storeMtx: sync.RWMutex{},
		maxItems: cfg.MaxItems,
		wait:     cfg.Wait,

		firesCh: make(chan string, runtime.NumCPU()),

		closedCh: make(chan struct{}, 1),
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
		Namespace: "traces",
		Name:      "service_graph_request_total",
		Help:      "Total count of requests between two nodes",
	}, []string{"client", "server"})
	p.serviceGraphRequestFailedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "traces",
		Name:      "service_graph_request_failed_total",
		Help:      "Total count of failed requests between two nodes",
	}, []string{"client", "server"})
	p.serviceGraphRequestServerHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "traces",
		Name:      "service_graph_request_server_seconds",
		Help:      "Time for a request between two nodes as seen from the server",
		Buckets:   prometheus.ExponentialBuckets(0.01, 2, 12),
	}, []string{"client", "server"})
	p.serviceGraphRequestClientHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "traces",
		Name:      "service_graph_request_client_seconds",
		Help:      "Time for a request between two nodes as seen from the client",
		Buckets:   prometheus.ExponentialBuckets(0.01, 2, 12),
	}, []string{"client", "server"})
	p.serviceGraphUnpairedSpansTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "traces",
		Name:      "service_graph_unpaired_spans_total",
		Help:      "Total count of unpaired spans",
	}, []string{"client", "server"})
	p.serviceGraphDroppedSpansTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "traces",
		Name:      "service_graph_dropped_spans_total",
		Help:      "Total count of dropped spans",
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

	go func() {
		for {
			select {
			case k := <-p.firesCh:
				// Take a read lock to check if the edge has already been processed
				// Processed edges still fire their fire func
				p.storeMtx.RLock()
				if _, ok := p.store[k]; !ok {
					p.storeMtx.RUnlock()
					continue
				}
				p.storeMtx.RUnlock()

				p.storeMtx.Lock()
				e := p.store[k]
				if shouldDelete := p.collectEdge(e); shouldDelete {
					delete(p.store, k)
				}
				p.storeMtx.Unlock()

			case <-p.closedCh:
				p.collectMetrics()
				return
			}
		}
	}()

	return nil
}

func (p *processor) Shutdown(context.Context) error {
	close(p.closedCh)
	p.unregisterMetrics()
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

	return p.nextConsumer.ConsumeTraces(ctx, td)
}

// collectEdge records the metrics for the given edge.
// Returns true if the edge is completed or expired and should be deleted.
func (p *processor) collectEdge(e *edge) bool {
	if e.isCompleted() {
		p.serviceGraphRequestTotal.WithLabelValues(e.clientService, e.serverService).Inc()
		if e.failed {
			p.serviceGraphRequestFailedTotal.WithLabelValues(e.clientService, e.serverService).Inc()
		}
		p.serviceGraphRequestServerHistogram.WithLabelValues(e.clientService, e.serverService).Observe(e.serverLatency.Seconds())
		p.serviceGraphRequestClientHistogram.WithLabelValues(e.clientService, e.serverService).Observe(e.clientLatency.Seconds())
		return true
	} else if e.isExpired() {
		p.serviceGraphUnpairedSpansTotal.WithLabelValues(e.clientService, e.serverService).Inc()
		return true
	}
	return false
}

// collectMetrics loops through all the stored edges and process them.
// If an edge is completed or expired, it's recorded through the processor's metrics and deleted from the store.
func (p *processor) collectMetrics() {
	p.storeMtx.Lock()
	for k, e := range p.store {
		if shouldDelete := p.collectEdge(e); shouldDelete {
			delete(p.store, k)
		}
	}
	p.storeMtx.Unlock()
}

func (p *processor) fireFn(k string) func() {
	// todo: find a way to return early when an edge is processed before expiring
	t := time.Tick(p.wait)
	return func() {
		for {
			select {
			case <-t:
				p.firesCh <- k
			}
		}
	}
}

func (p *processor) getOrCreateEdge(k string) *edge {
	if storedEdge, ok := p.store[k]; ok {
		return storedEdge
	}

	return newEdge(p.wait, p.fireFn(k))
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

				p.storeMtx.RLock()
				if len(p.store) >= p.maxItems {
					remainingSpans := float64(ils.Spans().Len() - k)
					p.serviceGraphDroppedSpansTotal.WithLabelValues(svc.StringVal()).Add(remainingSpans)
					p.storeMtx.RUnlock()

					return errTooManyItems
				}
				p.storeMtx.RUnlock()

				span := ils.Spans().At(k)

				switch span.Kind() {
				case pdata.SpanKindClient:
					k := key(span.TraceID().HexString(), span.SpanID().HexString())

					p.storeMtx.Lock()
					edge := p.getOrCreateEdge(k)

					edge.clientService = svc.StringVal()
					edge.clientLatency = spanDuration(span)
					edge.failed = p.spanFailed(span)

					p.store[k] = edge
					p.storeMtx.Unlock()

					if edge.isCompleted() {
						p.firesCh <- k
					}

				case pdata.SpanKindServer:
					k := key(span.TraceID().HexString(), span.ParentSpanID().HexString())

					p.storeMtx.Lock()
					edge := p.getOrCreateEdge(k)

					edge.serverService = svc.StringVal()
					edge.serverLatency = spanDuration(span)
					edge.failed = p.spanFailed(span)

					p.store[k] = edge
					p.storeMtx.Unlock()

					if edge.isCompleted() {
						p.firesCh <- k
					}

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
