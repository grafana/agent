package exporter_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/exporter"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
	otelconsumer "go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestExporter(t *testing.T) {
	ctx := componenttest.TestContext(t)

	// Channel where received traces will be written to.
	tracesCh := make(chan ptrace.Traces, 1)

	// Create an instance of a fake OpenTelemetry Collector exporter which our
	// Flow component will wrap around.
	innerExporter := &fakeExporter{
		ConsumeTracesFunc: func(_ context.Context, td ptrace.Traces) error {
			select {
			case tracesCh <- td:
			default:
			}
			return nil
		},
	}

	// Create and start our Flow component. We then wait for it to export a
	// consumer that we can send data to.
	te := newTestEnvironment(t, innerExporter)
	te.Start()

	require.NoError(t, te.Controller.WaitExports(1*time.Second), "test component did not generate exports")
	ce := te.Controller.Exports().(otelcol.ConsumerExports)

	// Create a test set of traces and send it to our consumer in the background.
	// We then wait for our channel to receive the traces, indicating that
	// everything was wired up correctly.
	testTraces := createTestTraces()
	go func() {
		var err error

		for {
			err = ce.Input.ConsumeTraces(ctx, testTraces)

			if errors.Is(err, otelcomponent.ErrDataTypeIsNotSupported) {
				// Our component may not have been fully initialized yet. Wait a little
				// bit before trying again.
				time.Sleep(100 * time.Millisecond)
				continue
			}

			require.NoError(t, err)
			break
		}
	}()

	select {
	case <-time.After(1 * time.Second):
		require.FailNow(t, "testcomponent did not receive traces")
	case td := <-tracesCh:
		require.Equal(t, testTraces, td)
	}
}

type testEnvironment struct {
	t *testing.T

	Controller *componenttest.Controller
}

func newTestEnvironment(t *testing.T, fe *fakeExporter) *testEnvironment {
	t.Helper()

	reg := component.Registration{
		Name:    "testcomponent",
		Args:    fakeExporterArgs{},
		Exports: otelcol.ConsumerExports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			// Create a factory which always returns our instance of fakeExporter
			// defined above.
			factory := otelcomponent.NewExporterFactory(
				"testcomponent",
				func() otelconfig.Exporter {
					res, err := fakeExporterArgs{}.Convert()
					require.NoError(t, err)
					return res
				},
				otelcomponent.WithTracesExporter(func(ctx context.Context, ecs otelcomponent.ExporterCreateSettings, e otelconfig.Exporter) (otelcomponent.TracesExporter, error) {
					return fe, nil
				}, otelcomponent.StabilityLevelUndefined),
			)

			return exporter.New(opts, factory, args.(exporter.Arguments))
		},
	}

	return &testEnvironment{
		t:          t,
		Controller: componenttest.NewControllerFromReg(util.TestLogger(t), reg),
	}
}

func (te *testEnvironment) Start() {
	go func() {
		ctx := componenttest.TestContext(te.t)
		err := te.Controller.Run(ctx, fakeExporterArgs{})
		require.NoError(te.t, err, "failed to run component")
	}()
}

type fakeExporterArgs struct {
}

var _ exporter.Arguments = fakeExporterArgs{}

func (fa fakeExporterArgs) Convert() (otelconfig.Exporter, error) {
	settings := otelconfig.NewExporterSettings(otelconfig.NewComponentID("testcomponent"))
	return &settings, nil
}

func (fa fakeExporterArgs) Extensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return nil
}

func (fa fakeExporterArgs) Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return nil
}

type fakeExporter struct {
	StartFunc         func(ctx context.Context, host otelcomponent.Host) error
	ShutdownFunc      func(ctx context.Context) error
	CapabilitiesFunc  func() otelconsumer.Capabilities
	ConsumeTracesFunc func(ctx context.Context, td ptrace.Traces) error
}

var _ otelcomponent.TracesExporter = (*fakeExporter)(nil)

func (fe *fakeExporter) Start(ctx context.Context, host otelcomponent.Host) error {
	if fe.StartFunc != nil {
		return fe.StartFunc(ctx, host)
	}
	return nil
}

func (fe *fakeExporter) Shutdown(ctx context.Context) error {
	if fe.ShutdownFunc != nil {
		return fe.ShutdownFunc(ctx)
	}
	return nil
}

func (fe *fakeExporter) Capabilities() otelconsumer.Capabilities {
	if fe.CapabilitiesFunc != nil {
		return fe.CapabilitiesFunc()
	}
	return otelconsumer.Capabilities{}
}

func (fe *fakeExporter) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	if fe.ConsumeTracesFunc != nil {
		return fe.ConsumeTracesFunc(ctx, td)
	}
	return nil
}

func createTestTraces() ptrace.Traces {
	// Matches format from the protobuf definition:
	// https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/trace/v1/trace.proto
	var bb = `{
		"resource_spans": [{
			"scope_spans": [{
				"spans": [{
					"name": "TestSpan"
				}]
			}]
		}]
	}`

	decoder := &ptrace.JSONUnmarshaler{}
	data, err := decoder.UnmarshalTraces([]byte(bb))
	if err != nil {
		panic(err)
	}
	return data
}
