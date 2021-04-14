package automaticloggingprocessor

import (
	"context"

	util "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/kit/log"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
)

type promServiceDiscoProcessor struct {
	nextConsumer consumer.TracesConsumer

	logger log.Logger
}

func newTraceProcessor(nextConsumer consumer.TracesConsumer) (component.TracesProcessor, error) {
	logger := log.With(util.Logger, "component", "tempo automatic logging")

	if nextConsumer == nil {
		return nil, componenterror.ErrNilNextConsumer
	}
	return &promServiceDiscoProcessor{
		nextConsumer: nextConsumer,
		logger:       logger,
	}, nil
}

func (p *promServiceDiscoProcessor) ConsumeTraces(ctx context.Context, td pdata.Traces) error {
	return p.nextConsumer.ConsumeTraces(ctx, td)
}

func (p *promServiceDiscoProcessor) GetCapabilities() component.ProcessorCapabilities {
	return component.ProcessorCapabilities{}
}

// Start is invoked during service startup.
func (p *promServiceDiscoProcessor) Start(_ context.Context, _ component.Host) error {
	return nil
}

// Shutdown is invoked during service shutdown.
func (p *promServiceDiscoProcessor) Shutdown(context.Context) error {
	return nil
}
