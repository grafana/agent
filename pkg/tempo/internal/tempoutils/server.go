package tempoutils

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/service"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Server is a Tempo testing server that invokes a function every time a span
// is received.
type Server struct {
	app *service.Application
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
	factories, err := tracingFactories(callback)
	if err != nil {
		return nil, fmt.Errorf("failed creating tracing factories: %s", err)
	}

	var (
		buildInfo component.BuildInfo
	)

	cp := newParserProvider(addr)

	params := service.Parameters{
		Factories:      factories,
		BuildInfo:      buildInfo,
		ParserProvider: cp,
		LoggingOptions: []zap.Option{zap.Development()},
	}
	app, err := service.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed creating tracing application: %s", err)
	}

	srv := &Server{
		app: app,
	}
	if err := srv.initTracingApp(20 * time.Second); err != nil {
		return nil, err
	}

	return srv, nil
}

func otelConfig(addr string) map[string]interface{} {
	conf := util.Untab(fmt.Sprintf(`
processors:
	func_processor:
receivers:
  otlp:
		protocols:
			grpc:
				endpoint: %s
exporters:
  otlp:
    endpoint: example.com:12345
service:
	pipelines:
		traces:
			receivers: [otlp]
			processors: [func_processor]
			exporters: [otlp]
	`, addr))

	var cfg map[string]interface{}
	if err := yaml.NewDecoder(strings.NewReader(conf)).Decode(&cfg); err != nil {
		panic("could not decode config: " + err.Error())
	}

	return cfg
}

func (s *Server) initTracingApp(timeout time.Duration) error {
	initCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	go func() {
		s.app.Command().SetArgs([]string{"--metrics-level=none"})
		err := s.app.Run()
		if err != nil {
			cancel()
		}
	}()

	for {
		select {
		case s := <-s.app.GetStateChannel():
			if s == service.Running {
				return nil
			}
		case err := <-initCtx.Done():
			return fmt.Errorf("failed to start tracing application: %s", err)
		}
	}
}

// Stop stops the testing server.
func (s *Server) Stop() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	s.app.Shutdown()
	for {
		select {
		case state := <-s.app.GetStateChannel():
			switch state {
			case service.Closed:
				return nil
			}
		case err := <-shutdownCtx.Done():
			return fmt.Errorf("failed to stop tracing application: %s", err)
		}
	}
}

func tracingFactories(callback func(traces pdata.Traces)) (component.Factories, error) {
	extensionsFactory, err := component.MakeExtensionFactoryMap()
	if err != nil {
		return component.Factories{}, fmt.Errorf("failed to make extension factory map: %w", err)
	}

	receiversFactory, err := component.MakeReceiverFactoryMap(otlpreceiver.NewFactory())
	if err != nil {
		return component.Factories{}, fmt.Errorf("failed to make receiver factory map: %w", err)
	}

	exportersFactory, err := component.MakeExporterFactoryMap(otlpexporter.NewFactory())
	if err != nil {
		return component.Factories{}, fmt.Errorf("failed to make exporter factory map: %w", err)
	}

	processorsFactory, err := component.MakeProcessorFactoryMap(
		newFuncProcessorFactory(callback),
	)
	if err != nil {
		return component.Factories{}, fmt.Errorf("failed to make processor factory map: %w", err)
	}

	return component.Factories{
		Extensions: extensionsFactory,
		Receivers:  receiversFactory,
		Processors: processorsFactory,
		Exporters:  exportersFactory,
	}, nil
}

func newFuncProcessorFactory(callback func(pdata.Traces)) component.ProcessorFactory {
	return processorhelper.NewFactory(
		"func_processor",
		func() config.Processor {
			processorSettings := config.NewProcessorSettings(config.NewIDWithName("func_processor", "func_processor"))
			return &processorSettings
		},
		processorhelper.WithTraces(func(
			_ context.Context,
			_ component.ProcessorCreateParams,
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

func newParserProvider(addr string) *parserProvider {
	return &parserProvider{
		addr: addr,
	}
}

type parserProvider struct {
	addr string
}

func (p *parserProvider) Get() (*config.Parser, error) {
	otelConfig := otelConfig(p.addr)
	return config.NewParserFromStringMap(otelConfig), nil
}
