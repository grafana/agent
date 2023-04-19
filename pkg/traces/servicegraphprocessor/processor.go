package servicegraphprocessor

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	util "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	otelprocessor "go.opentelemetry.io/collector/processor"
	semconv "go.opentelemetry.io/collector/semconv/v1.6.1"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"google.golang.org/grpc/codes"
)

// TODO: Do we need the component name in the metric name?
// TODO: Should the metric name be unique if there are multiple service graph processors?
// TODO: Make these const?
// TODO: Not sure what are good variable names for this?
var serviceGraphRequestTotal_name = "service_graph_request_total"
var serviceGraphRequestFailedTotal_name = "service_graph_request_failed_total"
var serviceGraphRequestServerHistogram_name = "service_graph_request_server_seconds"
var serviceGraphRequestClientHistogram_name = "service_graph_request_client_seconds"
var serviceGraphUnpairedSpansTotal_name = "service_graph_unpaired_spans_total"
var serviceGraphDroppedSpansTotal_name = "service_graph_dropped_spans_total"

type tooManySpansError struct {
	droppedSpans int
}

func (t tooManySpansError) Error() string {
	return fmt.Sprintf("dropped %d spans", t.droppedSpans)
}

// edge is an edge between two nodes in the graph
type edge struct {
	key string

	serverService, clientService string
	serverLatency, clientLatency time.Duration

	// If either the client or the server spans have status code error,
	// the edge will be considered as failed.
	failed bool

	// expiration is the time at which the edge expires, expressed as Unix time
	expiration int64
}

func newEdge(key string, ttl time.Duration) *edge {
	return &edge{
		key: key,

		expiration: time.Now().Add(ttl).Unix(),
	}
}

// isCompleted returns true if the corresponding client and server
// pair spans have been processed for the given edge
func (e *edge) isCompleted() bool {
	return len(e.clientService) != 0 && len(e.serverService) != 0
}

func (e *edge) isExpired() bool {
	return time.Now().Unix() >= e.expiration
}

var _ otelprocessor.Traces = (*processor)(nil)

type processor struct {
	nextConsumer consumer.Traces

	store *store

	wait     time.Duration
	maxItems int

	// completed edges are pushed through this channel to be processed.
	collectCh chan string

	serviceGraphRequestTotal           instrument.Float64Counter
	serviceGraphRequestFailedTotal     instrument.Float64Counter
	serviceGraphRequestServerHistogram instrument.Float64Histogram
	serviceGraphRequestClientHistogram instrument.Float64Histogram
	serviceGraphUnpairedSpansTotal     instrument.Float64Counter
	serviceGraphDroppedSpansTotal      instrument.Float64Counter

	httpSuccessCodeMap map[int]struct{}
	grpcSuccessCodeMap map[int]struct{}

	logger  log.Logger
	closeCh chan struct{}

	meterId string
}

func newProcessor(nextConsumer consumer.Traces, cfg *Config, set otelprocessor.CreateSettings) *processor {
	logger := log.With(util.Logger, "component", "service graphs")

	if cfg.Wait == 0 {
		cfg.Wait = DefaultWait
	}
	if cfg.MaxItems == 0 {
		cfg.MaxItems = DefaultMaxItems
	}
	if cfg.Workers == 0 {
		cfg.Workers = DefaultWorkers
	}

	var (
		httpSuccessCodeMap = make(map[int]struct{})
		grpcSuccessCodeMap = make(map[int]struct{})
	)
	if cfg.SuccessCodes != nil {
		for _, sc := range cfg.SuccessCodes.http {
			httpSuccessCodeMap[int(sc)] = struct{}{}
		}
		for _, sc := range cfg.SuccessCodes.grpc {
			grpcSuccessCodeMap[int(sc)] = struct{}{}
		}
	}

	p := &processor{
		nextConsumer: nextConsumer,
		logger:       logger,

		wait:               cfg.Wait,
		maxItems:           cfg.MaxItems,
		httpSuccessCodeMap: httpSuccessCodeMap,
		grpcSuccessCodeMap: grpcSuccessCodeMap,

		collectCh: make(chan string, cfg.Workers),

		closeCh: make(chan struct{}, 1),

		//TODO: Use this ot prefix the metric names?
		meterId: set.ID.String(),
	}

	for i := 0; i < cfg.Workers; i++ {
		go func() {
			for {
				select {
				case k := <-p.collectCh:
					p.store.evictEdgeWithLock(k)

				case <-p.closeCh:
					return
				}
			}
		}()
	}

	err := p.registerMetrics(set.MeterProvider)
	if err != nil {
		panic(err)
		//TODO: Should we panic?
		// level.Error(logger).Log("msg", "failed to register Otel metrics", "err", err)
		// return nil
	} else {
		//TODO: Do we want to log this?
		level.Info(logger).Log("msg", "successfully registered Otel metrics")
		//TODO: The logger doesn't include the config name?
		// ts=2023-04-17T17:15:11.772899Z caller=processor.go:171 level=info component="service graphs" msg="successfully registered Otel metrics"
	}

	return p
}

func (p *processor) Start(_ context.Context, _ component.Host) error {
	// initialize store
	p.store = newStore(p.wait, p.maxItems, p.collectEdge)

	//TODO: Check if the metrics are nil and error if they are?

	return nil
}

// TODO: This function needs to have a prefix to attach to the metric name?
func OtelMetricViews() []sdkmetric.View {
	return []sdkmetric.View{
		sdkmetric.NewView(
			sdkmetric.Instrument{Name: serviceGraphRequestServerHistogram_name},
			sdkmetric.Stream{Aggregation: aggregation.ExplicitBucketHistogram{
				//TODO: Are these buckets the same as the Prometheus ExponentialBuckets?
				Boundaries: []float64{0.01, 2, 12},
			}},
		),
		sdkmetric.NewView(
			sdkmetric.Instrument{Name: serviceGraphRequestClientHistogram_name},
			sdkmetric.Stream{Aggregation: aggregation.ExplicitBucketHistogram{
				//TODO: Are these buckets the same as the Prometheus ExponentialBuckets?
				Boundaries: []float64{0.01, 2, 12},
			}},
		),
	}
}

func (p *processor) registerMetrics(mp metric.MeterProvider) error {

	//TODO: What is a good meter name?
	meter := mp.Meter(p.meterId)

	var err error = nil
	//TODO: How to add a namespace of "traces"?
	p.serviceGraphRequestTotal, err = meter.Float64Counter(
		serviceGraphRequestTotal_name,
		instrument.WithDescription("Total count of requests between two nodes"),
	)
	if err != nil {
		return err
	}

	p.serviceGraphRequestFailedTotal, err = meter.Float64Counter(
		serviceGraphRequestFailedTotal_name,
		instrument.WithDescription("Total count of failed requests between two nodes"),
	)
	if err != nil {
		return err
	}

	p.serviceGraphRequestServerHistogram, err = meter.Float64Histogram(
		serviceGraphRequestServerHistogram_name,
		instrument.WithDescription("Time for a request between two nodes as seen from the server"),
	)
	if err != nil {
		return err
	}

	p.serviceGraphRequestClientHistogram, err = meter.Float64Histogram(
		serviceGraphRequestClientHistogram_name,
		instrument.WithDescription("Time for a request between two nodes as seen from the client"),
	)
	if err != nil {
		return err
	}

	p.serviceGraphUnpairedSpansTotal, err = meter.Float64Counter(
		serviceGraphUnpairedSpansTotal_name,
		instrument.WithDescription("Total count of unpaired spans"),
	)
	if err != nil {
		return err
	}

	p.serviceGraphDroppedSpansTotal, err = meter.Float64Counter(
		serviceGraphDroppedSpansTotal_name,
		instrument.WithDescription("Total count of dropped spans"),
	)
	if err != nil {
		return err
	}

	//TODO: Delete this later

	// p.serviceGraphRequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	// 	Namespace: "traces",
	// 	Name:      "service_graph_request_total",
	// 	Help:      "Total count of requests between two nodes",
	// }, []string{"client", "server"})
	// p.serviceGraphRequestFailedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	// 	Namespace: "traces",
	// 	Name:      "service_graph_request_failed_total",
	// 	Help:      "Total count of failed requests between two nodes",
	// }, []string{"client", "server"})
	// p.serviceGraphRequestServerHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	// 	Namespace: "traces",
	// 	Name:      "service_graph_request_server_seconds",
	// 	Help:      "Time for a request between two nodes as seen from the server",
	// 	Buckets:   prometheus.ExponentialBuckets(0.01, 2, 12),
	// }, []string{"client", "server"})
	// p.serviceGraphRequestClientHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	// 	Namespace: "traces",
	// 	Name:      "service_graph_request_client_seconds",
	// 	Help:      "Time for a request between two nodes as seen from the client",
	// 	Buckets:   prometheus.ExponentialBuckets(0.01, 2, 12),
	// }, []string{"client", "server"})
	// p.serviceGraphUnpairedSpansTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	// 	Namespace: "traces",
	// 	Name:      "service_graph_unpaired_spans_total",
	// 	Help:      "Total count of unpaired spans",
	// }, []string{"client", "server"})
	// p.serviceGraphDroppedSpansTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	// 	Namespace: "traces",
	// 	Name:      "service_graph_dropped_spans_total",
	// 	Help:      "Total count of dropped spans",
	// }, []string{"client", "server"})

	//TODO: Do we have to unregister the metrics at any point?
	return nil
}

func (p *processor) Shutdown(context.Context) error {
	close(p.closeCh)
	return nil
}

func (p *processor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

func (p *processor) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	// Evict expired edges
	p.store.expire()

	if err := p.consume(td); err != nil {
		if errors.As(err, &tooManySpansError{}) {
			level.Warn(p.logger).Log("msg", "skipped processing of spans", "maxItems", p.maxItems, "err", err)
		} else {
			level.Error(p.logger).Log("msg", "failed consuming traces", "err", err)
		}
		return nil
	}

	return p.nextConsumer.ConsumeTraces(ctx, td)
}

// collectEdge records the metrics for the given edge.
// Returns true if the edge is completed or expired and should be deleted.
func (p *processor) collectEdge(e *edge) {
	//TODO: What is a good context to use?
	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.Key("client").String(e.clientService),
		attribute.Key("server").String(e.serverService),
	}

	if e.isCompleted() {
		p.serviceGraphRequestTotal.Add(ctx, 1, attrs...)
		if e.failed {
			p.serviceGraphRequestFailedTotal.Add(ctx, 1, attrs...)
		}
		p.serviceGraphRequestServerHistogram.Record(ctx, e.serverLatency.Seconds(), attrs...)
		p.serviceGraphRequestClientHistogram.Record(ctx, e.clientLatency.Seconds(), attrs...)
	} else if e.isExpired() {
		p.serviceGraphUnpairedSpansTotal.Add(ctx, 1, attrs...)
	}
}

func (p *processor) consume(trace ptrace.Traces) error {
	var totalDroppedSpans int
	rSpansSlice := trace.ResourceSpans()
	for i := 0; i < rSpansSlice.Len(); i++ {
		rSpan := rSpansSlice.At(i)

		svc, ok := rSpan.Resource().Attributes().Get(semconv.AttributeServiceName)
		if !ok || svc.Str() == "" {
			continue
		}

		ctx := context.Background()
		//TODO: Do we really have to set the server/client to an empty string?
		attrsClient := []attribute.KeyValue{
			attribute.Key("client").String(svc.Str()),
			attribute.Key("server").String(""),
		}
		attrsServer := []attribute.KeyValue{
			attribute.Key("client").String(""),
			attribute.Key("server").String(svc.Str()),
		}

		ssSlice := rSpan.ScopeSpans()
		for j := 0; j < ssSlice.Len(); j++ {
			ils := ssSlice.At(j)

			for k := 0; k < ils.Spans().Len(); k++ {
				span := ils.Spans().At(k)

				switch span.Kind() {
				case ptrace.SpanKindClient:
					k := key(hex.EncodeToString([]byte(span.TraceID().String())), hex.EncodeToString([]byte(span.SpanID().String())))

					edge, err := p.store.upsertEdge(k, func(e *edge) {
						e.clientService = svc.Str()
						e.clientLatency = spanDuration(span)
						e.failed = e.failed || p.spanFailed(span) // keep request as failed if any span is failed
					})

					if errors.Is(err, errTooManyItems) {
						totalDroppedSpans++
						p.serviceGraphDroppedSpansTotal.Add(ctx, 1, attrsClient...)
						continue
					}
					// upsertEdge will only return this errTooManyItems
					if err != nil {
						return err
					}

					if edge.isCompleted() {
						p.collectCh <- k
					}

				case ptrace.SpanKindServer:
					k := key(hex.EncodeToString([]byte(span.TraceID().String())), hex.EncodeToString([]byte(span.ParentSpanID().String())))

					edge, err := p.store.upsertEdge(k, func(e *edge) {
						e.serverService = svc.Str()
						e.serverLatency = spanDuration(span)
						e.failed = e.failed || p.spanFailed(span) // keep request as failed if any span is failed
					})

					if errors.Is(err, errTooManyItems) {
						totalDroppedSpans++
						p.serviceGraphDroppedSpansTotal.Add(ctx, 1, attrsServer...)
						continue
					}
					// upsertEdge will only return this errTooManyItems
					if err != nil {
						return err
					}

					if edge.isCompleted() {
						p.collectCh <- k
					}

				default:
				}
			}
		}
	}

	if totalDroppedSpans > 0 {
		return &tooManySpansError{
			droppedSpans: totalDroppedSpans,
		}
	}
	return nil
}

func (p *processor) spanFailed(span ptrace.Span) bool {
	// Request considered failed if status is not 2XX or added as a successful status code
	if statusCode, ok := span.Attributes().Get(semconv.AttributeHTTPStatusCode); ok {
		sc := int(statusCode.Int())
		if _, ok := p.httpSuccessCodeMap[sc]; !ok && sc/100 != 2 {
			return true
		}
	}

	// Request considered failed if status is not OK or added as a successful status code
	if statusCode, ok := span.Attributes().Get(semconv.AttributeRPCGRPCStatusCode); ok {
		sc := int(statusCode.Int())
		if _, ok := p.grpcSuccessCodeMap[sc]; !ok && sc != int(codes.OK) {
			return true
		}
	}

	return span.Status().Code() == ptrace.StatusCodeError
}

func spanDuration(span ptrace.Span) time.Duration {
	return span.EndTimestamp().AsTime().Sub(span.StartTimestamp().AsTime())
}

func key(k1, k2 string) string {
	return fmt.Sprintf("%s-%s", k1, k2)
}
