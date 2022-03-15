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
	"go.opentelemetry.io/collector/config/configunmarshaler"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/service/external/builder"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Server is a Tracing testing server that invokes a function every time a span
// is received.
type Server struct {
	receivers builder.Receivers
	pipelines builder.BuiltPipelines
	exporters builder.Exporters
}

// NewTestServer creates a new Server for testing, where received traces will
// call the callback function. The returned string is the address where traces
// can be sent using OTLP.
func NewTestServer(t *testing.T, callback func(pdata.Traces)) string {
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
func NewServerWithRandomPort(callback func(pdata.Traces)) (srv *Server, addr string, err error) {
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
func NewServer(addr string, callback func(pdata.Traces)) (*Server, error) {
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

	configMap := config.NewMapFromStringMap(cfg)
	cfgUnmarshaler := configunmarshaler.NewDefault()
	otelCfg, err := cfgUnmarshaler.Unmarshal(configMap, factories)
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

	exporters, err := builder.BuildExporters(settings, startInfo, otelCfg, factories.Exporters)
	if err != nil {
		return nil, fmt.Errorf("failed to build exporters: %w", err)
	}
	if err := exporters.StartAll(context.Background(), nil); err != nil {
		return nil, fmt.Errorf("failed to start exporters: %w", err)
	}

	pipelines, err := builder.BuildPipelines(settings, startInfo, otelCfg, exporters, factories.Processors)
	if err != nil {
		return nil, fmt.Errorf("failed to build pipelines: %w", err)
	}
	if err := pipelines.StartProcessors(context.Background(), nil); err != nil {
		return nil, fmt.Errorf("failed to start pipelines: %w", err)
	}

	receivers, err := builder.BuildReceivers(settings, startInfo, otelCfg, pipelines, factories.Receivers)
	if err != nil {
		return nil, fmt.Errorf("failed to build receivers: %w", err)
	}
	h := &mocks.Host{}
	h.On("GetExtensions").Return(nil)
	if err := receivers.StartAll(context.Background(), h); err != nil {
		return nil, fmt.Errorf("failed to start receivers: %w", err)
	}

	return &Server{
		receivers: receivers,
		pipelines: pipelines,
		exporters: exporters,
	}, nil
}

// Stop stops the testing server.
func (s *Server) Stop() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var firstErr error

	deps := []func(context.Context) error{
		s.receivers.ShutdownAll,
		s.pipelines.ShutdownProcessors,
		s.exporters.ShutdownAll,
	}
	for _, dep := range deps {
		err := dep(shutdownCtx)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func newFuncProcessorFactory(callback func(pdata.Traces)) component.ProcessorFactory {
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
		}),
	)
}

type funcProcessor struct {
	Callback func(pdata.Traces)
	Next     consumer.Traces
}

func (p *funcProcessor) ConsumeTraces(ctx context.Context, td pdata.Traces) error {
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
		}),
	)
}

type noopExporter struct{}

func (n noopExporter) Start(context.Context, component.Host) error { return nil }

func (n noopExporter) Shutdown(context.Context) error { return nil }

func (n noopExporter) Capabilities() consumer.Capabilities { return consumer.Capabilities{} }

func (n noopExporter) ConsumeTraces(context.Context, pdata.Traces) error { return nil }
