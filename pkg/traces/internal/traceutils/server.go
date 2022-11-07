package traceutils

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/util"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanmetricsprocessor/mocks"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/service/external/configunmarshaler"
	"go.opentelemetry.io/collector/service/external/pipelines"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Server is a Tracing testing server that invokes a function every time a span
// is received.
type Server struct {
	pipelines *pipelines.Pipelines
}

// NewTestServer creates a new Server for testing, where received traces will
// call the callback function. The returned string is the address where traces
// can be sent using OTLP.
func NewTestServer(t *testing.T, callback func(ptrace.Traces)) string {
	t.Helper()

	srv, listenAddr, err := NewServerWithRandomPort(callback)
	if err != nil {
		t.Fatalf("failed to create OTLP server: %s", err)
	}
	t.Cleanup(func() {
		err := srv.Stop()
		assert.NoError(t, err)
	})

	return listenAddr
}

// NewServerWithRandomPort calls NewServer with a random port >49152 and
// <65535. It will try up to five times before failing.
func NewServerWithRandomPort(callback func(ptrace.Traces)) (srv *Server, addr string, err error) {
	var lastError error

	for i := 0; i < 5; i++ {
		port := rand.Intn(65535-49152) + 49152
		listenAddr := fmt.Sprintf("127.0.0.1:%d", port)

		srv, err = NewServer(listenAddr, callback)
		if err != nil {
			lastError = err
			continue
		}

		return srv, listenAddr, nil
	}

	return nil, "", fmt.Errorf("failed 5 times to create a server. last error: %w", lastError)
}

// NewServer creates an OTLP-accepting server that calls a function when a
// trace is received. This is primarily useful for testing.
func NewServer(addr string, callback func(ptrace.Traces)) (*Server, error) {
	conf := util.Untab(fmt.Sprintf(`
processors:
	func_processor:
receivers:
  otlp:
		protocols:
			grpc:
				endpoint: %s
exporters:
  noop:
service:
	pipelines:
		traces:
			receivers: [otlp]
			processors: [func_processor]
			exporters: [noop]
	`, addr))

	var cfg map[string]interface{}
	if err := yaml.NewDecoder(strings.NewReader(conf)).Decode(&cfg); err != nil {
		panic("could not decode config: " + err.Error())
	}

	extensionsFactory, err := component.MakeExtensionFactoryMap()
	if err != nil {
		return nil, fmt.Errorf("failed to make extension factory map: %w", err)
	}

	receiversFactory, err := component.MakeReceiverFactoryMap(otlpreceiver.NewFactory())
	if err != nil {
		return nil, fmt.Errorf("failed to make receiver factory map: %w", err)
	}

	exportersFactory, err := component.MakeExporterFactoryMap(newNoopExporterFactory())
	if err != nil {
		return nil, fmt.Errorf("failed to make exporter factory map: %w", err)
	}

	processorsFactory, err := component.MakeProcessorFactoryMap(
		newFuncProcessorFactory(callback),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make processor factory map: %w", err)
	}

	factories := component.Factories{
		Extensions: extensionsFactory,
		Receivers:  receiversFactory,
		Processors: processorsFactory,
		Exporters:  exportersFactory,
	}

	configMap := confmap.NewFromStringMap(cfg)
	otelCfg, err := configunmarshaler.Unmarshal(configMap, factories)
	if err != nil {
		return nil, fmt.Errorf("failed to make otel config: %w", err)
	}

	var (
		logger    = zap.NewNop()
		startInfo component.BuildInfo
	)

	settings := component.TelemetrySettings{
		Logger:         logger,
		TracerProvider: trace.NewNoopTracerProvider(),
		MeterProvider:  metric.NewNoopMeterProvider(),
	}

	pipelines, err := pipelines.Build(context.Background(), pipelines.Settings{
		Telemetry: settings,
		BuildInfo: startInfo,

		ReceiverFactories:  factories.Receivers,
		ReceiverConfigs:    otelCfg.Receivers,
		ProcessorFactories: factories.Processors,
		ProcessorConfigs:   otelCfg.Processors,
		ExporterFactories:  factories.Exporters,
		ExporterConfigs:    otelCfg.Exporters,

		PipelineConfigs: otelCfg.Pipelines,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build pipelines: %w", err)
	}

	h := &mocks.Host{}
	h.On("GetExtensions").Return(nil)
	if err := pipelines.StartAll(context.Background(), h); err != nil {
		return nil, fmt.Errorf("failed to start receivers: %w", err)
	}

	return &Server{
		pipelines: pipelines,
	}, nil
}

// Stop stops the testing server.
func (s *Server) Stop() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.pipelines.ShutdownAll(shutdownCtx)
}

func newFuncProcessorFactory(callback func(ptrace.Traces)) component.ProcessorFactory {
	return component.NewProcessorFactory(
		"func_processor",
		func() config.Processor {
			processorSettings := config.NewProcessorSettings(config.NewComponentIDWithName("func_processor", "func_processor"))
			return &processorSettings
		},
		component.WithTracesProcessor(func(
			_ context.Context,
			_ component.ProcessorCreateSettings,
			_ config.Processor,
			next consumer.Traces,
		) (component.TracesProcessor, error) {

			return &funcProcessor{
				Callback: callback,
				Next:     next,
			}, nil
		}, component.StabilityLevelUndefined),
	)
}

type funcProcessor struct {
	Callback func(ptrace.Traces)
	Next     consumer.Traces
}

func (p *funcProcessor) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	if p.Callback != nil {
		p.Callback(td)
	}
	return p.Next.ConsumeTraces(ctx, td)
}

func (p *funcProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (p *funcProcessor) Start(context.Context, component.Host) error { return nil }
func (p *funcProcessor) Shutdown(context.Context) error              { return nil }

func newNoopExporterFactory() component.ExporterFactory {
	return component.NewExporterFactory(
		"noop",
		func() config.Exporter {
			exporterSettings := config.NewExporterSettings(config.NewComponentIDWithName("noop", "noop"))
			return &exporterSettings
		},
		component.WithTracesExporter(func(
			context.Context,
			component.ExporterCreateSettings,
			config.Exporter) (
			component.TracesExporter,
			error) {

			return &noopExporter{}, nil
		}, component.StabilityLevelUndefined),
	)
}

type noopExporter struct{}

func (n noopExporter) Start(context.Context, component.Host) error { return nil }

func (n noopExporter) Shutdown(context.Context) error { return nil }

func (n noopExporter) Capabilities() consumer.Capabilities { return consumer.Capabilities{} }

func (n noopExporter) ConsumeTraces(context.Context, ptrace.Traces) error { return nil }
