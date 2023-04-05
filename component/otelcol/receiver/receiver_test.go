package receiver_test

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/internal/fakeconsumer"
	"github.com/grafana/agent/component/otelcol/receiver"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
	otelconsumer "go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestReceiver(t *testing.T) {
	var (
		consumer otelconsumer.Traces

		waitConsumerTrigger = util.NewWaitTrigger()
		onTracesConsumer    = func(t otelconsumer.Traces) {
			consumer = t
			waitConsumerTrigger.Trigger()
		}

		waitTracesTrigger = util.NewWaitTrigger()
		nextConsumer      = &fakeconsumer.Consumer{
			ConsumeTracesFunc: func(context.Context, ptrace.Traces) error {
				waitTracesTrigger.Trigger()
				return nil
			},
		}
	)

	// Create and start our Flow component. We then wait for it to export a
	// consumer that we can send data to.
	te := newTestEnvironment(t, onTracesConsumer)
	te.Start(fakeReceiverArgs{
		Output: &otelcol.ConsumerArguments{
			Metrics: []otelcol.Consumer{nextConsumer},
			Logs:    []otelcol.Consumer{nextConsumer},
			Traces:  []otelcol.Consumer{nextConsumer},
		},
	})

	require.NoError(t, waitConsumerTrigger.Wait(time.Second), "no traces consumer sent")

	err := consumer.ConsumeTraces(context.Background(), ptrace.NewTraces())
	require.NoError(t, err)

	require.NoError(t, waitTracesTrigger.Wait(time.Second), "consumer did not get invoked")
}

type testEnvironment struct {
	t *testing.T

	Controller *componenttest.Controller
}

func newTestEnvironment(t *testing.T, onTracesConsumer func(t otelconsumer.Traces)) *testEnvironment {
	t.Helper()

	reg := component.Registration{
		Name:    "testcomponent",
		Args:    fakeReceiverArgs{},
		Exports: otelcol.ConsumerExports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			// Create a factory which always returns our instance of fakeReceiver
			// defined above.
			factory := otelcomponent.NewReceiverFactory(
				"testcomponent",
				func() otelconfig.Receiver { return nil },
				otelcomponent.WithTracesReceiver(func(
					ctx context.Context,
					rcs otelcomponent.ReceiverCreateSettings,
					r otelconfig.Receiver,
					t otelconsumer.Traces,
				) (otelcomponent.TracesReceiver, error) {

					onTracesConsumer(t)
					return nil, nil
				}, otelcomponent.StabilityLevelUndefined),
			)

			return receiver.New(opts, factory, args.(receiver.Arguments))
		},
	}

	return &testEnvironment{
		t:          t,
		Controller: componenttest.NewControllerFromReg(util.TestLogger(t), reg),
	}
}

func (te *testEnvironment) Start(args component.Arguments) {
	go func() {
		ctx := componenttest.TestContext(te.t)
		err := te.Controller.Run(ctx, args)
		require.NoError(te.t, err, "failed to run component")
	}()
}

type fakeReceiverArgs struct {
	Output *otelcol.ConsumerArguments
}

var _ receiver.Arguments = fakeReceiverArgs{}

func (fa fakeReceiverArgs) Convert() (otelconfig.Receiver, error) {
	settings := otelconfig.NewReceiverSettings(otelconfig.NewComponentID("testcomponent"))
	return &settings, nil
}

func (fa fakeReceiverArgs) Extensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return nil
}

func (fa fakeReceiverArgs) Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return nil
}

func (fa fakeReceiverArgs) NextConsumers() *otelcol.ConsumerArguments {
	return fa.Output
}
